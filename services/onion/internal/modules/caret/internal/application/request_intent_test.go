package application

import (
	"strings"
	"testing"
)

func TestClassifyCaretRequestSeparatesConversationFromLiveBusinessWork(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		wantKind   CaretIntentKind
		wantLive   bool
		wantAction string
	}{
		{name: "smoke check", request: "test", wantKind: IntentConversation, wantLive: false, wantAction: "acknowledge"},
		{name: "revenue", request: "what was revenue today?", wantKind: IntentAnalytics, wantLive: true, wantAction: "read"},
		{name: "owner brief", request: "what should I focus on next?", wantKind: IntentSnapshot, wantLive: true, wantAction: "recommend"},
		{name: "mutation", request: "create a customer", wantKind: IntentAction, wantLive: false, wantAction: "create"},
		{name: "add product", request: "add a product Protein Bar", wantKind: IntentAction, wantLive: false, wantAction: "create"},
		{name: "memory", request: "remember that weekends are busiest", wantKind: IntentMemory, wantLive: false, wantAction: "remember"},
		{name: "navigation", request: "where do I configure payroll?", wantKind: IntentNavigation, wantLive: false, wantAction: "navigate"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := ClassifyCaretRequest(test.request, "")
			if got.Kind != test.wantKind || got.NeedsLiveData != test.wantLive || got.Action != test.wantAction {
				t.Fatalf("ClassifyCaretRequest(%q) = %#v, want kind=%q live=%v action=%q", test.request, got, test.wantKind, test.wantLive, test.wantAction)
			}
			if test.name == "add product" && got.Domain != "inventory" {
				t.Fatalf("product request domain = %q, want inventory", got.Domain)
			}
		})
	}
}

func TestClassifyCaretRequestUsesPageContextForEntityScope(t *testing.T) {
	got := ClassifyCaretRequest(
		"update this product",
		`Structured page context: {"selected":{"product_id":"prod_123"},"route":"/products"}`,
	)
	if !got.HasEntityContext {
		t.Fatalf("expected selected entity context: %#v", got)
	}
	if got.NeedsClarification {
		t.Fatalf("selected product should not require clarification: %#v", got)
	}
	if got.Scope != "selected entity and current page scope" {
		t.Fatalf("scope = %q", got.Scope)
	}
}

func TestClassifyCaretRequestFlagsAmbiguousEntityAction(t *testing.T) {
	got := ClassifyCaretRequest("update this product", "Route: /menus")
	if !got.NeedsClarification {
		t.Fatalf("expected clarification for unresolved entity: %#v", got)
	}
	if !got.NeedsApproval || !got.NeedsTool {
		t.Fatalf("expected tool and approval policy for action: %#v", got)
	}
}

func TestRequestPlanPromptTextExplainsEvidenceStrategy(t *testing.T) {
	plan := ClassifyCaretRequest("show my top customers this month", "Route: /crm")
	text := plan.PromptText()
	for _, want := range []string{"Intent: analytics", "Domain: customer", "Freshness: this month", "Live business evidence is required", "Use the server capability catalog"} {
		if !containsIntentText(text, want) {
			t.Fatalf("plan text missing %q: %s", want, text)
		}
	}
}

func containsIntentText(text, want string) bool {
	return strings.Contains(text, want)
}

func TestClassifyCaretRequestDoesNotTreatBusinessNounsAsMutations(t *testing.T) {
	for _, request := range []string{"what is a refund?", "show me refund policy", "how do refunds work?"} {
		plan := ClassifyCaretRequest(request, "")
		if plan.NeedsApproval || plan.Kind == IntentAction {
			t.Fatalf("request %q was treated as a mutation: %+v", request, plan)
		}
	}
	for _, request := range []string{"refund order 123", "issue refund for order 123", "process refund"} {
		plan := ClassifyCaretRequest(request, "")
		if !plan.NeedsApproval || plan.Kind != IntentAction {
			t.Fatalf("request %q was not treated as an approval-backed action: %+v", request, plan)
		}
	}
}
