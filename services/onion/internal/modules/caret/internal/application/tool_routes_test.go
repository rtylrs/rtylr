package application

import (
	"strings"
	"testing"
	"time"
)

func TestLearnedRoutesNeedASearchThenASuccess(t *testing.T) {
	calls := []RouteTarget{
		{Name: ToolSearchName, Success: true, Output: map[string]any{"tools": []map[string]string{{"name": "hr.list_leave_requests"}, {"name": "hr.get_leave_balances"}}}},
		{Name: "hr.get_leave_balances", Success: false},
		{Name: "hr.list_leave_requests", Success: true},
		{Name: "crm.list_customers", Success: true},
	}
	if got := LearnedRoutes(calls); len(got) != 1 || got[0] != "hr.list_leave_requests" {
		t.Fatalf("routes = %v", got)
	}
}

func TestRoutedToolsOfferWhatServedSimilarRequests(t *testing.T) {
	routes := ToolRoutes{"hr.list_leave_requests": {RouteWords("Who asked for time off next week?")}}
	if got := RoutedToolsForRequest("who asked for time off this week", routes); len(got) != 1 || got[0].Name != "hr.list_leave_requests" {
		t.Fatalf("routed = %#v", got)
	}
	if got := RoutedToolsForRequest("what is the weather", routes); len(got) != 0 {
		t.Fatalf("unrelated request routed: %#v", got)
	}
	prompt := BuildAssistantPromptFor(AssistantPromptInput{Request: "anyone asked for time off?", Now: time.Now(), Routes: routes})
	if !strings.Contains(prompt, "\n- hr.list_leave_requests [") {
		t.Fatal("a learned route was not offered in the prompt")
	}
}

func TestRouteWordsDropsNoise(t *testing.T) {
	if got := strings.Join(RouteWords("Show me ALL the unpaid invoices, please!"), " "); got != "unpaid invoices" {
		t.Fatalf("words = %q", got)
	}
}
