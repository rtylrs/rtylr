package application

import "testing"

func TestParseModeNormalizesWireInput(t *testing.T) {
	if got := ParseMode("  AGENT "); got != ModeAgent {
		t.Fatalf("ParseMode = %q, want %q", got, ModeAgent)
	}
	if got := ParseMode("unknown"); got != ModeAsk {
		t.Fatalf("unknown mode = %q, want ask", got)
	}
}
