package handlers

import "testing"

func TestGeneralToolSummaries(t *testing.T) {
	for _, tc := range []struct {
		name          string
		input, output map[string]any
		want          string
	}{
		{"weather.forecast", nil, map[string]any{"location": "Beirut, Lebanon", "units": "°C", "current": map[string]any{"conditions": "clear sky", "temperature_2m": 27.0}}, "Clear sky, 27°C in Beirut, Lebanon now."},
		{"currency.convert", nil, map[string]any{"from": "USD", "amount": 100.0, "conversions": []map[string]any{{"currency": "EUR", "amount": 91.5}}}, "100 USD = 91.50 EUR."},
		{"agents.run", nil, map[string]any{"reports": []map[string]any{{"status": "done"}, {"status": "incomplete"}}, "sources": []map[string]string{{"url": "https://a"}}}, "1 of 2 sub-agents reported back, citing 1 source."},
		{"utilities.calculate", nil, map[string]any{"result": 30.0}, "Result: 30."},
		{"utilities.random", nil, map[string]any{"picked": []string{"Sara"}}, "Drew Sara."},
		{"web.search", map[string]any{"query": "rate cut"}, map[string]any{"sources": []map[string]string{{}, {}}}, `Found 2 sources for "rate cut".`},
	} {
		if got := summarizeToolResult(tc.name, tc.input, tc.output, "success", ""); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}
