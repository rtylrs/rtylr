package handlers

import (
	"fmt"
	"strings"
	"time"
)

const promptContractVersion = "caret-business-v1"

const caretHumanVoiceContract = `Voice and answer style:
- Write like a calm, sharp business teammate speaking to a store owner.
- Lead with the answer, not the process. Prefer "Your top customers are..." over "Based on the provided query...".
- Do not mention internal mechanics unless the user asks how you got the answer. Avoid words like query, tool, grounding, spec, rows returned, server-validated, tenant, or JSON in normal answers.
- Never start an answer with "Based on the provided analytics query", "Based on the data provided", or similar process disclaimers.
- Use "your business", "your store", "your team", or the actual app/module name instead of "tenant".
- Keep headings natural and useful. Prefer "What stands out" or "Good next moves" over generic "Key Takeaways" and "Next Steps" unless those labels genuinely fit.
- If data is missing, say it plainly in business language: "I do not see sales for today yet" instead of "the revenue value is null".
- When the answer is a number, put the number first. When the answer is a list, put the list/table first and add interpretation after.
`

type caretPromptInput struct {
	Mode               CaretMode
	ConversationPrompt string
	SystemContext      string
	Now                time.Time
}

// buildCaretPrompt renders a durable system envelope around the memory module's
// conversation prompt. The provider API currently receives one text prompt, so
// strong section boundaries matter: they keep policy, app context, memory, and
// the user's current request from blurring together.
func buildCaretPrompt(in caretPromptInput) string {
	mode := in.Mode
	if mode == "" {
		mode = ModeAsk
	}
	now := in.Now
	if now.IsZero() {
		now = time.Now()
	}

	var b strings.Builder
	b.WriteString(caretSystemPreamble)
	b.WriteString("\nPrompt contract: ")
	b.WriteString(promptContractVersion)
	b.WriteString("\nToday: ")
	b.WriteString(now.Format("2006-01-02"))
	b.WriteString("\n\n")

	b.WriteString("Operating contract:\n")
	b.WriteString("- Act like a capable business operator, not a generic chatbot.\n")
	b.WriteString("- Be concise by default, but include the numbers, caveats, and next steps a real operator needs.\n")
	b.WriteString("- Never invent business data, IDs, policies, permissions, prices, dates, or execution results.\n")
	b.WriteString("- When live data or context is missing, say what is missing and give the safest next step.\n")
	b.WriteString("- Preserve tenant and user privacy; do not suggest cross-tenant access or expose hidden implementation details.\n")
	b.WriteString("- For destructive or mutating actions, explain the plan and ask for explicit approval before claiming anything changed.\n")
	b.WriteString("- Prefer tables for comparisons, bullets for action lists, and short prose for explanations.\n\n")
	b.WriteString(caretHumanVoiceContract)
	b.WriteString("\n")

	b.WriteString("Mode:\n")
	b.WriteString(string(mode))
	b.WriteString(" - ")
	b.WriteString(mode.SystemPromptAddendum())
	b.WriteString("\n\n")

	tools := PromptToolsForMode(mode)
	if len(tools) > 0 {
		b.WriteString("Available business capability catalog:\n")
		for _, tool := range tools {
			b.WriteString("- ")
			b.WriteString(tool.Name)
			b.WriteString(" [")
			b.WriteString(tool.Scope)
			if tool.RequiresApproval {
				b.WriteString(", approval required")
			}
			b.WriteString("]: ")
			b.WriteString(tool.Description)
			b.WriteString("\n")
		}
		b.WriteString("Use this catalog to reason about what the product can help with. Do not claim a tool ran unless the platform gives you a result.\n\n")
		b.WriteString("Structured stream controls:\n")
		b.WriteString("- If you need to propose a tool call, emit exactly one compact marker like <tool>{\"name\":\"tool.name\",\"input\":{}}</tool>.\n")
		b.WriteString("- Read tools run immediately when the platform can execute them; use their returned result instead of guessing.\n")
		b.WriteString("- Write tools are approval proposals only; after a write-tool marker, wait for approval and do not say the action succeeded.\n")
		b.WriteString("- Do not emit analytics.query yourself; the server runs tenant-scoped analytics grounding automatically when it is safe.\n")
		b.WriteString("- You may use <thinking>short user-facing progress note</thinking> for brief status notes, but never reveal private chain-of-thought.\n\n")
	}

	if ctx := strings.TrimSpace(in.SystemContext); ctx != "" {
		b.WriteString("Current application context:\n")
		b.WriteString(ctx)
		b.WriteString("\n\n")
	}

	b.WriteString("Conversation, memory, and current request:\n")
	conversation := strings.TrimSpace(in.ConversationPrompt)
	if conversation == "" {
		conversation = "Current request: "
	}
	b.WriteString(conversation)
	b.WriteString("\n\n")
	b.WriteString("Answer now using the operating contract above.")

	return strings.TrimSpace(b.String())
}

func newAssistantMessageID() string {
	return fmt.Sprintf("assistant-%d", time.Now().UnixNano())
}
