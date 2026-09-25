package review

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/nowire/nowire/internal/diff"
	"github.com/nowire/nowire/internal/ollama"
)

type Finding struct {
	Severity   string `json:"severity"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

func Run(ctx context.Context, c *ollama.Client, model, policy, template string, hunks []diff.Hunk, workers int) ([]Finding, error) {
	if len(hunks) == 0 {
		return nil, nil
	}
	if workers < 1 {
		workers = 1
	}
	if workers > len(hunks) {
		workers = len(hunks)
	}
	jobs := make(chan int)
	results := make([][]Finding, len(hunks))
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex
	worker := func() {
		defer wg.Done()
		for i := range jobs {
			h := hunks[i]
			raw, err := c.Generate(ctx, model, ollama.RenderTemplate(template, policy, h.File, h.Added))
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errMu.Unlock()
				continue
			}
			fs, err := parseFindings(raw)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("invalid model JSON for %s: %w", h.File, err)
				}
				errMu.Unlock()
				continue
			}
			for j := range fs {
				if fs[j].File == "" {
					fs[j].File = h.File
				}
				if fs[j].Line == 0 {
					fs[j].Line = h.Start
				}
			}
			results[i] = fs
		}
	}
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go worker()
	}
	for i := range hunks {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	var all []Finding
	for _, fs := range results {
		all = append(all, fs...)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].File != all[j].File {
			return all[i].File < all[j].File
		}
		if all[i].Line != all[j].Line {
			return all[i].Line < all[j].Line
		}
		return rank(all[i].Severity) > rank(all[j].Severity)
	})
	return dedupe(all), nil
}

func parseFindings(raw string) ([]Finding, error) {
	clean := strings.TrimSpace(raw)
	clean = strings.TrimPrefix(clean, "```json")
	clean = strings.TrimPrefix(clean, "```")
	clean = strings.TrimSuffix(clean, "```")
	clean = strings.TrimSpace(clean)
	var fs []Finding
	if err := json.Unmarshal([]byte(clean), &fs); err == nil {
		return fs, nil
	}
	var envelope struct {
		Findings []Finding `json:"findings"`
	}
	if err := json.Unmarshal([]byte(clean), &envelope); err == nil && envelope.Findings != nil {
		return envelope.Findings, nil
	}
	return nil, fmt.Errorf("expected a JSON array of findings")
}
func rank(s string) int {
	switch strings.ToLower(s) {
	case "high":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}
func dedupe(in []Finding) []Finding {
	seen := map[string]bool{}
	out := []Finding{}
	for _, f := range in {
		k := fmt.Sprintf("%s:%d:%s", f.File, f.Line, f.Title)
		if !seen[k] {
			seen[k] = true
			out = append(out, f)
		}
	}
	return out
}
func Fails(fs []Finding, threshold string) bool {
	t := rank(threshold)
	for _, f := range fs {
		if rank(f.Severity) >= t {
			return true
		}
	}
	return false
}
