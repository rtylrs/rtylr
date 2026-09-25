package application

import _ "embed"

// caretSystemPromptTemplate is the single source of truth for Caret's durable
// behavior. Request-specific values are inserted by BuildCaretPrompt.
//
//go:embed system_prompt.txt
var caretSystemPromptTemplate string
