package api

import (
	"testing"
	"time"
)

func TestParseTimelineCursor(t *testing.T) {
	createdAt := time.Date(2026, time.August, 13, 12, 0, 0, 0, time.UTC)
	cursor := "1786622400000000000:42"
	gotTime, gotID, err := parseTimelineCursor(cursor)
	if err != nil {
		t.Fatalf("parseTimelineCursor: %v", err)
	}
	if gotID != 42 {
		t.Fatalf("id = %d, want 42", gotID)
	}
	if !gotTime.Equal(createdAt) {
		t.Fatalf("created_at = %s, want %s", gotTime, createdAt)
	}
}

func TestParseTimelineCursorRejectsMalformedValues(t *testing.T) {
	for _, cursor := range []string{"bad", "1:0", "bad:1", "1:bad"} {
		if _, _, err := parseTimelineCursor(cursor); err == nil {
			t.Errorf("parseTimelineCursor(%q) succeeded", cursor)
		}
	}
}
