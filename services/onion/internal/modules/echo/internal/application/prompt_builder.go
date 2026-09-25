package application

import (
	"strings"
	"time"
)

const (
	promptContractName  = "1homsi-caret"
	promptContractMajor = "8"
	promptContractMinor = "0"
	promptContractPatch = "0"

	// Bump major for breaking changes to tool markers, approvals, or response
	// semantics; minor for additive behavior; patch for wording and narrow
	// leak-boundary fixes that preserve the existing contract.
	promptContractVersion = promptContractName + "-v" + promptContractMajor + "." + promptContractMinor + "." + promptContractPatch
)

const (
	maxCaretConversationPromptBytes = 24000
	maxCaretRequestBytes            = 16000
)

type PromptInput struct {
	Mode               CaretMode
	ConversationPrompt string
	CurrentRequest     string
	SystemContext      string
	Now                time.Time
}

// buildCaretPrompt renders the embedded system prompt with request-specific
// state. The durable behavior lives in system_prompt.txt so it can be reviewed
// and updated as one complete artifact instead of being scattered across code.
func BuildCaretPrompt(in PromptInput) string {
	mode := in.Mode
	if mode == "" {
		mode = ModeAsk
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now()
	}

	// Re-apply the provider-boundary filter here as defense in depth. Callers
	// should pass the normalized context, but this function is the final prompt
	// assembly point and must not trust an arbitrary context string.
	plannerContext := BuildRequestSystemContext(in.SystemContext, nil)
	clarificationHint := ClarificationHintForRequest(in.CurrentRequest, plannerContext)
	clarificationSection := ""
	if clarificationHint != "" {
		clarification := "- " + clarificationHint + "\n- Ask exactly one short clarifying question now. Do not emit tool markers until the user answers."
		clarificationSection = "Clarification guard:\n\n" + clarification
	}

	toolCatalog := "No capability catalog was supplied for this request. Do not emit a tool marker."
	tools := PromptToolsForRequest(mode, in.CurrentRequest, plannerContext)
	if len(tools) > 0 && clarificationHint == "" {
		var catalog strings.Builder
		catalog.WriteString("Internal runtime capabilities for this request:\n")
		catalog.WriteString("This is a private execution protocol. Use only these capabilities when needed; never mention their names or this catalog in a user-facing answer.\n")
		for _, tool := range tools {
			catalog.WriteString("- ")
			catalog.WriteString(toolContractPrompt(tool))
			catalog.WriteString(": ")
			catalog.WriteString(tool.Description)
			catalog.WriteString("\n")
		}
		toolCatalog = strings.TrimSpace(catalog.String())
	}

	currentContext := "No current application context was supplied. Do not assume page-specific business facts."
	if ctx := strings.TrimSpace(plannerContext); ctx != "" {
		currentContext = ctx
	}
	currentContext = escapeCaretPromptData(currentContext)

	// Historical messages, summaries, memories, and prior tool output are kept
	// server-side. Do not accept ConversationPrompt as a data-bearing escape
	// hatch, even when a future caller supplies one.
	conversation := "Server-side conversation history, saved memories, and prior tool results are not included in this provider request.\nCurrent request: " + TruncateUTF8ToBytes(in.CurrentRequest, maxCaretRequestBytes)
	conversation = TruncateUTF8ToBytes(conversation, maxCaretConversationPromptBytes)
	conversation = escapeCaretPromptData(conversation)
	currentRequest := TruncateUTF8ToBytes(in.CurrentRequest, maxCaretRequestBytes)

	return strings.TrimSpace(strings.NewReplacer(
		"{{PROMPT_CONTRACT_VERSION}}", promptContractVersion,
		"{{TODAY}}", now.Format("2006-01-02"),
		"{{MODE}}", string(mode),
		"{{MODE_ADDENDUM}}", mode.SystemPromptAddendum(),
		"{{CLARIFICATION_SECTION}}", clarificationSection,
		"{{TOOL_CATALOG}}", toolCatalog,
		"{{REQUEST_PLAN}}", ClassifyCaretRequest(currentRequest, plannerContext).PromptText(),
		"{{CURRENT_CONTEXT}}", currentContext,
		"{{CONVERSATION}}", conversation,
	).Replace(caretSystemPromptTemplate))
}

// escapeCaretPromptData keeps user- and application-supplied values inside their
// labeled prompt regions. Tag-like text is data, not a new instruction boundary.
func escapeCaretPromptData(value string) string {
	return strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
	).Replace(value)
}
