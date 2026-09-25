package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/nowire/nowire/internal/config"
	"github.com/nowire/nowire/internal/diff"
	"github.com/nowire/nowire/internal/ollama"
	"github.com/nowire/nowire/internal/output"
	"github.com/nowire/nowire/internal/review"
	"os"
	"strings"
	"time"
)

const version = "0.3.0"

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}
	switch os.Args[1] {
	case "init":
		initCmd()
	case "review":
		reviewCmd(os.Args[2:])
	case "models":
		modelsCmd()
	case "doctor":
		doctorCmd()
	case "version":
		fmt.Println("nowire", version)
	default:
		usage()
		os.Exit(2)
	}
}
func usage() {
	fmt.Println(`nowire — local AI code review via Ollama

Commands:
  nowire init                         create .nowire.yml and policy
  nowire review [path] [flags]        review changed code
  nowire models                       show recommended local models
  nowire doctor                       check Git, Ollama, and model readiness
  nowire version                      print version

Review flags: --staged, --base REF, --workers N, --timeout DURATION, --model,
			--format terminal|markdown|github|json|sarif, --output FILE,
              --fail-on high|medium|low, --ollama URL`)
}
func initCmd() {
	c := config.Default()
	if err := c.Save(""); err != nil {
		fail(err)
	}
	if _, err := os.Stat(".nowire-policy.md"); os.IsNotExist(err) {
		os.WriteFile(".nowire-policy.md", []byte("# Repository review policy\n\n- Prefer small, testable functions.\n- Never log secrets or personal data.\n- Validate external input at system boundaries.\n"), 0644)
	}
	fmt.Println("created .nowire.yml and .nowire-policy.md")
}
func modelsCmd() {
	fmt.Println("Recommended models (choose based on RAM):\n  8 GB   qwen2.5-coder:1.5b\n  16 GB  qwen2.5-coder:7b\n  32 GB  qwen2.5-coder:14b\n\nPull one with: ollama pull qwen2.5-coder:7b")
}
func doctorCmd() {
	c := config.Load("")
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	if _, err := os.Stat(".git"); err != nil {
		fmt.Println("✕ git repository not found")
		os.Exit(2)
	} else {
		fmt.Println("✓ git repository")
	}
	models, err := ollama.New(c.OllamaURL).Tags(ctx)
	if err != nil {
		fmt.Printf("✕ Ollama %s (%v)\n", c.OllamaURL, err)
		os.Exit(2)
	}
	fmt.Printf("✓ Ollama %s\n", c.OllamaURL)
	if ollama.HasModel(models, c.Model) {
		fmt.Printf("✓ model %s\n", c.Model)
	} else {
		fmt.Printf("! model %s not found — run: ollama pull %s\n", c.Model, c.Model)
		os.Exit(1)
	}
}
func reviewCmd(args []string) {
	fs := flag.NewFlagSet("review", flag.ExitOnError)
	staged := fs.Bool("staged", false, "review staged changes")
	base := fs.String("base", "", "review the diff from REF to HEAD (useful in CI)")
	workers := fs.Int("workers", 1, "maximum concurrent Ollama reviews")
	timeout := fs.Duration("timeout", 2*time.Minute, "maximum review duration")
	model := fs.String("model", "", "Ollama model")
	format := fs.String("format", "", "output format")
	out := fs.String("output", "", "write output to file")
	failOn := fs.String("fail-on", "", "fail threshold")
	url := fs.String("ollama", "", "Ollama URL")
	fs.Parse(args)
	path := ""
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}
	c := config.Load("")
	if *model != "" {
		c.Model = *model
	}
	if *format != "" {
		c.Output = *format
	}
	if *failOn != "" {
		c.FailOn = *failOn
	}
	if *url != "" {
		c.OllamaURL = *url
	}
	h, err := diff.Collect(path, *staged, *base)
	if err != nil {
		fail(err)
	}
	if len(h) == 0 {
		fmt.Print("✓ No changed lines to review.\n")
		return
	}
	policy := ""
	if c.Policy != "" {
		if b, e := os.ReadFile(c.Policy); e == nil {
			policy = string(b)
		}
	}
	if *timeout <= 0 {
		fail(fmt.Errorf("timeout must be positive"))
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	findings, err := review.Run(ctx, ollama.New(c.OllamaURL), c.Model, policy, h, *workers)
	if err != nil {
		fail(err)
	}
	var data []byte
	var text string
	switch strings.ToLower(c.Output) {
	case "json":
		data, err = jsonBytes(findings)
	case "sarif":
		data, err = output.SARIF(findings)
	case "markdown", "md":
		text = output.Markdown(findings)
	case "github", "annotations":
		text = output.GitHubAnnotations(findings)
	default:
		text = output.Terminal(findings)
	}
	if err != nil {
		fail(err)
	}
	if len(data) > 0 {
		text = string(data) + "\n"
	}
	if *out != "" {
		if err = os.WriteFile(*out, []byte(text), 0644); err != nil {
			fail(err)
		}
	} else {
		fmt.Print(text)
	}
	if review.Fails(findings, c.FailOn) {
		os.Exit(1)
	}
}
func jsonBytes(v any) ([]byte, error) { return json.MarshalIndent(v, "", "  ") }
func fail(err error)                  { fmt.Fprintln(os.Stderr, "nowire:", err); os.Exit(2) }
