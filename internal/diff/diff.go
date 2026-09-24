package diff

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type Hunk struct {
	File  string
	Added string
	Start int
}

func Collect(path string, staged bool) ([]Hunk, error) {
	args := []string{"diff", "--unified=40"}
	if staged {
		args = append(args, "--staged")
	}
	if path != "" {
		args = append(args, "--", path)
	}
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("git diff: %w: %s", err, out)
	}
	return parse(string(out)), nil
}

var header = regexp.MustCompile(`^\+\+\+ b/(.+)$`)
var rangeRe = regexp.MustCompile(`@@ -\d+(?:,\d+)? \+(\d+)(?:,(\d+))? @@`)

func parse(s string) []Hunk {
	var result []Hunk
	file := ""
	line := 0
	start := 0
	var buf []string
	flush := func() {
		if file != "" && len(buf) > 0 {
			result = append(result, Hunk{File: file, Added: strings.Join(buf, "\n"), Start: start})
		}
		buf = nil
	}
	for _, l := range strings.Split(s, "\n") {
		if m := header.FindStringSubmatch(l); len(m) > 0 {
			flush()
			file = m[1]
			continue
		}
		if m := rangeRe.FindStringSubmatch(l); len(m) > 0 {
			n, _ := strconv.Atoi(m[1])
			line = n
			start = n
			continue
		}
		if file == "" || strings.HasPrefix(l, "---") || strings.HasPrefix(l, "+++") {
			continue
		}
		if strings.HasPrefix(l, "+") && !strings.HasPrefix(l, "+++") {
			buf = append(buf, strings.TrimPrefix(l, "+"))
			line++
		} else if strings.HasPrefix(l, " ") {
			buf = append(buf, l[1:])
			line++
		}
	}
	flush()
	return result
}
