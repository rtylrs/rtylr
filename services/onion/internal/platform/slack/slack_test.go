package slack

import (
	"strings"
	"testing"
)

func TestChunkCombinesUnderLimit(t *testing.T) {
	lines := []string{"a", "b", "c"}
	got := chunk(lines, 1000)
	if len(got) != 1 {
		t.Fatalf("expected 1 combined chunk, got %d", len(got))
	}
	for _, l := range lines {
		if !strings.Contains(got[0], l) {
			t.Errorf("chunk missing %q", l)
		}
	}
}

func TestChunkSplitsOverLimit(t *testing.T) {
	long := strings.Repeat("x", 30)
	lines := []string{long, long, long}
	got := chunk(lines, 40) // each line + separator can't all fit in 40
	if len(got) < 2 {
		t.Fatalf("expected multiple chunks, got %d", len(got))
	}
	for _, c := range got {
		if len(c) > 40 {
			t.Errorf("chunk exceeds limit: %d chars", len(c))
		}
	}
}

func TestChunkTruncatesOversizedLine(t *testing.T) {
	got := chunk([]string{strings.Repeat("y", 100)}, 40)
	if len(got) != 1 || len(got[0]) > 40 {
		t.Fatalf("oversized line not truncated to limit: %v", len(got[0]))
	}
	if !strings.HasSuffix(got[0], "...") {
		t.Error("expected truncation marker")
	}
}
