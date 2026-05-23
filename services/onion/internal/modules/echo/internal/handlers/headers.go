package handlers

import "net/http"

// SetSSEHeaders applies the Server-Sent Events response headers and disables
// every intermediate buffer we know about.
func SetSSEHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache, no-transform")
	h.Set("Connection", "keep-alive")
	// nginx + Cloudflare both honor X-Accel-Buffering: no.
	h.Set("X-Accel-Buffering", "no")
}
