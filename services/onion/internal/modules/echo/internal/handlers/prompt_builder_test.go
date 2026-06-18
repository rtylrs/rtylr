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
		"Mode:\nanalyst",
		"server-provided analytics grounding",
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

func TestBuildCaretPromptAgentModeIncludesWriteToolProposalsOnly(t *testing.T) {
	got := buildCaretPrompt(caretPromptInput{
		Mode:               ModeAgent,
		ConversationPrompt: "Current request: create a customer",
		Now:                time.Date(2026, 6, 18, 12, 0, 0, 0, time.UTC),
	})

	mustContain := []string{
		"Available business capability catalog:",
		"crm.create_customer [write, approval required]",
		"Structured stream controls:",
		`<tool>{"name":"tool.name","input":{}}</tool>`,
		"never reveal private chain-of-thought",
	}
	for _, want := range mustContain {
		if !strings.Contains(got, want) {
			t.Fatalf("agent prompt missing %q\nprompt:\n%s", want, got)
		}
	}
	for _, forbidden := range []string{"crm.list_customers [read]", "pos.list_orders [read]", "analytics.query [read]"} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("agent prompt should not advertise non-executable read tool %q\nprompt:\n%s", forbidden, got)
		}
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
