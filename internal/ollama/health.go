package ollama

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Tag struct {
	Name string `json:"name"`
}
type tagsResponse struct {
	Models []Tag `json:"models"`
}

func (c *Client) Tags(ctx context.Context) ([]Tag, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama unavailable: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return nil, fmt.Errorf("ollama: %s", res.Status)
	}
	var out tagsResponse
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("decode ollama models: %w", err)
	}
	return out.Models, nil
}

func HasModel(models []Tag, wanted string) bool {
	for _, model := range models {
		if model.Name == wanted {
			return true
		}
	}
	return false
}
