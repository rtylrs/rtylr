package handlers

import (
	"context"
	"fmt"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/caret/internal/application"
)

type portalExecutor struct {
	mu    sync.Mutex
	calls []string
}

func (e *portalExecutor) ExecuteCaretTool(_ context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if !strings.HasPrefix(name, "portal.") {
		return nil, false, nil
	}
	// The first change is the slowest, so running them at once would land
	// them out of order.
	if id, _ := input["id"].(string); id == "home" {
		time.Sleep(30 * time.Millisecond)
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.calls = append(e.calls, name+":"+fmt.Sprint(input["id"]))
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
	if len(executor.calls) != 1 || executor.calls[0] != "portal.customize:caret" {
		t.Fatalf("calls = %v", executor.calls)
	}
	if renderer.PendingApprovalCount() != 0 || !renderer.HasSuccessfulTool("portal.customize") {
		t.Fatalf("pending approvals = %d, succeeded = %v", renderer.PendingApprovalCount(), renderer.HasSuccessfulTool("portal.customize"))
	}
	if strings.Contains(rec.Body.String(), `"requires_approval":true`) {
		t.Fatalf("the change asked for approval:\n%s", rec.Body.String())
	}
}

// Two changes in one reply apply in the order the model wrote them, so "put
// Home first, then Work under it" can't end with Work on top.
func TestPortalChangesApplyInOrder(t *testing.T) {
	rec := httptest.NewRecorder()
	executor := &portalExecutor{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", &strings.Builder{},
		withStreamMode(application.ModeAgent),
		withStreamToolExecution(context.Background(), []application.ToolExecutor{executor}))
	text := `<tool>{"name":"portal.customize","input":{"area":"rail","id":"home","position":"first"}}</tool>` +
		`<tool>{"name":"portal.customize","input":{"area":"rail","id":"flow","position":"after:home"}}</tool>`
	if err := renderer.AppendText(text); err != nil {
		t.Fatal(err)
	}
	if err := renderer.Finish(); err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(executor.calls, ","); got != "portal.customize:home,portal.customize:flow" {
		t.Fatalf("calls ran as %s", got)
	}
}

// DeepSeek sometimes writes its own function-call markup instead of a <tool>
// marker. The call still runs and the markup never reaches the person.
func TestNativeToolCallMarkupRunsTheTool(t *testing.T) {
	rec := httptest.NewRecorder()
	executor := &portalExecutor{}
	answer := &strings.Builder{}
	renderer := newStreamEventRenderer(rec, nil, "assistant-1", answer,
		withStreamMode(application.ModeAgent),
		withStreamToolExecution(context.Background(), []application.ToolExecutor{executor}))
	markup := "Moving it now.\n\n<｜DSML｜function_calls>\n<｜DSML｜invoke name=\"portal.customize\">\n" +
		"<｜DSML｜parameter name=\"area\" string=\"true\">rail</｜DSML｜parameter>\n" +
		"<｜DSML｜parameter name=\"id\" string=\"true\">profile</｜DSML｜parameter>\n" +
		"<｜DSML｜parameter name=\"position\" string=\"true\">first</｜DSML｜parameter>\n" +
		"</｜DSML｜invoke>\n</｜DSML｜function_calls>"
	// Streamed in small pieces, as the provider sends it.
	for len(markup) > 0 {
		n := min(7, len(markup))
		if err := renderer.AppendText(markup[:n]); err != nil {
			t.Fatal(err)
		}
		markup = markup[n:]
	}
	if err := renderer.Finish(); err != nil {
		t.Fatal(err)
	}
	if len(executor.calls) != 1 || executor.calls[0] != "portal.customize:profile" {
		t.Fatalf("calls = %v", executor.calls)
	}
	if strings.Contains(rec.Body.String(), "DSML") || strings.Contains(answer.String(), "DSML") {
		t.Fatalf("markup leaked:\n%s", rec.Body.String())
	}
	if !strings.Contains(answer.String(), "Moving it now.") {
		t.Fatalf("answer = %q", answer.String())
	}
}
