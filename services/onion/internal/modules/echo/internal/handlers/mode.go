package handlers

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
		return "You are in Agent mode. You may call tools to perform actions. Summarize " +
			"your plan before mutating tools and request approval for irreversible actions."
	case ModeAnalyst:
		return "You are in Analyst mode. Prefer concise tables and bullets. Use the " +
			"server-provided analytics grounding when it is present; do not propose unrelated tools."
	case ModeSupport:
		return "You are in Support mode. Help the user navigate the product and explain " +
			"features. Never mutate data without explicit approval."
	default:
		return "You are in Ask mode. Answer questions concisely. Do not invoke any tool " +
			"that modifies data."
	}
}

// AllowsTools returns whether tool execution is enabled for this mode.
func (m CaretMode) AllowsTools() bool {
	return m != ModeSupport
}

// ParseMode coerces an unknown string to a CaretMode, defaulting to Ask.
func ParseMode(s string) CaretMode {
	switch CaretMode(s) {
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
