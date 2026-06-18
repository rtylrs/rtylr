package handlers

import (
	"strings"
	"testing"
	"time"
)

func TestBuildCaretPromptIncludesBusinessContractModeToolsAndContext(t *testing.T) {
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
		"Use read-only tools to gather data before answering.",
		"Available business capability catalog:",
		"analytics.query [read]",
		"crm.list_customers [read]",
		"crm.create_customer [write, approval required]",
		"Structured stream controls:",
		`<tool>{"name":"tool.name","input":{}}</tool>`,
		"never reveal private chain-of-thought",
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
