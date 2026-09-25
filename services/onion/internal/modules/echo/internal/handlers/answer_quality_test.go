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

func TestPolishCaretAnswerTextHidesInternalCapabilityDetails(t *testing.T) {
	got := polishCaretAnswerText("I obtained this information using the insights.get_customer_kpis and insights.get_product_kpis tools, which are part of the shortlisted capability catalog for the current request.")

	for _, banned := range []string{
		"insights.get_customer_kpis",
		"insights.get_product_kpis",
		"shortlisted capability catalog",
	} {
		if strings.Contains(strings.ToLower(got), strings.ToLower(banned)) {
			t.Fatalf("polished answer still contains internal detail %q: %q", banned, got)
		}
	}
	if !strings.Contains(got, "business data") {
		t.Fatalf("expected business-language attribution, got %q", got)
	}
}

func TestPolishCaretAnswerTextRemovesProcessNarrationAroundInternalDetails(t *testing.T) {
	got := polishCaretAnswerText("I obtained this information using the customer data and product data tools, which are part of the available business capabilities for the current request.")

	if strings.Contains(strings.ToLower(got), "using") || strings.Contains(strings.ToLower(got), "capabilities") {
		t.Fatalf("process narration crossed the public answer boundary: %q", got)
	}
	if got != "I checked your business data." {
		t.Fatalf("expected a concise business-language answer, got %q", got)
	}
}

func TestPolishCaretAnswerTextRemovesInternalNextAction(t *testing.T) {
	got := polishCaretAnswerText("The customer check is complete.\nNext Move: Run the insights.get_customer_kpis and insights.get_product_kpis tools to obtain the latest data.")

	if strings.Contains(got, "Next Move") || strings.Contains(got, "insights.") || strings.Contains(got, "tools") {
		t.Fatalf("internal next action crossed the public answer boundary: %q", got)
	}
	if !strings.Contains(got, "I can check the latest business data if you want.") {
		t.Fatalf("expected safe next action, got %q", got)
	}
}

func TestSafePublicAnswerPrefixDoesNotSplitInternalCapabilityDetails(t *testing.T) {
	text := "I used insights.get_customer_kpis to check the customer data."
	cut := strings.Index(text, "customer_kpis")
	if cut <= 0 {
		t.Fatal("test input did not contain the expected internal name")
	}

	got := safePublicAnswerPrefix(text, cut+4)
	if got != strings.Index(text, "insights.get_customer_kpis") {
		t.Fatalf("safe prefix ended inside an internal name: got %d from %q", got, text[:cut+4])
	}
}

func TestSafePublicAnswerPrefixHoldsUnknownInternalCapabilityPrefix(t *testing.T) {
	text := "I used erp.create_product to add the item."
	cut := strings.Index(text, "create_product")
	if cut <= 0 {
		t.Fatal("test input did not contain the expected internal name")
	}

	got := safePublicAnswerPrefix(text, cut+5)
	if got != strings.Index(text, "erp.create_product") {
		t.Fatalf("safe prefix ended inside an unknown internal name: got %d from %q", got, text[:cut+5])
	}
}
