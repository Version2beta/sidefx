package main

import (
	"os"
	"strings"
	"testing"
)

// This guards the documented route rules and their order; it does not establish agent behavior.
func TestGuidanceDocumentsStateBothRouteRules(t *testing.T) {
	guides := map[string]struct {
		path         string
		focused      string
		focusedRules []string
		broad        string
		broadRule    string
	}{
		"canonical": {
			path:    "../../skills/sidefx/SKILL.md",
			focused: "**Directed investigation:**",
			focusedRules: []string{
				"Start with a relevant focused sidefx CLI query",
				"without first acquiring/indexing all categories",
			},
			broad:     "**Open-ended discovery:**",
			broadRule: "Complete category orientation within the declared scope before selecting leads",
		},
		"function-first": {
			path:    "../../.agents/skills/sidefx-function-first/SKILL.md",
			focused: "For a **focused question or directed lead**",
			focusedRules: []string{
				"start with one narrow **sidefx query**",
				"Do not require complete category orientation",
			},
			broad:     "For **broad prospective function discovery**",
			broadRule: "Read EVERY category ID and coverage before prospective selection",
		},
		"operation-index": {
			path:    "../../docs/operation-index.md",
			focused: "For a focused question or directed lead",
			focusedRules: []string{
				"use the relevant narrow CLI query",
				"complete category orientation is not required",
			},
			broad:     "For broad prospective discovery",
			broadRule: "complete the category orientation before selecting leads",
		},
	}
	for name, guide := range guides {
		data, err := os.ReadFile(guide.path)
		if err != nil {
			t.Fatal(name, err)
		}
		text := strings.ToLower(strings.Join(strings.Fields(string(data)), " "))
		focusedAt := strings.Index(text, strings.ToLower(guide.focused))
		broadAt := strings.Index(text, strings.ToLower(guide.broad))
		if focusedAt < 0 || broadAt <= focusedAt {
			t.Errorf("%s guidance must present the focused route before the broad route", name)
			continue
		}
		for _, rule := range guide.focusedRules {
			if !strings.Contains(text[focusedAt:broadAt], strings.ToLower(rule)) {
				t.Errorf("%s focused route is missing rule %q", name, rule)
			}
		}
		if !strings.Contains(text[broadAt:], strings.ToLower(guide.broadRule)) {
			t.Errorf("%s broad route is missing its orientation rule", name)
		}
	}
}
