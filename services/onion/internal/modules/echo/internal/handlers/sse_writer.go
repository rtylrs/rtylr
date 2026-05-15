package handlers

import (
	"net/http"
)

// sseWriter binds together the ResponseWriter and Flusher so handler code
// can emit events without touching the http.Flusher cast every line.
type sseWriter struct {
	w       http.ResponseWriter
	flush   func()
	enabled bool
}

func newSseWriter(w http.ResponseWriter) *sseWriter {
	flusher, ok := w.(http.Flusher)
	var flushFn func()
	if ok {
		flushFn = flusher.Flush
	}
	return &sseWriter{w: w, flush: flushFn, enabled: ok}
}

// Emit writes a single StreamEvent to the connection.
func (s *sseWriter) Emit(event StreamEvent) error {
	return WriteEvent(s.w, s.flush, event)
}

// EmitSession is sugar that sends a "session" event.
func (s *sseWriter) EmitSession(sessionID string) error {
	return s.Emit(StreamEvent{Type: EventSession, SessionID: sessionID})
}

// EmitText sends a text_delta event for the named message.
func (s *sseWriter) EmitText(messageID, delta string) error {
	return s.Emit(StreamEvent{Type: EventTextDelta, MessageID: messageID, Delta: delta})
}

// EmitDone sends a final done event so the client can release its reader.
func (s *sseWriter) EmitDone() error {
	return s.Emit(StreamEvent{Type: EventDone})
}
