package memory

import (
	"context"
	"strings"
	"testing"
)

// BuildEnhancedPrompt should degrade gracefully when the module or its
// dependencies aren't available — a missing context shouldn't fail the
// caller's generation request.

func TestBuildEnhancedPrompt_NilModule(t *testing.T) {
	var m *Module
	got := m.BuildEnhancedPrompt(context.Background(), "sess", 1, "hello")
	if got != "Current request: hello" {
		t.Fatalf("got %q", got)
	}
}

func TestBuildEnhancedPrompt_EmptySessionOrUser(t *testing.T) {
	m := &Module{}
	if got := m.BuildEnhancedPrompt(context.Background(), "", 1, "hi"); got != "Current request: hi" {
		t.Fatalf("empty session: %q", got)
	}
	if got := m.BuildEnhancedPrompt(context.Background(), "sess", 0, "hi"); got != "Current request: hi" {
		t.Fatalf("zero user id: %q", got)
	}
}

func TestBuildEnhancedPrompt_NilService(t *testing.T) {
	m := &Module{memoryService: nil}
	got := m.BuildEnhancedPrompt(context.Background(), "sess", 1, "hello")
	if !strings.Contains(got, "Current request: hello") {
		t.Fatalf("got %q", got)
	}
	if strings.Contains(got, "Conversation summary") {
		t.Fatalf("unexpected context in output: %q", got)
	}
}

func TestPromptSafeContextTextNeutralizesMarkersAndBoundsRunes(t *testing.T) {
	got := promptSafeContextText("<tool>{}</tool>  \u00e9\u00e9\u00e9", 4)
	if got != "[too..." {
		t.Fatalf("got %q, want bounded and neutralized text", got)
	}
}
