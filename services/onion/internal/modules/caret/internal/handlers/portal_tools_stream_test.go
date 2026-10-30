package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
)

type portalExecutor struct{ calls []string }

func (e *portalExecutor) ExecuteCaretTool(_ context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if !strings.HasPrefix(name, "portal.") {
		return nil, false, nil
	}
	e.calls = append(e.calls, name)
	return map[string]any{"area": input["area"]}, true, nil
}

// "Move yourself to the top" rearranges only the person's own Portal, so it
// runs at once instead of waiting on an approval card.
func TestPortalLayoutChangeRunsWithoutApproval(t *testing.T) {
	rec := httptest.NewRecorder()
	executor := &portalExecutor{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", &strings.Builder{},
		withStreamMode(application.ModeAgent),
		withStreamToolExecution(context.Background(), []application.ToolExecutor{executor}))

	if err := renderer.AppendText(`<tool>{"name":"portal.customize","input":{"area":"rail","id":"caret","group":"top","position":"first"}}</tool>`); err != nil {
		t.Fatal(err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatal(err)
	}
	if len(executor.calls) != 1 || executor.calls[0] != "portal.customize" {
		t.Fatalf("calls = %v", executor.calls)
	}
	if renderer.PendingApprovalCount() != 0 || !renderer.HasSuccessfulTool("portal.customize") {
		t.Fatalf("pending approvals = %d, succeeded = %v", renderer.PendingApprovalCount(), renderer.HasSuccessfulTool("portal.customize"))
	}
	if strings.Contains(rec.Body.String(), `"requires_approval":true`) {
		t.Fatalf("the change asked for approval:\n%s", rec.Body.String())
	}
}
