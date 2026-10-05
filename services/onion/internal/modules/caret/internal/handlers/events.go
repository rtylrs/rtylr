package handlers

import (
	"encoding/json"
	"fmt"
	"io"
)

// StreamEventType discriminates the tagged-union payloads emitted over SSE.
type StreamEventType string

const (
	EventSession       StreamEventType = "session"
	EventMessageStart  StreamEventType = "message_start"
	EventTextDelta     StreamEventType = "text_delta"
	EventThinkingStart StreamEventType = "thinking_start"
	EventThinkingDelta StreamEventType = "thinking_delta"
	EventThinkingEnd   StreamEventType = "thinking_end"
	EventToolCall      StreamEventType = "tool_call"
	EventToolResult    StreamEventType = "tool_result"
	EventMessageEnd    StreamEventType = "message_end"
	EventError         StreamEventType = "error"
	EventDone          StreamEventType = "done"
)

// StreamEvent is the JSON envelope written to the wire after `data: `.
type StreamEvent struct {
	Type             StreamEventType `json:"type"`
	SessionID        string          `json:"session_id,omitempty"`
	MessageID        string          `json:"message_id,omitempty"`
	Role             string          `json:"role,omitempty"`
	BlockID          string          `json:"block_id,omitempty"`
	Delta            string          `json:"delta,omitempty"`
	Name             string          `json:"name,omitempty"`
	Mode             string          `json:"mode,omitempty"`
	Input            any             `json:"input,omitempty"`
	Preview          string          `json:"preview,omitempty"`
	Status           string          `json:"status,omitempty"`
	Output           any             `json:"output,omitempty"`
	Summary          string          `json:"summary,omitempty"`
	Error            string          `json:"error,omitempty"`
	Recoverable      bool            `json:"recoverable,omitempty"`
	DurationMs       int64           `json:"duration_ms,omitempty"`
	RequiresApproval bool            `json:"requires_approval,omitempty"`
}

// WriteEvent serializes the event to the SSE stream and flushes immediately.
func WriteEvent(w io.Writer, flush func(), event StreamEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
		return err
	}
	if flush != nil {
		flush()
	}
	return nil
}

// WriteSSEHeartbeat keeps intermediaries from treating a quiet generation as
// an idle connection. SSE comment frames are ignored by clients but still
// reset proxy idle timers.
func WriteSSEHeartbeat(w io.Writer, flush func()) error {
	if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
		return err
	}
	if flush != nil {
		flush()
	}
	return nil
}
