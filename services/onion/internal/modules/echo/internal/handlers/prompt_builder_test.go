package handlers

import (
	"strings"
	"testing"
	"time"
)

func TestBuildCaretPromptIncludesBusinessContractModeAndContext(t *testing.T) {
	got := buildCaretPrompt(caretPromptInput{
		Mode:               ModeAnalyst,
		ConversationPrompt: "Current request: revenue today",
		SystemContext:      "Route: /finance/dashboard\nVisible module: finance",
		Now:                time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
	})

	mustContain := []string{
		caretSystemPreamble,
		"Prompt contract: caret-business-v1",
		"Today: 2026-06-18",
		"Act like a capable business operator",
		"Write like a calm, sharp business teammate",
		"Lead with the answer, not the process",
		"Never start an answer with \"Based on the provided analytics query\"",
		"Mode:\nanalyst",
		"server-provided analytics evidence silently",
		"Current application context:",
		"Route: /finance/dashboard",
		"Conversation, memory, and current request:",
		"Current request: revenue today",
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Fatalf("prompt missing %q\nprompt:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Available business capability catalog:") {
		t.Fatalf("analyst prompt should not invite model-emitted tool calls:\n%s", got)
	}
}

func TestBuildCaretPromptAgentModeIncludesExecutableToolProposalsOnly(t *testing.T) {
	got := buildCaretPrompt(caretPromptInput{
		Mode:               ModeAgent,
		ConversationPrompt: "Current request: create a customer",
		CurrentRequest:     "create a customer",
		Now:                time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
	})

	mustContain := []string{
		"Available business capability catalog:",
		"This catalog is shortlisted for the current request",
		"business.snapshot [read]",
		"business.get_kpis [read]",
		"insights.get_dashboard [read]",
		"crm.list_customers [read]",
		"crm.get_customer [read]",
		"crm.list_campaigns [read]",
		"crm.get_campaign [read]",
		"crm.list_segments [read]",
		"crm.create_customer [write, approval required]",
		"Structured stream controls:",
		`<tool>{"name":"tool.name","input":{}}</tool>`,
		"Read tools run immediately",
		"never reveal private chain-of-thought",
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Fatalf("agent prompt missing %q\nprompt:\n%s", want, got)
		}
	}
	for _, forbidden := range []string{"analytics.query [read]", "finance.list_invoices [read]", "pos.list_tables [read]"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("agent prompt should not advertise non-executable read tool %q\nprompt:\n%s", forbidden, got)
		}
	}
}

func TestPromptToolsForRequestShortlistsRelevantTools(t *testing.T) {
	tools := PromptToolsForRequest(ModeAgent, "show payroll runs for this month", "/hr/payroll")
	names := make([]string, 0, len(tools))
	for _, tool := range tools {
		names = append(names, tool.Name)
	}
	joined := strings.Join(names, ",")
	for _, want := range []string{"business.snapshot", "business.get_kpis", "hr.list_payroll_runs", "hr.get_payroll_run"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("shortlist missing %s: %v", want, names)
		}
	}
	for _, forbidden := range []string{"store.get_product", "support.get_ticket"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("shortlist included unrelated %s: %v", forbidden, names)
		}
	}
	if len(tools) > promptToolShortlistLimit {
		t.Fatalf("shortlist length = %d, want <= %d", len(tools), promptToolShortlistLimit)
	}
}

func TestBuildCaretPromptClarifiesAmbiguousEntityBeforeTools(t *testing.T) {
	got := buildCaretPrompt(caretPromptInput{
		Mode:               ModeAgent,
		ConversationPrompt: "Current request: update this product",
		CurrentRequest:     "update this product",
		SystemContext:      "Structured page context:\n{\"app\":\"dash\",\"route\":\"/menus\"}",
		Now:                time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
	})

	for _, want := range []string{
		"Clarification guard:",
		"The product is not identified",
		"Ask exactly one short clarifying question",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("prompt missing clarification guard %q\nprompt:\n%s", want, got)
		}
	}
	if strings.Contains(got, "Available business capability catalog:") {
		t.Fatalf("ambiguous prompt should not advertise tool calls:\n%s", got)
	}
}

func TestBuildCaretPromptAllowsToolsWithSelectedEntityContext(t *testing.T) {
	got := buildCaretPrompt(caretPromptInput{
		Mode:               ModeAgent,
		ConversationPrompt: "Current request: update this product",
		CurrentRequest:     "update this product",
		SystemContext:      "Structured page context:\n{\"selected\":{\"product_id\":\"prod_123\"}}",
		Now:                time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
	})

	if strings.Contains(got, "Clarification guard:") {
		t.Fatalf("selected entity context should avoid clarification guard:\n%s", got)
	}
	if !strings.Contains(got, "Available business capability catalog:") {
		t.Fatalf("selected entity context should keep relevant tools available:\n%s", got)
	}
}

func TestBuildCaretPromptSupportModeOmitsToolCatalog(t *testing.T) {
	got := buildCaretPrompt(caretPromptInput{
		Mode:               ModeSupport,
		ConversationPrompt: "Current request: where do I find invoices?",
		Now:                time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
	})

	if strings.Contains(got, "Available business capability catalog:") {
		t.Fatalf("support prompt should not expose tool catalog:\n%s", got)
	}
	if !strings.Contains(got, "Mode:\nsupport") {
		t.Fatalf("support prompt missing mode:\n%s", got)
	}
}
