package output

import (
	"encoding/json"
	"fmt"
	"github.com/nowire/nowire/internal/review"
	"strings"
)

func Terminal(fs []review.Finding) string {
	if len(fs) == 0 {
		return "✓ No issues found.\n"
	}
	var b strings.Builder
	for _, f := range fs {
		icon := "•"
		if f.Severity == "high" {
			icon = "✕"
		} else if f.Severity == "medium" {
			icon = "⚠"
		}
		fmt.Fprintf(&b, "%s %s:%d [%s] %s\n  %s\n", icon, f.File, f.Line, strings.ToUpper(f.Severity), f.Title, f.Message)
		if f.Suggestion != "" {
			fmt.Fprintf(&b, "  fix: %s\n", f.Suggestion)
		}
	}
	return b.String()
}
func Markdown(fs []review.Finding) string {
	var b strings.Builder
	b.WriteString("## nowire review\n\n")
	if len(fs) == 0 {
		return b.String() + "✅ No issues found.\n"
	}
	for _, f := range fs {
		fmt.Fprintf(&b, "- **%s** `%s:%d` — **%s**: %s", strings.ToUpper(f.Severity), f.File, f.Line, f.Title, f.Message)
		if f.Suggestion != "" {
			fmt.Fprintf(&b, " _(suggestion: %s)_", f.Suggestion)
		}
		b.WriteString("\n")
	}
	return b.String()
}

type sarif struct {
	Version string `json:"version"`
	Schema  string `json:"$schema"`
	Runs    []run  `json:"runs"`
}
type run struct {
	Tool    tool     `json:"tool"`
	Results []result `json:"results"`
}
type tool struct {
	Driver driver `json:"driver"`
}
type driver struct {
	Name           string `json:"name"`
	InformationURI string `json:"informationUri"`
	Rules          []rule `json:"rules"`
}
type rule struct {
	ID               string            `json:"id"`
	ShortDescription map[string]string `json:"shortDescription"`
}
type result struct {
	RuleID    string            `json:"ruleId"`
	Level     string            `json:"level"`
	Message   map[string]string `json:"message"`
	Locations []location        `json:"locations"`
}
type location struct {
	PhysicalLocation physical `json:"physicalLocation"`
}
type physical struct {
	ArtifactLocation map[string]string `json:"artifactLocation"`
	Region           map[string]int    `json:"region"`
}

func SARIF(fs []review.Finding) ([]byte, error) {
	rules := []rule{}
	results := []result{}
	for i, f := range fs {
		id := fmt.Sprintf("NW%03d", i+1)
		rules = append(rules, rule{id, map[string]string{"text": f.Title}})
		level := "note"
		if f.Severity == "high" {
			level = "error"
		} else if f.Severity == "medium" {
			level = "warning"
		}
		results = append(results, result{id, level, map[string]string{"text": f.Message}, []location{{physical{map[string]string{"uri": f.File}, map[string]int{"startLine": f.Line}}}}})
	}
	return json.MarshalIndent(sarif{"2.1.0", "https://json.schemastore.org/sarif-2.1.0.json", []run{{tool{driver{"nowire", "https://github.com/nowire/nowire", rules}}, results}}}, "", "  ")
}
