package review

import "testing"

func TestParseFindingsAcceptsFencedJSON(t *testing.T) {
	findings, err := parseFindings("```json\n[{\"severity\":\"high\",\"title\":\"bug\"}]\n```")
	if err != nil || len(findings) != 1 || findings[0].Title != "bug" {
		t.Fatalf("unexpected parse result: %#v, %v", findings, err)
	}
}

func TestParseFindingsAcceptsEnvelope(t *testing.T) {
	findings, err := parseFindings(`{"findings":[{"severity":"low","title":"nit"}]}`)
	if err != nil || len(findings) != 1 || findings[0].Title != "nit" {
		t.Fatalf("unexpected parse result: %#v, %v", findings, err)
	}
}
