package httpmetrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestStatusRecorderPreservesFlusher(t *testing.T) {
	rec := httptest.NewRecorder()
	rw := &statusRecorder{ResponseWriter: rec}

	flusher, ok := any(rw).(http.Flusher)
	if !ok {
		t.Fatal("statusRecorder should preserve http.Flusher")
	}

	flusher.Flush()
	if !rec.Flushed {
		t.Fatal("Flush was not forwarded to the wrapped response writer")
	}
}
