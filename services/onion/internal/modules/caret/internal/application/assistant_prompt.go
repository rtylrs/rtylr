package application

import (
	"strings"
	"time"
)

// BuildAssistantPrompt is the unified chat runtime. Context comes from an
// ownership-checked server adapter, never from client identity fields.
// MaxAssistantContextBytes bounds saved chat context in one prompt. The memory
// module budgets each section below this, so the cap is a backstop only.
const MaxAssistantContextBytes = 72000

// AssistantBaseTools are offered on every turn, whatever the request is
// about: search the web, find more tools, delegate, remember, and the small
// utilities that keep answers exact.
var AssistantBaseTools = []string{ToolSearchName, "web.search", "memory.remember", "account.profile", "utilities.calculate", "utilities.time"}

func BuildAssistantPrompt(request, scope, context string, pageContext map[string]any, now time.Time) string {
	var catalog strings.Builder
	plannerContext := BuildRequestSystemContext(scope, pageContext)
	tools := PromptToolsForRequest(ModeAgent, request, plannerContext)
	seen := map[string]bool{}
	for _, name := range AssistantBaseTools {
		if tool, ok := ToolByName(name); ok {
			tools = append(tools, tool)
		}
	}
	for _, tool := range tools {
		if seen[tool.Name] {
			continue
		}
		seen[tool.Name] = true
		catalog.WriteString("- " + toolContractPrompt(tool) + ": " + tool.Description + "\n")
	}
	return strings.TrimSpace(strings.NewReplacer(
		"{{TODAY}}", now.Format(time.RFC3339),
		"{{TOOLS}}", catalog.String(),
		"{{SCOPE}}", escapeCaretPromptData(plannerContext),
		"{{CONTEXT}}", escapeCaretPromptData(TruncateUTF8ToBytes(context, MaxAssistantContextBytes)),
		"{{REQUEST}}", escapeCaretPromptData(SanitizeUserInput(request, MaxUserPromptLen)),
	).Replace(assistantPromptTemplate))
}
