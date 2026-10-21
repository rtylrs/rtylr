package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"unicode/utf8"
)

// Every possible cut point of the hold-back window must keep characters whole.
func TestStreamRendererNeverSplitsMultiByteCharacters(t *testing.T) {
	for pad := 0; pad < 4; pad++ {
		rec := httptest.NewRecorder()
		var full strings.Builder
		renderer := newStreamEventRenderer(rec, nil, "m1", &full, withStreamMode(ModeAgent))

		text := strings.Repeat("a", pad) + "I don't have that here — the tools I have are limited. مرحبا بك " + strings.Repeat("b", markerHoldBytes)
		for _, r := range text {
			if err := renderer.AppendText(string(r)); err != nil {
				t.Fatalf("append: %v", err)
			}
		}
		if err := renderer.FlushPending(); err != nil {
			t.Fatalf("flush: %v", err)
		}

		var streamed strings.Builder
		for _, line := range strings.Split(rec.Body.String(), "\n") {
			payload, ok := strings.CutPrefix(line, "data: ")
			if !ok {
				continue
			}
			var event StreamEvent
			if err := json.Unmarshal([]byte(payload), &event); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if event.Type != EventTextDelta {
				continue
			}
			if !utf8.ValidString(event.Delta) || strings.ContainsRune(event.Delta, utf8.RuneError) {
				t.Fatalf("pad %d: delta split a character: %q", pad, event.Delta)
			}
			streamed.WriteString(event.Delta)
		}
		if !strings.Contains(streamed.String(), "the tools I have") || !strings.Contains(streamed.String(), "مرحبا") {
			t.Fatalf("pad %d: text mangled: %q", pad, streamed.String())
		}
	}
}
