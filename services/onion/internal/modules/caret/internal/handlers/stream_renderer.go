package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
	"unicode/utf8"

	echoapp "github.com/voxire/onion/internal/modules/caret/internal/application"
	"github.com/voxire/onion/internal/platform/logger"
)

const (
	thinkingOpenTag  = "<thinking>"
	thinkingCloseTag = "</thinking>"
	toolOpenTag      = "<tool>"
	toolCloseTag     = "</tool>"
	followUpsOpen    = "<followups>"
	followUpsClose   = "</followups>"
	maxFollowUps     = 3
	maxFollowUpRunes = 120
	markerHoldBytes  = answerQualityHoldBytes
)

type markerKind int

const (
	markerNone markerKind = iota
	markerThinking
	markerTool
	markerFollowUps
)

type streamEventRenderer struct {
	w                io.Writer
	flush            func()
	messageID        string
	fullResponse     *strings.Builder
	pending          string
	ctx              context.Context
	executors        []ToolExecutor
	toolResults      []streamToolExecution
	mode             CaretMode
	toolState        *toolCallExecutionState
	plannerOnly      bool
	pageContext      map[string]any
	pendingApprovals int
	// toolsDisabled drops tool markers, for the final wrap-up round that must
	// answer from the results it already has.
	toolsDisabled bool
	// toolTokens are model tokens tools spent on their own calls, metered
	// with the reply.
	toolTokens int64
	// inflight reads started during this turn; progress carries their
	// status lines.
	inflight []*inflightTool
	// storedID is the saved reply's message id, sent with message_end.
	storedID string
	progress chan string
	// statusBlock is the open thinking block progress lines go into, so one
	// wait shows as one status rather than a stack of them.
	statusBlock string
	// failures counts errors per tool in this reply; a tool that keeps
	// failing is blocked so the reply moves on instead of spinning.
	failures map[string]int
	// followUps are the next questions the model suggested, sent as their
	// own event when the reply finishes rather than shown as text.
	followUps []string
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

func withStreamPlannerOnly(enabled bool) streamRendererOption {
	return func(r *streamEventRenderer) {
		r.plannerOnly = enabled
	}
}

func withStreamPageContext(pageContext map[string]any) streamRendererOption {
	return func(r *streamEventRenderer) {
		r.pageContext = pageContext
	}
}

func withStreamToolBudget(limit int) streamRendererOption {
	return func(r *streamEventRenderer) {
		if limit > 0 && r.toolState != nil {
			if limit > echoapp.MaxToolCallsPerTurn {
				limit = echoapp.MaxToolCallsPerTurn
			}
			r.toolState.SetLimit(limit)
		}
	}
}

func newStreamEventRenderer(w io.Writer, flush func(), messageID string, fullResponse *strings.Builder, opts ...streamRendererOption) *streamEventRenderer {
	r := &streamEventRenderer{
		w:            w,
		flush:        flush,
		messageID:    messageID,
		fullResponse: fullResponse,
		mode:         ModeAsk,
		toolState:    echoapp.NewToolCallExecutionState(),
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

// AppendServerText writes application-generated text. Provider prose is
// suppressed for planner-only turns, but the server may still need to render
// a bounded fallback after a tool result or execution failure.
func (r *streamEventRenderer) AppendServerText(text string) error {
	return r.emitPublicText(text)
}

func (r *streamEventRenderer) Finish() error {
	if err := r.FlushPending(); err != nil {
		return err
	}
	if len(r.followUps) > 0 && !r.plannerOnly {
		if err := WriteEvent(r.w, r.flush, StreamEvent{Type: EventFollowUps, MessageID: r.messageID, Items: r.followUps}); err != nil {
			return err
		}
	}
	if err := WriteEvent(r.w, r.flush, StreamEvent{Type: EventMessageEnd, MessageID: r.messageID, StoredID: r.storedID}); err != nil {
		return err
	}
	return WriteEvent(r.w, r.flush, StreamEvent{Type: EventDone})
}

// FlushPending ends the model's turn: it renders what is buffered and
// collects the reads still running.
func (r *streamEventRenderer) FlushPending() error {
	if err := r.drain(true); err != nil {
		return err
	}
	return r.awaitTools()
}

func (r *streamEventRenderer) ToolResultCount() int {
	return len(r.toolResults)
}

func (r *streamEventRenderer) HasSuccessfulTool(name string) bool {
	for _, result := range r.toolResults {
		if result.Name == name && result.Status == "success" {
			return true
		}
	}
	return false
}

func (r *streamEventRenderer) PendingApprovalCount() int {
	return r.pendingApprovals
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
		case markerFollowUps:
			handled = r.drainFollowUps(final)
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
	emitLen = safePublicAnswerPrefix(r.pending, emitLen)
	// The hold-back cut is in bytes; never split a multi-byte character, or
	// it reaches the client as replacement characters (e.g. an em dash as ���).
	for emitLen > 0 && emitLen < len(r.pending) && !utf8.RuneStart(r.pending[emitLen]) {
		emitLen--
	}
	if emitLen <= 0 {
		return nil
	}
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
			if !r.plannerOnly {
				if err := r.emitText(r.pending); err != nil {
					return false, err
				}
			}
			r.pending = ""
			return true, nil
		}
		return false, nil
	}
	if r.plannerOnly {
		r.pending = r.pending[end+len(thinkingCloseTag):]
		return true, nil
	}

	content := strings.TrimSpace(r.pending[len(thinkingOpenTag):end])
	blockID := newCaretBlockID("thinking")
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
	if r.plannerOnly {
		return nil
	}
	return r.emitPublicText(text)
}

func (r *streamEventRenderer) emitPublicText(text string) error {
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
	if r.toolsDisabled {
		return nil
	}

	def, ok := echoapp.ToolByName(call.Name)
	if !ok {
		return nil
	}
	if !echoapp.ToolAllowedForMode(r.mode, def) {
		return nil
	}
	call.Input = echoapp.NormalizeToolInput(call.Name, call.Input)
	call.Input = echoapp.BindToolInputToPageScope(call.Name, call.Input, r.pageContext)
	call.Input = echoapp.NormalizeToolInput(call.Name, call.Input)
	if err := echoapp.ValidateToolInvocation(r.mode, def, call.Input); err != nil {
		return r.emitToolExecutionNotice(call, "error", "I could not run that business check because the requested inputs were invalid.", err.Error())
	}

	requiresApproval := def.RequiresApproval || def.Scope == "write"
	if requiresApproval && call.Name == "memory.remember" && len(r.inflight) > 0 {
		// A memory proposal stops the reply until it is approved, so one made
		// while lookups are still running would leave the person with no
		// answer. It belongs at the end, after the answer.
		return r.emitToolExecutionNotice(call, "skipped", "Answer first; propose a memory only at the end of the reply.", "")
	}
	if !requiresApproval && len(r.executors) == 0 {
		return nil
	}
	preview := buildToolActionPreview(def, call.Input)

	blockID := newCaretBlockID("tool")
	if !requiresApproval {
		return r.emitReadToolMarker(blockID, call, preview)
	}
	if err := echoapp.RegisterPendingToolApproval(r.ctx, blockID, def, call.Input, r.mode); err != nil {
		return r.emitToolExecutionNotice(call, "error", "I could not create a safe approval request for that action. No changes were made.", "approval request registry is full")
	}
	r.pendingApprovals++

	return WriteEvent(r.w, r.flush, StreamEvent{
		Type:             EventToolCall,
		MessageID:        r.messageID,
		BlockID:          blockID,
		Name:             call.Name,
		Mode:             string(r.mode),
		Input:            call.Input,
		Preview:          preview,
		RequiresApproval: requiresApproval,
	})
}

func (r *streamEventRenderer) emitReadToolMarker(blockID string, call ToolCallMarker, preview string) error {
	ctx := r.ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if r.failures[call.Name] >= maxToolFailures {
		return r.emitToolExecutionNotice(call, "blocked", call.Name+" already failed twice in this reply; use another route or say what could not be checked.", "")
	}
	if err := r.toolState.Admit(call.Name, call.Input); err != nil {
		if errors.Is(err, echoapp.ErrDuplicateToolCall) {
			return r.emitToolExecutionNotice(call, "skipped", "I already checked that exact request; use the existing result.", "")
		}
		return r.emitToolExecutionNotice(call, "blocked", "I stopped additional checks for this response so the answer stays focused.", "")
	}

	callEvent := StreamEvent{
		Type:             EventToolCall,
		MessageID:        r.messageID,
		BlockID:          blockID,
		Name:             call.Name,
		Mode:             string(r.mode),
		Input:            call.Input,
		Preview:          preview,
		RequiresApproval: false,
	}
	// A slow tool shows its card as soon as it starts, so the person sees work
	// happening; a quick one appears with its result.
	announced := echoapp.ToolTimeout(call.Name) > echoapp.ToolExecutionTimeout
	if announced {
		if err := WriteEvent(r.w, r.flush, callEvent); err != nil {
			return err
		}
	}
	r.startTool(ctx, &inflightTool{blockID: blockID, call: call, event: callEvent, announced: announced})
	return nil
}

// inflightTool is a read that started when its marker closed and is
// collected when the model's turn ends, so several reads in one turn run at
// the same time and overlap with the rest of the model's output.
type inflightTool struct {
	blockID   string
	call      ToolCallMarker
	event     StreamEvent
	announced bool
	done      chan toolOutcome
}

type toolOutcome struct {
	output   map[string]any
	executed bool
	duration time.Duration
	err      error
}

func (r *streamEventRenderer) startTool(ctx context.Context, tool *inflightTool) {
	if r.progress == nil {
		r.progress = make(chan string, 64)
	}
	progress := r.progress
	ctx = echoapp.WithToolProgress(ctx, func(message string) {
		select {
		case progress <- message:
		default:
		}
	})
	tool.done = make(chan toolOutcome, 1)
	go func() {
		output, executed, duration, err := echoapp.ExecuteToolWithTimeout(ctx, r.executors, tool.call.Name, tool.call.Input)
		tool.done <- toolOutcome{output, executed, duration, err}
	}()
	r.inflight = append(r.inflight, tool)
}

// awaitTools collects every running read in the order the model called them,
// keeping the stream alive meanwhile: heartbeats stop proxies from closing a
// quiet connection, and progress the tools report is shown as status lines.
// Only this goroutine writes to the stream.
func (r *streamEventRenderer) awaitTools() (err error) {
	heartbeat := time.NewTicker(toolHeartbeatInterval)
	defer heartbeat.Stop()
	defer func() {
		if closeErr := r.closeStatus(); err == nil {
			err = closeErr
		}
	}()
	for len(r.inflight) > 0 {
		tool := r.inflight[0]
		var outcome toolOutcome
	wait:
		for {
			select {
			case outcome = <-tool.done:
				break wait
			case message := <-r.progress:
				if err := r.emitStatus(message); err != nil {
					return err
				}
			case <-heartbeat.C:
				if err := WriteSSEHeartbeat(r.w, r.flush); err != nil {
					return err
				}
			}
		}
		r.inflight = r.inflight[1:]
		if err := r.finishTool(tool, outcome); err != nil {
			return err
		}
	}
	for {
		select {
		case message := <-r.progress:
			if err := r.emitStatus(message); err != nil {
				return err
			}
		default:
			return nil
		}
	}
}

func (r *streamEventRenderer) finishTool(tool *inflightTool, outcome toolOutcome) error {
	call, blockID := tool.call, tool.blockID
	output, executed, duration, execErr := outcome.output, outcome.executed, outcome.duration, outcome.err
	logToolCall(r.ctx, call.Name, executed, duration, execErr)
	if execErr == nil && !executed {
		if !tool.announced {
			// No executor claimed the tool; keep unsupported markers invisible rather
			// than suggesting that an action ran.
			return nil
		}
		execErr = fmt.Errorf("tool %q is not available", call.Name)
	}
	if !tool.announced {
		if err := WriteEvent(r.w, r.flush, tool.event); err != nil {
			return err
		}
	}
	if execErr != nil {
		if r.failures == nil {
			r.failures = map[string]int{}
		}
		r.failures[call.Name]++
		userError := userFacingToolError(execErr)
		summary := summarizeToolResult(call.Name, call.Input, nil, "error", userError)
		r.toolResults = append(r.toolResults, streamToolExecution{Name: call.Name, Input: call.Input, Status: "error", Summary: summary, Error: userError})
		return WriteEvent(r.w, r.flush, StreamEvent{Type: EventToolResult, MessageID: r.messageID, BlockID: blockID, Status: "error", Summary: summary, Error: userError, DurationMs: duration.Milliseconds()})
	}
	switch tokens := output[echoapp.ToolUsageTokensKey].(type) {
	case int64:
		r.toolTokens += tokens
	case float64: // a shortened output went through JSON
		r.toolTokens += int64(tokens)
	}
	delete(output, echoapp.ToolUsageTokensKey)
	if call.Name == "web.search" {
		if answer, ok := output["answer"].(string); ok {
			output["answer"] = removeCaretEmDashes(answer)
		}
	}

	summary := summarizeToolResult(call.Name, call.Input, output, "success", "")
	r.toolResults = append(r.toolResults, streamToolExecution{Name: call.Name, Input: call.Input, Status: "success", Summary: summary, Output: output})
	return WriteEvent(r.w, r.flush, StreamEvent{Type: EventToolResult, MessageID: r.messageID, BlockID: blockID, Status: "success", Summary: summary, Output: output, DurationMs: duration.Milliseconds()})
}

// toolHeartbeatInterval keeps a connection alive while a tool runs.
var toolHeartbeatInterval = 10 * time.Second

// maxToolFailures is how many errors one tool may return in a reply before
// further calls to it are blocked.
const maxToolFailures = 2

// emitStatus adds a progress line to the open status block, starting one if
// needed; closeStatus ends it once the tools it describes are done.
func (r *streamEventRenderer) emitStatus(message string) error {
	message = strings.TrimSpace(message)
	if message == "" || r.plannerOnly {
		return nil
	}
	delta := message
	if r.statusBlock == "" {
		r.statusBlock = newCaretBlockID("thinking")
		if err := WriteEvent(r.w, r.flush, StreamEvent{Type: EventThinkingStart, MessageID: r.messageID, BlockID: r.statusBlock}); err != nil {
			return err
		}
	} else {
		delta = "\n" + message
	}
	return WriteEvent(r.w, r.flush, StreamEvent{Type: EventThinkingDelta, MessageID: r.messageID, BlockID: r.statusBlock, Delta: delta})
}

func (r *streamEventRenderer) closeStatus() error {
	if r.statusBlock == "" {
		return nil
	}
	blockID := r.statusBlock
	r.statusBlock = ""
	return WriteEvent(r.w, r.flush, StreamEvent{Type: EventThinkingEnd, MessageID: r.messageID, BlockID: blockID})
}

func (r *streamEventRenderer) emitToolExecutionNotice(call ToolCallMarker, status, summary, errorMessage string) error {
	blockID := newCaretBlockID("tool")
	r.toolResults = append(r.toolResults, streamToolExecution{Name: call.Name, Input: call.Input, Status: status, Summary: summary, Error: errorMessage})
	return WriteEvent(r.w, r.flush, StreamEvent{Type: EventToolResult, MessageID: r.messageID, BlockID: blockID, Name: call.Name, Status: status, Summary: summary, Error: errorMessage})
}

func firstMarker(s string) (int, markerKind) {
	best, kind := -1, markerNone
	for _, candidate := range []struct {
		tag  string
		kind markerKind
	}{{thinkingOpenTag, markerThinking}, {toolOpenTag, markerTool}, {followUpsOpen, markerFollowUps}} {
		if index := strings.Index(s, candidate.tag); index >= 0 && (best < 0 || index < best) {
			best, kind = index, candidate.kind
		}
	}
	return best, kind
}

func newCaretBlockID(prefix string) string {
	return fmt.Sprintf("%s-%d-%d", prefix, time.Now().UnixNano(), caretIDCounter.Add(1))
}

// drainFollowUps takes a <followups>["...", ...]</followups> marker out of
// the reply. Malformed or unfinished markers are dropped, never shown.
func (r *streamEventRenderer) drainFollowUps(final bool) bool {
	end := strings.Index(r.pending, followUpsClose)
	if end < 0 {
		if final {
			r.pending = ""
			return true
		}
		return false
	}
	raw := strings.TrimSpace(r.pending[len(followUpsOpen):end])
	r.pending = r.pending[end+len(followUpsClose):]
	var items []string
	if json.Unmarshal([]byte(raw), &items) != nil {
		return true
	}
	seen := map[string]bool{}
	followUps := make([]string, 0, maxFollowUps)
	for _, item := range items {
		item = strings.Join(strings.Fields(item), " ")
		key := strings.ToLower(item)
		if item == "" || seen[key] || utf8.RuneCountInString(item) > maxFollowUpRunes {
			continue
		}
		seen[key] = true
		followUps = append(followUps, item)
		if len(followUps) == maxFollowUps {
			break
		}
	}
	r.followUps = followUps
	return true
}

// logToolCall records one tool call for operations: which tool, how long,
// and whether it failed. Inputs and outputs stay out of the logs.
func logToolCall(ctx context.Context, name string, executed bool, duration time.Duration, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	fields := []logger.Field{logger.String("tool", name), logger.Int("duration_ms", int(duration.Milliseconds())), logger.Bool("executed", executed)}
	if err != nil {
		logger.Ctx(ctx).Warn("Caret tool failed", append(fields, logger.String("error", userFacingToolError(err)))...)
		return
	}
	logger.Ctx(ctx).Info("Caret tool ran", fields...)
}
