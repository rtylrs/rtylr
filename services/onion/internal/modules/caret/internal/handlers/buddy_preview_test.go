package handlers

import (
	"strings"
	"testing"
)

func TestBuddyRoutinePreviewReadsBackTheSchedule(t *testing.T) {
	got := buddyRoutinePreview("buddy.create_routine", map[string]any{
		"name": "Gym check-in", "every_minutes": float64(1440), "first_run_at": "2026-10-30T20:30:00+03:00",
	})
	for _, want := range []string{`"Gym check-in"`, "every day", "2026-10-30T20:30:00+03:00", "Buddy chat"} {
		if !strings.Contains(got, want) {
			t.Fatalf("preview %q is missing %q", got, want)
		}
	}
	if got := buddyRoutinePreview("buddy.update_routine", map[string]any{"id": float64(4), "enabled": false}); !strings.Contains(got, "pause") {
		t.Fatalf("pause preview = %q", got)
	}
	if got := everyMinutesText(float64(120)); got != "every 2 hours" {
		t.Fatalf("every = %q", got)
	}
	if buddyRoutinePreview("crm.create_customer", nil) != "" {
		t.Fatal("other tools keep the generic preview")
	}
}
