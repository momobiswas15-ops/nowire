package review

import "testing"

func TestFails(t *testing.T) {
	fs := []Finding{{Severity: "medium"}}
	if Fails(fs, "high") {
		t.Fatal("medium should not fail high")
	}
	if !Fails(fs, "medium") {
		t.Fatal("medium should fail medium")
	}
}
