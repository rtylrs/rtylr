package handlers

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// backgroundReplyLimit bounds how long a reply keeps generating after its
// client left (a closed tab, a refresh, a phone putting the browser away).
const backgroundReplyLimit = 5 * time.Minute

// resilientWriter lets a reply outlive its client. Once the client is gone,
// writes and flushes succeed silently, so the stream finishes, the
// transcript recorder still sees every event, and the reply is saved.
type resilientWriter struct {
	http.ResponseWriter
	detached atomic.Bool
}

func newResilientWriter(w http.ResponseWriter) *resilientWriter {
	return &resilientWriter{ResponseWriter: w}
}

func (w *resilientWriter) Write(p []byte) (int, error) {
	if w.detached.Load() {
		return len(p), nil
	}
	n, err := w.ResponseWriter.Write(p)
	if err != nil {
		w.detached.Store(true)
		return len(p), nil
	}
	return n, nil
}

func (w *resilientWriter) WriteHeader(status int) {
	if !w.detached.Load() {
		w.ResponseWriter.WriteHeader(status)
	}
}

func (w *resilientWriter) Flush() {
	if w.detached.Load() {
		return
	}
	if flusher, ok := w.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (w *resilientWriter) Unwrap() http.ResponseWriter { return w.ResponseWriter }

func (w *resilientWriter) Detach() { w.detached.Store(true) }

// detachClient switches the response to background mode, wherever the
// resilient writer sits in the writer chain. It reports whether it found one.
func detachClient(w http.ResponseWriter) bool {
	for w != nil {
		switch t := w.(type) {
		case *resilientWriter:
			t.Detach()
			return true
		case interface{ Unwrap() http.ResponseWriter }:
			w = t.Unwrap()
		default:
			return false
		}
	}
	return false
}

// replyRegistry tracks the reply generating in each chat so the Stop button
// can end it: a stop and a closed tab look the same on the connection, and
// only a stop should cancel the reply.
type replyRegistry struct {
	mu     sync.Mutex
	nextID uint64
	active map[string]activeReply
}

type activeReply struct {
	id     uint64
	cancel context.CancelFunc
}

var activeReplies = &replyRegistry{active: map[string]activeReply{}}

func replyKey(userID int64, sessionID string) string {
	return strconv.FormatInt(userID, 10) + ":" + sessionID
}

// start registers a running reply and returns the function that removes it.
func (r *replyRegistry) start(userID int64, sessionID string, cancel context.CancelFunc) func() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	id, key := r.nextID, replyKey(userID, sessionID)
	r.active[key] = activeReply{id: id, cancel: cancel}
	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()
		if current, ok := r.active[key]; ok && current.id == id {
			delete(r.active, key)
		}
	}
}

// stop cancels the reply generating in a chat, if any.
func (r *replyRegistry) stop(userID int64, sessionID string) bool {
	r.mu.Lock()
	reply, ok := r.active[replyKey(userID, sessionID)]
	r.mu.Unlock()
	if ok {
		reply.cancel()
	}
	return ok
}
