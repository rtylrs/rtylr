package handlers

import (
	"strings"
	"testing"
)

func TestPolishCaretAnswerTextRemovesInternalPhrases(t *testing.T) {
	got := polishCaretAnswerText("Based on the provided analytics query, your tenant has 2 rows returned. Invalid tool call marker ignored.")

	for _, banned := range []string{"Based on", "analytics query", "tenant", "rows returned", "Invalid tool call"} {
		if strings.Contains(got, banned) {
			t.Fatalf("polished answer still contains %q: %q", banned, got)
		}
	}
	if !strings.Contains(got, "your business") {
		t.Fatalf("expected tenant language to become business language, got %q", got)
	}
}
