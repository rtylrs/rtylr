package application

import (
	"context"
	"testing"
	"time"
)

func TestDateMathSpanCountsWorkingDays(t *testing.T) {
	e := &AssistantToolExecutor{}
	output, _, err := e.ExecuteCaretTool(context.Background(), "utilities.date_math", map[string]any{"from": "2026-10-01", "to": "2026-10-15"})
	if err != nil {
		t.Fatal(err)
	}
	if output["days"] != 14 || output["working_days"] != 10 || output["calendar"] != "0 years, 0 months, 14 days" {
		t.Fatalf("span = %#v", output)
	}
	output, _, _ = e.ExecuteCaretTool(context.Background(), "utilities.date_math", map[string]any{"from": "2026-10-01", "to": "2026-10-15", "weekend": "fri-sat"})
	if output["working_days"] != 10 {
		t.Fatalf("fri-sat working days = %v", output["working_days"])
	}
}

func TestDateMathAddsMonthsAndWorkingDays(t *testing.T) {
	e := &AssistantToolExecutor{}
	output, _, err := e.ExecuteCaretTool(context.Background(), "utilities.date_math", map[string]any{"from": "2026-10-09", "business_days": 3.0})
	if err != nil {
		t.Fatal(err)
	}
	result := output["result"].(map[string]any)
	if result["date"] != "2026-10-14" || result["weekday"] != "Wednesday" {
		t.Fatalf("Friday + 3 working days = %#v", result)
	}
	output, _, _ = e.ExecuteCaretTool(context.Background(), "utilities.date_math", map[string]any{"from": "2026-01-31", "months": 1.0})
	if output["result"].(map[string]any)["date"] != "2026-02-28" {
		t.Fatalf("Jan 31 + 1 month = %#v", output["result"])
	}
	if _, _, err := e.ExecuteCaretTool(context.Background(), "utilities.date_math", map[string]any{"from": "next blursday"}); err == nil {
		t.Fatal("nonsense date accepted")
	}
}

func TestDateMathTodayIsThePersonsLocalDay(t *testing.T) {
	e := &AssistantToolExecutor{}
	far := time.FixedZone("UTC+14", 14*3600)
	ctx := WithUserLocation(context.Background(), far)
	output, _, err := e.ExecuteCaretTool(ctx, "utilities.date_math", map[string]any{"from": "today", "days": 0.0})
	if err != nil {
		t.Fatal(err)
	}
	want := time.Now().In(far).Format(time.DateOnly)
	if got := output["result"].(map[string]any)["date"]; got != want {
		t.Fatalf("today = %v, want %v", got, want)
	}
}
