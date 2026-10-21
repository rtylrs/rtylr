package application

import (
	"context"
	"testing"
)

func TestConvertTimeAcrossZonesAndDays(t *testing.T) {
	e := &AssistantToolExecutor{}
	output, _, err := e.ExecuteCaretTool(context.Background(), "utilities.convert_time", map[string]any{"time": "2026-10-11 23:30", "from": "Asia/Beirut", "to": "America/New_York, Asia/Tokyo"})
	if err != nil {
		t.Fatal(err)
	}
	to := output["to"].([]map[string]any)
	if to[0]["local_time"] != "Sun 11 Oct 2026, 16:30" || to[0]["utc_offset"] != "UTC-04:00" || to[0]["day_shift"] != "same day" {
		t.Fatalf("new york = %#v", to[0])
	}
	if to[1]["local_time"] != "Mon 12 Oct 2026, 05:30" || to[1]["day_shift"] != "next day" {
		t.Fatalf("tokyo = %#v", to[1])
	}
	if _, _, err := e.ExecuteCaretTool(context.Background(), "utilities.convert_time", map[string]any{"from": "Mars/Base", "to": "UTC"}); err == nil {
		t.Fatal("bad zone accepted")
	}
}
