package application

import (
	"context"
	"testing"
)

func TestStockQuoteResolvesNamesAndReportsChange(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		quoteSearchEndpoint:         `{"quotes":[{"symbol":"TSLA"}]}`,
		quoteChartEndpoint + "TSLA": `{"chart":{"result":[{"meta":{"currency":"USD","symbol":"TSLA","longName":"Tesla, Inc.","regularMarketPrice":110,"regularMarketTime":1791576000,"exchangeTimezoneName":"America/New_York"},"timestamp":[1791489600,1791576000],"indicators":{"quote":[{"close":[100,110]}]}}]}}`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "markets.quote", map[string]any{"symbol": "Tesla motors"})
	if err != nil {
		t.Fatal(err)
	}
	if output["symbol"] != "TSLA" || output["name"] != "Tesla, Inc." || output["change_percent_vs_previous_close"] != 10.0 {
		t.Fatalf("output = %#v", output)
	}
}

func TestCryptoPricePrefersExactSymbol(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		cryptoSearchEndpoint: `{"coins":[{"id":"ethereum-classic","name":"Ethereum Classic","symbol":"ETC"},{"id":"ethereum","name":"Ethereum","symbol":"ETH"}]}`,
		cryptoPriceEndpoint:  `{"ethereum":{"eur":2000.5,"eur_24h_change":-1.234,"eur_market_cap":240000000000}}`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "markets.crypto", map[string]any{"coin": "eth", "currency": "EUR"})
	if err != nil {
		t.Fatal(err)
	}
	if output["coin"] != "Ethereum" || output["price"] != 2000.5 || output["change_24h_percent"] != -1.23 {
		t.Fatalf("output = %#v", output)
	}
}
