package diff

import "testing"

func TestParseHunk(t *testing.T) {
	got := parse("diff --git a/a.go b/a.go\n--- a/a.go\n+++ b/a.go\n@@ -1 +1,2 @@\n package main\n+fmt.Println(1)\n")
	if len(got) != 1 || got[0].File != "a.go" || got[0].Start != 1 {
		t.Fatalf("unexpected: %#v", got)
	}
}
