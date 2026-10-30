package application

import (
	"strings"
	"testing"
	"time"
)

func TestPortalLayoutToolsComeWithLayoutWords(t *testing.T) {
	for _, request := range []string{
		"hey caret move yourself to the top",
		"restore the default rtylr sidebar",
		"hide POS from the app bar",
		"رتب الشريط الجانبي",
		"remets la barre latérale par défaut",
	} {
		names := []string{}
		for _, tool := range PortalToolsForRequest(request) {
			names = append(names, tool.Name)
		}
		if strings.Join(names, ",") != "portal.layout,portal.customize,portal.restore_default" {
			t.Errorf("%q offered %v", request, names)
		}
	}
	if tools := PortalToolsForRequest("what were sales yesterday"); len(tools) != 0 {
		t.Errorf("a sales question offered %v", tools)
	}
}

func TestEveryPromptCanFindPortalPages(t *testing.T) {
	prompt := BuildAssistantPromptFor(AssistantPromptInput{Request: "where can I add team members", Now: time.Now()})
	if !strings.Contains(prompt, "- portal.find_page") {
		t.Fatal("portal.find_page is missing from the prompt")
	}
	if strings.Contains(prompt, "- portal.customize") {
		t.Fatal("layout tools should wait for a layout request")
	}
	prompt = BuildAssistantPromptFor(AssistantPromptInput{Request: "move yourself to the top", Now: time.Now()})
	if !strings.Contains(prompt, "- portal.customize") {
		t.Fatal("portal.customize is missing for a layout request")
	}
}

func TestPortalLayoutChangesRunWithoutApproval(t *testing.T) {
	for _, name := range []string{"portal.customize", "portal.restore_default"} {
		tool, ok := ToolByName(name)
		if !ok || tool.Scope != "write" || !tool.Personal || tool.RequiresApproval {
			t.Errorf("%s = %+v", name, tool)
		}
		if ToolAllowedForMode(ModeAsk, tool) {
			t.Errorf("%s must stay out of read-only modes", name)
		}
	}
}
