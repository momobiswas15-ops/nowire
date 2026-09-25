package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	BaseURL string
	HTTP    *http.Client
}
type request struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
	Format string `json:"format,omitempty"`
}
type response struct {
	Response string `json:"response"`
	Error    string `json:"error"`
}

func New(url string) *Client {
	return &Client{strings.TrimRight(url, "/"), &http.Client{Timeout: 120 * time.Second}}
}

func (c *Client) Generate(ctx context.Context, model, prompt string) (string, error) {
	body, err := json.Marshal(request{Model: model, Prompt: prompt, Stream: false, Format: "json"})
	if err != nil {
		return "", err
	}
	var last error
	for attempt := 0; attempt < 3; attempt++ {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/generate", bytes.NewReader(body))
		if err != nil {
			return "", err
		}
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			last = fmt.Errorf("ollama unavailable: %w", err)
		} else {
			var out response
			decodeErr := json.NewDecoder(res.Body).Decode(&out)
			res.Body.Close()
			if decodeErr != nil {
				last = decodeErr
			} else if res.StatusCode >= 300 || out.Error != "" {
				if out.Error == "" {
					out.Error = res.Status
				}
				last = fmt.Errorf("ollama: %s", out.Error)
				if res.StatusCode < 429 && res.StatusCode < 500 {
					return "", last
				}
			} else {
				return out.Response, nil
			}
		}
		if attempt < 2 {
			delay := time.Duration(1<<attempt) * 250 * time.Millisecond
			select {
			case <-ctx.Done():
				return "", ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return "", last
}

const DefaultTemplate = `You are a strict but constructive senior code reviewer. Review only the supplied changed context. Return JSON array only, no markdown, with objects matching: {"severity":"high|medium|low","file":"string","line":number,"title":"short title","message":"specific actionable explanation","suggestion":"minimal fix"}. Ignore formatting nits. Report real bugs, security issues, data loss, broken error handling, and maintainability risks. Do not invent APIs or complain about code outside the context.
Repository policy:
{{policy}}
File: {{file}}
Changed context:
{{code}}`

func Prompt(policy, file, code string) string {
	return RenderTemplate(DefaultTemplate, policy, file, code)
}

func RenderTemplate(template, policy, file, code string) string {
	if template == "" {
		template = DefaultTemplate
	}
	return strings.NewReplacer("{{policy}}", policy, "{{file}}", file, "{{code}}", code).Replace(template)
}
