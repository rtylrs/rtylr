package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// maxTranscriptToolPayload bounds one stored tool input or output. Larger
// payloads keep their summary so a reopened chat still shows what ran.
const maxTranscriptToolPayload = 16 * 1024

// TranscriptBlock is one stored step of an assistant reply, in the order the
// user saw it. Field names match the client's content blocks.
type TranscriptBlock struct {
	ID      string          `json:"id"`
	Type    string          `json:"type"`
	Content string          `json:"content,omitempty"`
	Name    string          `json:"name,omitempty"`
	Mode    string          `json:"mode,omitempty"`
	Input   json.RawMessage `json:"input,omitempty"`
	Status  string          `json:"status,omitempty"`
	Output  json.RawMessage `json:"output,omitempty"`
	Summary string          `json:"summary,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// Transcript is what a reopened chat needs beyond the reply's plain text.
type Transcript struct {
	Blocks     []TranscriptBlock
	ResponseMs int64
}

// transcriptWriter records the stream's text and tool events while passing
// every byte through, so the saved reply matches what was rendered.
type transcriptWriter struct {
	http.ResponseWriter
	startedAt time.Time

	mu     sync.Mutex
	blocks []TranscriptBlock
}

func newTranscriptWriter(w http.ResponseWriter, startedAt time.Time) *transcriptWriter {
	return &transcriptWriter{ResponseWriter: w, startedAt: startedAt}
}

func (t *transcriptWriter) Flush() {
	if flusher, ok := t.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Unwrap lets http.ResponseController reach the underlying writer.
func (t *transcriptWriter) Unwrap() http.ResponseWriter { return t.ResponseWriter }

func (t *transcriptWriter) recordEvent(event StreamEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()

	switch event.Type {
	case EventTextDelta:
		if event.Delta == "" {
			return
		}
		if n := len(t.blocks); n > 0 && t.blocks[n-1].Type == "text" {
			t.blocks[n-1].Content += event.Delta
			return
		}
		t.blocks = append(t.blocks, TranscriptBlock{
			ID:      event.MessageID + "-text-" + strconv.Itoa(len(t.blocks)),
			Type:    "text",
			Content: event.Delta,
		})
	case EventToolCall:
		status := "running"
		if event.RequiresApproval {
			status = "pending"
		}
		t.blocks = append(t.blocks, TranscriptBlock{
			ID:     event.BlockID,
			Type:   "tool_call",
			Name:   event.Name,
			Mode:   event.Mode,
			Input:  boundedJSON(event.Input),
			Status: status,
		})
	case EventToolResult:
		for i := range t.blocks {
			block := &t.blocks[i]
			if block.Type != "tool_call" || block.ID != event.BlockID {
				continue
			}
			block.Status = event.Status
			block.Summary = event.Summary
			block.Error = event.Error
			block.Output = boundedJSON(event.Output)
			return
		}
	}
}

// Transcript returns the finished steps. Tools that never finished in this
// turn (still awaiting approval or running) are dropped: a reopened chat
// cannot act on them.
func (t *transcriptWriter) Transcript() Transcript {
	t.mu.Lock()
	defer t.mu.Unlock()

	blocks := make([]TranscriptBlock, 0, len(t.blocks))
	for _, block := range t.blocks {
		if block.Type == "tool_call" && (block.Status == "pending" || block.Status == "running" || block.Status == "approved") {
			continue
		}
		blocks = append(blocks, block)
	}
	return Transcript{Blocks: blocks, ResponseMs: time.Since(t.startedAt).Milliseconds()}
}

func boundedJSON(value any) json.RawMessage {
	if value == nil {
		return nil
	}
	payload, err := json.Marshal(value)
	if err != nil || len(payload) > maxTranscriptToolPayload || string(payload) == "null" {
		return nil
	}
	return payload
}
