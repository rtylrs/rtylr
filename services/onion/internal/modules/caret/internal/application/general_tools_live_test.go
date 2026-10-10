package application

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"
)

// TestGeneralToolsLive calls the real public APIs. Run with CARET_LIVE=1.
func TestGeneralToolsLive(t *testing.T) {
	if os.Getenv("CARET_LIVE") == "" {
		t.Skip("set CARET_LIVE=1 to call the public APIs")
	}
	executor := NewAssistantToolExecutor(nil, nil)
	yesterday := time.Now().AddDate(0, 0, -3).Format(time.DateOnly)
	for name, input := range map[string]map[string]any{
		"web.fetch":                {"url": "https://example.com"},
		"places.geocode":           {"name": "Beirut"},
		"weather.forecast":         {"location": "Beirut", "days": 2},
		"weather.history":          {"location": "Beirut", "date": yesterday},
		"weather.air_quality":      {"location": "Beirut"},
		"currency.convert":         {"amount": 100, "from": "USD", "to": "EUR,LBP"},
		"calendar.public_holidays": {"country_code": "US", "upcoming": true},
		"knowledge.wikipedia":      {"query": "Byblos"},
		"knowledge.define":         {"word": "resilient"},
		"knowledge.country":        {"country": "Lebanon"},
		"knowledge.papers":         {"query": "customer loyalty programs retail"},
		"markets.quote":            {"symbol": "apple"},
		"markets.crypto":           {"coin": "btc"},
		"places.search":            {"category": "cafe", "location": "Hamra, Beirut"},
	} {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), ToolTimeout(name))
			defer cancel()
			output, _, err := executor.ExecuteCaretTool(ctx, name, input)
			if err != nil {
				t.Fatalf("%s: %v", name, err)
			}
			data, _ := json.Marshal(output)
			if len(data) > 300 {
				data = data[:300]
			}
			t.Logf("%s", data)
		})
	}
}
