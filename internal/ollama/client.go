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
	body, _ := json.Marshal(request{Model: model, Prompt: prompt, Stream: false, Format: "json"})
	req, err := http.NewRequestWithContext(ctx, "POST", c.BaseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	r, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama unavailable: %w", err)
	}
	defer r.Body.Close()
	var out response
	if err = json.NewDecoder(r.Body).Decode(&out); err != nil {
		return "", err
	}
	if r.StatusCode >= 300 || out.Error != "" {
		if out.Error == "" {
			out.Error = r.Status
		}
		return "", fmt.Errorf("ollama: %s", out.Error)
	}
	return out.Response, nil
}
func Prompt(policy, file, code string) string {
	return fmt.Sprintf(`You are a strict but constructive senior code reviewer. Review only the supplied changed context. Return JSON array only, no markdown, with objects matching: {"severity":"high|medium|low","file":"string","line":number,"title":"short title","message":"specific actionable explanation","suggestion":"minimal fix"}. Ignore formatting nits. Report real bugs, security issues, data loss, broken error handling, and maintainability risks. Do not invent APIs or complain about code outside the context.
Repository policy:
%s
File: %s
Changed context:
%s`, policy, file, code)
}
