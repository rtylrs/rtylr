package application

import "strings"

// CaretMode is the conversation persona requested by the client. The list
// mirrors the dash app's `CARET_MODES` enum so the wire stays stable.
type CaretMode string

const (
	ModeAsk     CaretMode = "ask"
	ModeAgent   CaretMode = "agent"
	ModeAnalyst CaretMode = "analyst"
	ModeSupport CaretMode = "support"
)

// SystemPromptAddendum returns the per-mode text we append to the system
// prompt before invoking the model.
func (m CaretMode) SystemPromptAddendum() string {
	switch m {
	case ModeAgent:
		return "You are in Agent mode. You may use read tools and propose write tools. " +
			"Move broad requests forward with the current page context, summarize the plain-language plan before any mutating tool, and request approval before claiming changes."
	case ModeAnalyst:
		return "You are in Analyst mode. Use read-only business and analytics tools for numbers, trends, and comparisons. " +
			"Put the number or table first, keep interpretation tied to the current business area, do not describe internal queries, and do not propose write tools."
	case ModeSupport:
		return "You are in Support mode. Help the user navigate the product and explain features from the current app context. " +
			"Do not call business tools or mutate data."
	default:
		return "You are in Ask mode. Answer questions concisely and use read-only tools when live business data is needed. " +
			"Use the current screen as context for vague wording, and do not invoke or propose any tool that modifies data."
	}
}

// AllowsTools returns whether tool execution is enabled for this mode.
func (m CaretMode) AllowsTools() bool {
	return m != ModeSupport
}

// ParseMode coerces an unknown string to a CaretMode, defaulting to Ask.
func ParseMode(s string) CaretMode {
	switch CaretMode(strings.ToLower(strings.TrimSpace(s))) {
	case ModeAgent:
		return ModeAgent
	case ModeAnalyst:
		return ModeAnalyst
	case ModeSupport:
		return ModeSupport
	default:
		return ModeAsk
	}
}
