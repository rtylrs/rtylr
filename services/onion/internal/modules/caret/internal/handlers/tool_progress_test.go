package handlers

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	echoapp "github.com/voxire/onion/internal/modules/caret/internal/application"
)

func TestSlowToolShowsItsCardProgressAndHeartbeats(t *testing.T) {
	previous := toolHeartbeatInterval
	toolHeartbeatInterval = 5 * time.Millisecond
	defer func() { toolHeartbeatInterval = previous }()

	executor := fakeNamedToolExecutorFunc(func(ctx context.Context, name string, _ map[string]any) (map[string]any, bool, error) {
		if name != "web.fetch" {
			return nil, false, nil
		}
		echoapp.ReportToolProgress(ctx, "Reading the pricing page")
		time.Sleep(40 * time.Millisecond)
		return map[string]any{"title": "Pricing", "text": "Pro is $20"}, true, nil
	})
	rec := httptest.NewRecorder()
	var full strings.Builder
	renderer := newStreamEventRenderer(rec, nil, "m1", &full, withStreamMode(ModeAgent),
		withStreamToolExecution(context.Background(), []ToolExecutor{executor}))
	if err := renderer.AppendText(`<tool>{"name":"web.fetch","input":{"url":"https://example.com/pricing"}}</tool>`); err != nil {
		t.Fatal(err)
	}
	if err := renderer.FlushPending(); err != nil {
		t.Fatal(err)
	}
	body := rec.Body.String()
	call := strings.Index(body, `"type":"tool_call"`)
	status := strings.Index(body, "Reading the pricing page")
	result := strings.Index(body, `"type":"tool_result"`)
	if call < 0 || status < 0 || result < 0 || !(call < status && status < result) {
		t.Fatalf("want tool_call, then progress, then tool_result:\n%s", body)
	}
	if !strings.Contains(body, ": ping") {
		t.Fatalf("no heartbeat while the tool ran:\n%s", body)
	}
}
