package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTranscriptWriterKeepsReplyOrderAndDropsUnfinishedTools(t *testing.T) {
	recorder := httptest.NewRecorder()
	w := newTranscriptWriter(recorder, time.Now().Add(-1500*time.Millisecond))

	events := []StreamEvent{
		{Type: EventTextDelta, MessageID: "m1", Delta: "I'll search. "},
		{Type: EventTextDelta, MessageID: "m1", Delta: "One moment."},
		{Type: EventToolCall, BlockID: "tool-1", Name: "web.search", Input: map[string]string{"query": "news"}},
		{Type: EventToolResult, BlockID: "tool-1", Status: "success", Summary: "3 results", Output: map[string]string{"answer": "ok"}},
		{Type: EventToolCall, BlockID: "tool-2", Name: "crm.update", RequiresApproval: true},
		{Type: EventTextDelta, MessageID: "m1", Delta: "Here's what I found."},
	}
	for _, event := range events {
		if err := WriteEvent(w, w.Flush, event); err != nil {
			t.Fatalf("write event: %v", err)
		}
	}

	if got := strings.Count(recorder.Body.String(), "data: "); got != len(events) {
		t.Fatalf("stream frames = %d, want %d", got, len(events))
	}

	transcript := w.Transcript()
	if transcript.ResponseMs < 1500 {
		t.Fatalf("response ms = %d, want at least 1500", transcript.ResponseMs)
	}
	got := make([]string, 0, len(transcript.Blocks))
	for _, block := range transcript.Blocks {
		got = append(got, block.Type+":"+block.Content+block.Name)
	}
	want := []string{"text:I'll search. One moment.", "tool_call:web.search", "text:Here's what I found."}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("blocks = %v, want %v", got, want)
	}
	tool := transcript.Blocks[1]
	if tool.Status != "success" || tool.Summary != "3 results" || string(tool.Output) != `{"answer":"ok"}` {
		t.Fatalf("tool block = %+v", tool)
	}
}

func TestTranscriptWriterDropsOversizedToolOutput(t *testing.T) {
	w := newTranscriptWriter(httptest.NewRecorder(), time.Now())
	big := strings.Repeat("x", maxTranscriptToolPayload)
	_ = WriteEvent(w, nil, StreamEvent{Type: EventToolCall, BlockID: "t", Name: "web.search"})
	_ = WriteEvent(w, nil, StreamEvent{Type: EventToolResult, BlockID: "t", Status: "success", Summary: "done", Output: big})

	blocks, _ := json.Marshal(w.Transcript().Blocks)
	if strings.Contains(string(blocks), big) || !strings.Contains(string(blocks), `"summary":"done"`) {
		t.Fatalf("oversized output kept or summary lost: %.200s", blocks)
	}
}
