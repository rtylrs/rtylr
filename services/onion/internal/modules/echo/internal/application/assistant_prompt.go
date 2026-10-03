package application

import (
	"strings"
	"time"
)

// BuildAssistantPrompt is the unified chat runtime. Context comes from an
// ownership-checked server adapter, never from client identity fields.
func BuildAssistantPrompt(request, scope, context string, pageContext map[string]any, now time.Time) string {
	var catalog strings.Builder
	tools := PromptToolsForMode(ModeAgent)
	seen := map[string]bool{}
	for _, name := range []string{"web.search", "account.profile", "utilities.calculate", "utilities.time"} {
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
		"{{SCOPE}}", escapeCaretPromptData(BuildRequestSystemContext(scope, pageContext)),
		"{{CONTEXT}}", escapeCaretPromptData(TruncateUTF8ToBytes(context, 32000)),
		"{{REQUEST}}", escapeCaretPromptData(SanitizeUserInput(request, MaxUserPromptLen)),
	).Replace(assistantPromptTemplate))
}
