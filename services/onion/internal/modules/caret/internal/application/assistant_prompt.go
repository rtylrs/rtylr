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
var AssistantBaseTools = []string{ToolSearchName, "web.search", BusinessFindToolName, PortalFindPageToolName, "memory.remember", "account.profile", "utilities.calculate", "utilities.time"}

func BuildAssistantPrompt(request, scope, context string, pageContext map[string]any, now time.Time) string {
	return BuildAssistantPromptWithFiles(request, scope, context, nil, pageContext, now)
}

// BuildAssistantPromptWithFiles also carries the text files attached to the
// request, in their own data region.
func BuildAssistantPromptWithFiles(request, scope, context string, files []AttachedFile, pageContext map[string]any, now time.Time) string {
	return BuildAssistantPromptFor(AssistantPromptInput{Request: request, Scope: scope, Context: context, Files: files, PageContext: pageContext, Now: now})
}

// AssistantPromptInput is everything one turn's prompt is built from.
type AssistantPromptInput struct {
	Request, Scope, Context string
	Files                   []AttachedFile
	PageContext             map[string]any
	Now                     time.Time
	// Routes are tools this person's similar requests needed before.
	Routes ToolRoutes
}

func BuildAssistantPromptFor(in AssistantPromptInput) string {
	request, scope, context, files, pageContext, now := in.Request, in.Scope, in.Context, in.Files, in.PageContext, in.Now
	var catalog strings.Builder
	plannerContext := BuildRequestSystemContext(scope, pageContext)
	tools := append(RoutedToolsForRequest(request, in.Routes), GeneralToolsForRequest(request)...)
	tools = append(tools, PortalToolsForRequest(request)...)
	tools = append(tools, PromptToolsForRequest(ModeAgent, request, plannerContext)...)
	seen := map[string]bool{}
	base := AssistantBaseTools
	if AgentsRequested(request) {
		base = append([]string{AgentsToolName}, base...)
	}
	for _, name := range base {
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
	attached := ""
	if rendered := RenderAttachedFiles(files); rendered != "" {
		attached = "Files the person attached to this message:\n<attached_files>\n" + rendered + "\n</attached_files>\n\n"
	}
	return strings.TrimSpace(strings.NewReplacer(
		"{{FILES}}", attached,
		"{{TODAY}}", describeNow(now),
		"{{TOOLS}}", catalog.String(),
		"{{SCOPE}}", escapeCaretPromptData(plannerContext),
		"{{CONTEXT}}", escapeCaretPromptData(TruncateUTF8ToBytes(context, MaxAssistantContextBytes)),
		"{{REQUEST}}", escapeCaretPromptData(SanitizeUserInput(request, MaxUserPromptLen)),
	).Replace(assistantPromptTemplate))
}

// describeNow spells out the moment in the person's timezone, so "today",
// "this morning" and "end of the week" mean what they mean to them.
func describeNow(now time.Time) string {
	_, offset := now.Zone()
	zone := now.Location().String()
	if zone == "Local" || zone == "" {
		zone = "UTC"
	}
	return now.Format("Monday, January 2, 2006, 15:04") + " in " + zone + " (" + formatUTCOffset(offset) + ")"
}

// LocationFor resolves the person's IANA timezone, falling back to UTC.
func LocationFor(zone string) *time.Location {
	if location, err := loadToolLocation(zone); err == nil && zone != "" {
		return location
	}
	return time.UTC
}
