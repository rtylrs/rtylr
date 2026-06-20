package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"
)

const (
	thinkingOpenTag  = "<thinking>"
	thinkingCloseTag = "</thinking>"
	toolOpenTag      = "<tool>"
	toolCloseTag     = "</tool>"
	markerHoldBytes  = answerQualityHoldBytes
)

type markerKind int

const (
	markerNone markerKind = iota
	markerThinking
	markerTool
)

type streamEventRenderer struct {
	w            io.Writer
	flush        func()
	messageID    string
	fullResponse *strings.Builder
	pending      string
	ctx          context.Context
	executors    []ToolExecutor
	toolResults  []streamToolExecution
	mode         CaretMode
}

type streamRendererOption func(*streamEventRenderer)

type streamToolExecution struct {
	Name    string         `json:"name"`
	Input   map[string]any `json:"input,omitempty"`
	Status  string         `json:"status"`
	Summary string         `json:"summary,omitempty"`
	Output  map[string]any `json:"output,omitempty"`
	Error   string         `json:"error,omitempty"`
}

func withStreamToolExecution(ctx context.Context, executors []ToolExecutor) streamRendererOption {
	return func(r *streamEventRenderer) {
		r.ctx = ctx
		r.executors = executors
	}
}

func withStreamMode(mode CaretMode) streamRendererOption {
	return func(r *streamEventRenderer) {
		r.mode = mode
	}
}

func newStreamEventRenderer(w io.Writer, flush func(), messageID string, fullResponse *strings.Builder, opts ...streamRendererOption) *streamEventRenderer {
	r := &streamEventRenderer{
		w:            w,
		flush:        flush,
		messageID:    messageID,
		fullResponse: fullResponse,
		mode:         ModeAsk,
	}
	for _, opt := range opts {
		opt(r)
	}
	if r.mode == "" {
		r.mode = ModeAsk
	}
	return r
}

func (r *streamEventRenderer) AppendText(text string) error {
	if text == "" {
		return nil
	}
	r.pending += text
	return r.drain(false)
}

func (r *streamEventRenderer) Finish() error {
	if err := r.drain(true); err != nil {
		return err
	}
	if err := WriteEvent(r.w, r.flush, StreamEvent{Type: EventMessageEnd, MessageID: r.messageID}); err != nil {
		return err
	}
	return WriteEvent(r.w, r.flush, StreamEvent{Type: EventDone})
}

func (r *streamEventRenderer) FlushPending() error {
	return r.drain(true)
}

func (r *streamEventRenderer) ToolResultCount() int {
	return len(r.toolResults)
}

func (r *streamEventRenderer) ToolResultsSince(index int) []streamToolExecution {
	if index < 0 {
		index = 0
	}
	if index >= len(r.toolResults) {
		return nil
	}
	out := make([]streamToolExecution, len(r.toolResults[index:]))
	copy(out, r.toolResults[index:])
	return out
}

func (r *streamEventRenderer) drain(final bool) error {
	for {
		idx, kind := firstMarker(r.pending)
		if idx < 0 {
			return r.emitPlainRemainder(final)
		}
		if idx > 0 {
			if err := r.emitText(r.pending[:idx]); err != nil {
				return err
			}
			r.pending = r.pending[idx:]
			continue
		}

		var handled bool
		var err error
		switch kind {
		case markerThinking:
			handled, err = r.drainThinking(final)
		case markerTool:
			handled, err = r.drainTool(final)
		default:
			handled = true
		}
		if err != nil {
			return err
		}
		if !handled {
			return nil
		}
	}
}

func (r *streamEventRenderer) emitPlainRemainder(final bool) error {
	if r.pending == "" {
		return nil
	}
	if final {
		if err := r.emitText(r.pending); err != nil {
			return err
		}
		r.pending = ""
		return nil
	}
	if len(r.pending) <= markerHoldBytes {
		return nil
	}
	emitLen := len(r.pending) - markerHoldBytes
	if err := r.emitText(r.pending[:emitLen]); err != nil {
		return err
	}
	r.pending = r.pending[emitLen:]
	return nil
}

func (r *streamEventRenderer) drainThinking(final bool) (bool, error) {
	end := strings.Index(r.pending, thinkingCloseTag)
	if end < 0 {
		if final {
			if err := r.emitText(r.pending); err != nil {
				return false, err
			}
			r.pending = ""
			return true, nil
		}
		return false, nil
	}

	content := strings.TrimSpace(r.pending[len(thinkingOpenTag):end])
	blockID := fmt.Sprintf("thinking-%d", time.Now().UnixNano())
	if err := WriteEvent(r.w, r.flush, StreamEvent{Type: EventThinkingStart, MessageID: r.messageID, BlockID: blockID}); err != nil {
		return false, err
	}
	if content != "" {
		if err := WriteEvent(r.w, r.flush, StreamEvent{Type: EventThinkingDelta, MessageID: r.messageID, BlockID: blockID, Delta: content}); err != nil {
			return false, err
		}
	}
	if err := WriteEvent(r.w, r.flush, StreamEvent{Type: EventThinkingEnd, MessageID: r.messageID, BlockID: blockID}); err != nil {
		return false, err
	}

	r.pending = r.pending[end+len(thinkingCloseTag):]
	return true, nil
}

func (r *streamEventRenderer) drainTool(final bool) (bool, error) {
	end := strings.Index(r.pending, toolCloseTag)
	if end < 0 {
		if final {
			r.pending = ""
			return true, nil
		}
		return false, nil
	}

	raw := strings.TrimSpace(r.pending[len(toolOpenTag):end])
	if err := r.emitToolMarker(raw); err != nil {
		return false, err
	}
	r.pending = r.pending[end+len(toolCloseTag):]
	return true, nil
}

func (r *streamEventRenderer) emitText(text string) error {
	if text == "" {
		return nil
	}
	text = polishCaretAnswerText(text)
	if text == "" {
		return nil
	}
	if err := WriteEvent(r.w, r.flush, StreamEvent{
		Type:      EventTextDelta,
		MessageID: r.messageID,
		Delta:     text,
	}); err != nil {
		return err
	}
	if r.fullResponse != nil {
		r.fullResponse.WriteString(text)
	}
	return nil
}

func (r *streamEventRenderer) emitToolMarker(raw string) error {
	var call ToolCallMarker
	if err := json.Unmarshal([]byte(raw), &call); err != nil || strings.TrimSpace(call.Name) == "" {
		return nil
	}

	def, ok := ToolByName(call.Name)
	if !ok {
		return nil
	}
	if !toolAllowedForMode(r.mode, def) {
		return nil
	}

	requiresApproval := def.RequiresApproval || def.Scope == "write"
	if !requiresApproval && len(r.executors) == 0 {
		return nil
	}
	preview := buildToolActionPreview(def, call.Input)

	blockID := fmt.Sprintf("tool-%d", time.Now().UnixNano())
	if err := WriteEvent(r.w, r.flush, StreamEvent{
		Type:             EventToolCall,
		MessageID:        r.messageID,
		BlockID:          blockID,
		Name:             call.Name,
		Mode:             string(r.mode),
		Input:            call.Input,
		Preview:          preview,
		RequiresApproval: requiresApproval,
	}); err != nil {
		return err
	}

	if requiresApproval {
		return nil
	}

	ctx := r.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	for _, executor := range r.executors {
		if executor == nil {
			continue
		}
		output, executed, err := executor.ExecuteCaretTool(ctx, call.Name, call.Input)
		if err != nil {
			userError := userFacingToolError(err)
			summary := summarizeToolResult(call.Name, call.Input, nil, "error", userError)
			r.toolResults = append(r.toolResults, streamToolExecution{
				Name:    call.Name,
				Input:   call.Input,
				Status:  "error",
				Summary: summary,
				Error:   userError,
			})
			return WriteEvent(r.w, r.flush, StreamEvent{
				Type:      EventToolResult,
				MessageID: r.messageID,
				BlockID:   blockID,
				Status:    "error",
				Summary:   summary,
				Error:     userError,
			})
		}
		if executed {
			summary := summarizeToolResult(call.Name, call.Input, output, "success", "")
			r.toolResults = append(r.toolResults, streamToolExecution{
				Name:    call.Name,
				Input:   call.Input,
				Status:  "success",
				Summary: summary,
				Output:  output,
			})
			return WriteEvent(r.w, r.flush, StreamEvent{
				Type:      EventToolResult,
				MessageID: r.messageID,
				BlockID:   blockID,
				Status:    "success",
				Summary:   summary,
				Output:    output,
			})
		}
	}

	summary := summarizeToolResult(call.Name, call.Input, nil, "error", "This tool is registered, but execution is not available in this stream yet.")
	r.toolResults = append(r.toolResults, streamToolExecution{
		Name:    call.Name,
		Input:   call.Input,
		Status:  "error",
		Summary: summary,
		Error:   "This tool is registered, but execution is not available in this stream yet.",
	})
	return WriteEvent(r.w, r.flush, StreamEvent{
		Type:      EventToolResult,
		MessageID: r.messageID,
		BlockID:   blockID,
		Status:    "error",
		Summary:   summary,
		Error:     "This tool is registered, but execution is not available in this stream yet.",
	})
}

func firstMarker(s string) (int, markerKind) {
	thinking := strings.Index(s, thinkingOpenTag)
	tool := strings.Index(s, toolOpenTag)
	switch {
	case thinking < 0 && tool < 0:
		return -1, markerNone
	case tool < 0 || (thinking >= 0 && thinking < tool):
		return thinking, markerThinking
	default:
		return tool, markerTool
	}
}
