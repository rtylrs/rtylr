package application

import (
	"strings"
	"time"
)

const promptContractVersion = "caret-business-v5"

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

	clarificationHint := ClarificationHintForRequest(in.CurrentRequest, in.SystemContext)
	clarification := "No clarification is needed. Continue with the current request."
	clarificationSection := ""
	if clarificationHint != "" {
		clarification = "- " + clarificationHint + "\n- Ask exactly one short clarifying question now. Do not emit tool markers until the user answers."
		clarificationSection = "Clarification guard:\n\n" + clarification
	}

	toolCatalog := "No capability catalog was supplied for this request. Do not emit a tool marker."
	tools := PromptToolsForRequest(mode, in.CurrentRequest, in.SystemContext)
	if len(tools) > 0 && clarificationHint == "" {
		var catalog strings.Builder
		catalog.WriteString("Available business capability catalog:\n")
		catalog.WriteString("This catalog is shortlisted for the current request; use these tools before reaching for unrelated capabilities.\n")
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
	if ctx := strings.TrimSpace(in.SystemContext); ctx != "" {
		currentContext = ctx
	}

	conversation := strings.TrimSpace(in.ConversationPrompt)
	if conversation == "" {
		conversation = "Current request: "
	}
	conversation = TruncateUTF8ToBytes(conversation, maxCaretConversationPromptBytes)
	currentRequest := TruncateUTF8ToBytes(in.CurrentRequest, maxCaretRequestBytes)

	return strings.TrimSpace(strings.NewReplacer(
		"{{PROMPT_CONTRACT_VERSION}}", promptContractVersion,
		"{{TODAY}}", now.Format("2006-01-02"),
		"{{MODE}}", string(mode),
		"{{MODE_ADDENDUM}}", mode.SystemPromptAddendum(),
		"{{CLARIFICATION_SECTION}}", clarificationSection,
		"{{TOOL_CATALOG}}", toolCatalog,
		"{{REQUEST_PLAN}}", ClassifyCaretRequest(currentRequest, in.SystemContext).PromptText(),
		"{{CURRENT_CONTEXT}}", currentContext,
		"{{CONVERSATION}}", conversation,
	).Replace(caretSystemPromptTemplate))
}
