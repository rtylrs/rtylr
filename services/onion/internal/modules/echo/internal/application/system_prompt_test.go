package application

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
)

func TestEmbeddedSystemPromptContainsAtLeast200UniqueNumberedRules(t *testing.T) {
	rulePattern := regexp.MustCompile(`(?m)^(\d{3})\.`)
	matches := rulePattern.FindAllStringSubmatch(caretSystemPromptTemplate, -1)
	if len(matches) < 200 {
		t.Fatalf("embedded system prompt contains %d numbered rules, want at least 200", len(matches))
	}

	seen := make(map[int]bool, len(matches))
	for _, match := range matches {
		ruleID, err := strconv.Atoi(match[1])
		if err != nil {
			t.Fatalf("invalid rule id %q: %v", match[1], err)
		}
		if seen[ruleID] {
			t.Fatalf("system prompt repeats rule %03d", ruleID)
		}
		seen[ruleID] = true
	}
	for ruleID := 1; ruleID <= 200; ruleID++ {
		if !seen[ruleID] {
			t.Fatalf("system prompt is missing rule %03d", ruleID)
		}
	}
}

func TestRenderedSystemPromptHasNoUnresolvedControlsOrSampleBusinessFacts(t *testing.T) {
	got := BuildCaretPrompt(PromptInput{
		Mode:           ModeAsk,
		CurrentRequest: "test",
	})

	if regexp.MustCompile(`\{\{[A-Z_]+\}\}`).MatchString(got) {
		t.Fatalf("rendered system prompt contains unresolved controls: %s", got)
	}
	for _, forbidden := range []string{"John Smith", "Jane Doe", "Bob Brown", "$10,000", "C001"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("system prompt contains fabricated sample business fact %q", forbidden)
		}
	}
}
