package application

import (
	"context"
	"strings"
	"testing"
)

func TestPublicHolidaysListsDatesWithWeekdays(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		holidaysEndpoint + "2026/AE": `[{"date":"2026-12-02","localName":"اليوم الوطني","name":"National Day","global":true}]`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "calendar.public_holidays", map[string]any{"country_code": "ae", "year": 2026.0})
	if err != nil {
		t.Fatal(err)
	}
	holiday := output["holidays"].([]map[string]any)[0]
	if holiday["weekday"] != "Wednesday" || holiday["local_name"] != "اليوم الوطني" {
		t.Fatalf("holiday = %#v", holiday)
	}
	empty := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{holidaysEndpoint: ``})}
	if _, _, err := empty.ExecuteCaretTool(context.Background(), "calendar.public_holidays", map[string]any{"country_code": "LB", "year": 2026.0}); err == nil || !strings.Contains(err.Error(), "search the web") {
		t.Fatalf("an unsupported country must point to web search: %v", err)
	}
	if _, _, err := executor.ExecuteCaretTool(context.Background(), "calendar.public_holidays", map[string]any{"country": "Lebanon"}); err == nil {
		t.Fatal("a country name instead of a code was accepted")
	}
}
