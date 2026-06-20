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

func TestPolishCaretAnswerTextSoftensProcessNarration(t *testing.T) {
	got := polishCaretAnswerText("To answer your question, I ran a server-validated, read-only analytics query scoped to your tenant. The query results are:\n| Customers |\n| --- |\n| 2 |")

	for _, banned := range []string{"ran a", "server-validated", "read-only analytics query", "tenant", "The query results are"} {
		if strings.Contains(got, banned) {
			t.Fatalf("polished answer still contains %q: %q", banned, got)
		}
	}
	if !strings.Contains(got, "I checked your business data.") {
		t.Fatalf("expected softer business phrasing, got %q", got)
	}
}
