package infrastructure

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestTruncateReason_PreservesValidUTF8(t *testing.T) {
	// 254 ASCII bytes followed by a 3-byte Arabic character: byte slicing at
	// 255 would split the multi-byte character and leave invalid UTF-8.
	in := strings.Repeat("a", 254) + "ع" + "extra"
	got := truncateReason(in, 255)
	if !utf8.ValidString(got) {
		t.Fatalf("truncateReason produced invalid UTF-8: %q", got)
	}
	if utf8.RuneCountInString(got) != 255 {
		t.Fatalf("rune count = %d, want 255", utf8.RuneCountInString(got))
	}
	if !strings.HasSuffix(got, "ع") {
		t.Fatalf("expected last rune to be preserved, got %q", got)
	}
}
