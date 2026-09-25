package ollama

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGenerateRequestsJSONAndRetriesTransientFailures(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if r.URL.Path != "/api/generate" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if attempts < 2 {
			http.Error(w, "busy", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"response":"[]"}`))
	}))
	defer server.Close()
	got, err := New(server.URL).Generate(context.Background(), "test", "review")
	if err != nil || got != "[]" || attempts != 2 {
		t.Fatalf("got %q, err %v, attempts %d", got, err, attempts)
	}
}

func TestHasModel(t *testing.T) {
	if !HasModel([]Tag{{Name: "coder:7b"}}, "missing") == false {
		t.Fatal("unexpected model match")
	}
	if !HasModel([]Tag{{Name: "coder:7b"}}, "coder:7b") {
		t.Fatal("expected model match")
	}
}

func TestRenderTemplateReplacesSupportedPlaceholders(t *testing.T) {
	got := RenderTemplate("policy={{policy}} file={{file}} code={{code}}", "strict", "main.go", "return nil")
	if got != "policy=strict file=main.go code=return nil" {
		t.Fatalf("unexpected template rendering: %q", got)
	}
}
