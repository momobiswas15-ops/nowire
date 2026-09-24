package review

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nowire/nowire/internal/diff"
	"github.com/nowire/nowire/internal/ollama"
	"sort"
	"strings"
)

type Finding struct {
	Severity   string `json:"severity"`
	File       string `json:"file"`
	Line       int    `json:"line"`
	Title      string `json:"title"`
	Message    string `json:"message"`
	Suggestion string `json:"suggestion,omitempty"`
}

func Run(ctx context.Context, c *ollama.Client, model, policy string, hunks []diff.Hunk) ([]Finding, error) {
	var all []Finding
	for _, h := range hunks {
		raw, err := c.Generate(ctx, model, ollama.Prompt(policy, h.File, h.Added))
		if err != nil {
			return nil, err
		}
		var fs []Finding
		if err := json.Unmarshal([]byte(strings.TrimSpace(raw)), &fs); err != nil {
			return nil, fmt.Errorf("invalid model JSON for %s: %w", h.File, err)
		}
		for i := range fs {
			if fs[i].File == "" {
				fs[i].File = h.File
			}
			if fs[i].Line == 0 {
				fs[i].Line = h.Start
			}
		}
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
