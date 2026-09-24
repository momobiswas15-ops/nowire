package output

import (
	"strings"
	"testing"

	"github.com/nowire/nowire/internal/review"
)

func TestGitHubAnnotationsEscapeControlCharacters(t *testing.T) {
	got := GitHubAnnotations([]review.Finding{{Severity: "high", File: "a,b.go", Line: 4, Title: "bad: input", Message: "line one\nline two"}})
	for _, want := range []string{"file=a%2Cb.go", "title=bad%3A input", "%0A"} {
		if !strings.Contains(got, want) {
			t.Fatalf("annotation %q missing %q", got, want)
		}
	}
}
