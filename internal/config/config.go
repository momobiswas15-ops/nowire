package config

import (
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Model     string
	OllamaURL string
	FailOn    string
	Output    string
	Policy    string
	Template  string
	Ignore    []string
}

func Default() Config {
	return Config{Model: "qwen2.5-coder:7b", OllamaURL: "http://127.0.0.1:11434", FailOn: "high", Output: "terminal", Ignore: []string{"vendor/", "node_modules/", ".git/", "dist/"}}
}

func Load(path string) Config {
	c := Default()
	if v := os.Getenv("NOWIRE_MODEL"); v != "" {
		c.Model = v
	}
	if v := os.Getenv("OLLAMA_HOST"); v != "" {
		c.OllamaURL = strings.TrimRight(v, "/")
	}
	if path == "" {
		path = ".nowire.yml"
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return c
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		p := strings.SplitN(line, ":", 2)
		if len(p) != 2 {
			continue
		}
		k, v := strings.TrimSpace(p[0]), strings.Trim(strings.TrimSpace(p[1]), "\"'")
		switch k {
		case "model":
			c.Model = v
		case "ollama_url":
			c.OllamaURL = v
		case "fail_on":
			c.FailOn = v
		case "output":
			c.Output = v
		case "policy":
			c.Policy = v
		case "template":
			c.Template = v
		}
	}
	if c.Policy != "" && !filepath.IsAbs(c.Policy) {
		c.Policy = filepath.Clean(c.Policy)
	}
	if c.Template != "" && !filepath.IsAbs(c.Template) {
		c.Template = filepath.Clean(c.Template)
	}
	return c
}

func (c Config) Save(path string) error {
	if path == "" {
		path = ".nowire.yml"
	}
	return os.WriteFile(path, []byte("# nowire configuration\nmodel: "+c.Model+"\nollama_url: "+c.OllamaURL+"\nfail_on: "+c.FailOn+"\noutput: "+c.Output+"\npolicy: .nowire-policy.md\n# template: .nowire-roast.md\n"), 0644)
}
