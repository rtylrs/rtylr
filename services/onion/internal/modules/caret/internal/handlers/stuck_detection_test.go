package handlers

import (
	"context"
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	echoapp "github.com/voxire/onion/internal/modules/caret/internal/application"
)

func TestToolThatKeepsFailingIsBlocked(t *testing.T) {
	calls := 0
	executor := fakeNamedToolExecutorFunc(func(_ context.Context, _ string, _ map[string]any) (map[string]any, bool, error) {
		calls++
		return nil, true, errors.New("page unavailable")
	})
	rec := httptest.NewRecorder()
	var full strings.Builder
	renderer := newStreamEventRenderer(rec, nil, "m1", &full, withStreamMode(ModeAgent),
		withStreamToolExecution(context.Background(), []ToolExecutor{executor}))
	for _, page := range []string{"a", "b", "c"} {
		if err := renderer.AppendText(`<tool>{"name":"web.fetch","input":{"url":"https://example.com/` + page + `"}}</tool>`); err != nil {
			t.Fatal(err)
		}
		if err := renderer.FlushPending(); err != nil {
			t.Fatal(err)
		}
	}
	if calls != 2 {
		t.Fatalf("calls = %d, want the third blocked after two failures", calls)
	}
	last := renderer.toolResults[len(renderer.toolResults)-1]
	if last.Status != "blocked" || anyToolRan(renderer.toolResults[2:]) {
		t.Fatalf("third call = %+v", last)
	}
}

func TestMemoryProposalWaitsForTheAnswer(t *testing.T) {
	executor := fakeNamedToolExecutorFunc(func(_ context.Context, _ string, _ map[string]any) (map[string]any, bool, error) {
		return map[string]any{"title": "News"}, true, nil
	})
	rec := httptest.NewRecorder()
	var full strings.Builder
	renderer := newStreamEventRenderer(rec, nil, "m1", &full, withStreamMode(ModeAgent),
		withStreamToolExecution(context.Background(), []ToolExecutor{executor}))
	if err := renderer.AppendText(`<tool>{"name":"web.fetch","input":{"url":"https://example.com/news"}}</tool>` +
		`<tool>{"name":"memory.remember","input":{"key":"interest","value":"Saudi news"}}</tool>`); err != nil {
		t.Fatal(err)
	}
	if err := renderer.FlushPending(); err != nil {
		t.Fatal(err)
	}
	if renderer.PendingApprovalCount() != 0 {
		t.Fatal("a memory proposal alongside running reads must not stop the reply")
	}
}

func TestProgressLinesShareOneStatusBlock(t *testing.T) {
	executor := fakeNamedToolExecutorFunc(func(ctx context.Context, _ string, _ map[string]any) (map[string]any, bool, error) {
		echoapp.ReportToolProgress(ctx, "Step one")
		echoapp.ReportToolProgress(ctx, "Step two")
		return map[string]any{"title": "Done"}, true, nil
	})
	rec := httptest.NewRecorder()
	var full strings.Builder
	renderer := newStreamEventRenderer(rec, nil, "m1", &full, withStreamMode(ModeAgent),
		withStreamToolExecution(context.Background(), []ToolExecutor{executor}))
	if err := renderer.AppendText(`<tool>{"name":"web.fetch","input":{"url":"https://example.com/x"}}</tool>`); err != nil {
		t.Fatal(err)
	}
	if err := renderer.FlushPending(); err != nil {
		t.Fatal(err)
	}
	body := rec.Body.String()
	if starts, ends := strings.Count(body, `"type":"thinking_start"`), strings.Count(body, `"type":"thinking_end"`); starts != 1 || ends != 1 {
		t.Fatalf("thinking starts=%d ends=%d, want one block:\n%s", starts, ends, body)
	}
}
