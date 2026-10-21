package application

import (
	"context"
	"strings"
	"testing"
	"time"
)

func toolNames(tools []ToolDef) []string {
	names := make([]string, len(tools))
	for i, tool := range tools {
		names[i] = tool.Name
	}
	return names
}

func TestSearchToolsFindsToolsFromEverydayWords(t *testing.T) {
	cases := map[string]string{
		"who do i have in the team": "hr.list_employees",
		"unpaid invoices":           "finance.list_invoices",
		"my tasks for this week":    "flow.list_tasks",
		"open support tickets":      "support.list_tickets",
	}
	for query, want := range cases {
		got := toolNames(SearchTools(ModeAgent, query))
		if !strings.Contains(strings.Join(got, " "), want) {
			t.Errorf("SearchTools(%q) = %v, want %s among them", query, got, want)
		}
		if len(got) > toolSearchLimit {
			t.Errorf("SearchTools(%q) returned %d tools, limit is %d", query, len(got), toolSearchLimit)
		}
	}
}

func TestSearchToolsOnlyReturnsExecutableToolsAndNeverItself(t *testing.T) {
	for _, tool := range SearchTools(ModeAgent, "tools search everything list get create") {
		if tool.Name == ToolSearchName {
			t.Fatal("tools.search must not return itself")
		}
		if !promptToolCanExecute(tool.Name) {
			t.Fatalf("returned non-executable tool %s", tool.Name)
		}
	}
	for _, tool := range SearchTools(ModeAsk, "create a customer") {
		if tool.Scope != "read" {
			t.Fatalf("ask mode returned write tool %s", tool.Name)
		}
	}
}

func TestToolSearchExecutorReturnsCallContracts(t *testing.T) {
	executor := NewAssistantToolExecutor(nil, nil)
	output, handled, err := executor.ExecuteCaretTool(context.Background(), ToolSearchName, map[string]any{"query": "team members"})
	if err != nil || !handled {
		t.Fatalf("handled=%v err=%v", handled, err)
	}
	tools := output["tools"].([]map[string]string)
	if len(tools) == 0 || !strings.Contains(tools[0]["contract"], "[read") {
		t.Fatalf("expected call contracts, got %v", tools)
	}
	if _, _, err := executor.ExecuteCaretTool(context.Background(), ToolSearchName, map[string]any{"query": " "}); err == nil {
		t.Fatal("empty query should be rejected")
	}
}

func TestAssistantPromptAlwaysOffersToolSearch(t *testing.T) {
	prompt := BuildAssistantPrompt("cool who do i have in the team?", "", "", nil, time.Now())
	if !strings.Contains(prompt, "tools.search [read") || !strings.Contains(prompt, "call tools.search") {
		t.Fatal("assistant prompt must always offer tools.search and explain it")
	}
}

func TestSearchToolsFindsGeneralTools(t *testing.T) {
	for query, want := range map[string]string{
		"weather forecast":         "weather.forecast",
		"stock price for a ticker": "markets.quote",
		"convert dollars to euros": "currency.convert",
		"read a link":              "web.fetch",
		"public holidays":          "calendar.public_holidays",
	} {
		found := false
		for _, tool := range SearchTools(ModeAgent, query) {
			if tool.Name == want {
				found = true
			}
		}
		if !found {
			t.Errorf("tools.search %q did not find %s", query, want)
		}
	}
}
