package application

import (
	"context"
	"strings"
	"testing"
)

func TestCurrencyConvertUsesLatestAndHistoricalRates(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		latestRatesEndpoint + "USD":            `{"result":"success","time_last_update_utc":"Sat, 10 Oct 2026 00:00:01 +0000","rates":{"EUR":0.9,"LBP":89500}}`,
		historicalRatesEndpoint + "2020-01-02": `{"date":"2020-01-02","rates":{"EUR":0.8933}}`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "currency.convert", map[string]any{"amount": 250.0, "from": "usd", "to": "EUR, LBP"})
	if err != nil {
		t.Fatal(err)
	}
	conversions := output["conversions"].([]map[string]any)
	if len(conversions) != 2 || conversions[0]["amount"] != 225.0 || conversions[1]["amount"] != 22375000.0 {
		t.Fatalf("conversions = %#v", conversions)
	}
	output, _, err = executor.ExecuteCaretTool(context.Background(), "currency.convert", map[string]any{"from": "USD", "to": "EUR", "date": "2020-01-02"})
	if err != nil || !strings.Contains(output["source"].(string), "European Central Bank") {
		t.Fatalf("historical = %#v, %v", output, err)
	}
}

func TestCurrencyConvertRejectsBadCodes(t *testing.T) {
	executor := &AssistantToolExecutor{}
	for _, input := range []map[string]any{{"from": "dollars", "to": "EUR"}, {"from": "USD"}, {"from": "USD", "to": "EURO"}} {
		if _, _, err := executor.ExecuteCaretTool(context.Background(), "currency.convert", input); err == nil {
			t.Errorf("accepted %v", input)
		}
	}
}
