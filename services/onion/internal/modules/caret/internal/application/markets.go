package application

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	quoteChartEndpoint   = "https://query1.finance.yahoo.com/v8/finance/chart/"
	quoteSearchEndpoint  = "https://query1.finance.yahoo.com/v1/finance/search"
	cryptoPriceEndpoint  = "https://api.coingecko.com/api/v3/simple/price"
	cryptoSearchEndpoint = "https://api.coingecko.com/api/v3/search"
	tickerSymbol         = regexp.MustCompile(`^[A-Z0-9.\-^=]{1,15}$`)
)

// stockQuote reads a delayed market quote for a stock, ETF, index or FX pair.
func (e *AssistantToolExecutor) stockQuote(ctx context.Context, input map[string]any) (map[string]any, error) {
	query := inputText(input, "symbol", "ticker", "company", "query", "name")
	if query == "" || len([]rune(query)) > 60 {
		return nil, errors.New("give a ticker such as AAPL or a company name")
	}
	symbol := strings.ToUpper(query)
	if !tickerSymbol.MatchString(symbol) {
		resolved, err := e.resolveTicker(ctx, query)
		if err != nil {
			return nil, err
		}
		symbol = resolved
	}
	var response struct {
		Chart struct {
			Result []struct {
				Meta struct {
					Currency      string  `json:"currency"`
					Symbol        string  `json:"symbol"`
					Exchange      string  `json:"fullExchangeName"`
					LongName      string  `json:"longName"`
					ShortName     string  `json:"shortName"`
					Price         float64 `json:"regularMarketPrice"`
					PreviousClose float64 `json:"chartPreviousClose"`
					DayHigh       float64 `json:"regularMarketDayHigh"`
					DayLow        float64 `json:"regularMarketDayLow"`
					YearHigh      float64 `json:"fiftyTwoWeekHigh"`
					YearLow       float64 `json:"fiftyTwoWeekLow"`
					MarketTime    int64   `json:"regularMarketTime"`
					Timezone      string  `json:"exchangeTimezoneName"`
				} `json:"meta"`
				Timestamps []int64 `json:"timestamp"`
				Indicators struct {
					Quote []struct {
						Close []*float64 `json:"close"`
					} `json:"quote"`
				} `json:"indicators"`
			} `json:"result"`
			Error *struct {
				Description string `json:"description"`
			} `json:"error"`
		} `json:"chart"`
	}
	err := getMarketJSON(ctx, e.httpClient(), quoteChartEndpoint+url.PathEscape(symbol)+"?range=1mo&interval=1d", &response)
	if errors.Is(err, errNotFound) && symbol == strings.ToUpper(query) {
		// "tesla" looks like a ticker but is a name; look it up once.
		if resolved, resolveErr := e.resolveTicker(ctx, query); resolveErr == nil && resolved != symbol {
			symbol = resolved
			err = getMarketJSON(ctx, e.httpClient(), quoteChartEndpoint+url.PathEscape(symbol)+"?range=1mo&interval=1d", &response)
		}
	}
	if err != nil {
		if errors.Is(err, errNotFound) {
			return nil, fmt.Errorf("no quote for %s; check the ticker", symbol)
		}
		return nil, err
	}
	if len(response.Chart.Result) == 0 {
		return nil, fmt.Errorf("no quote for %s; check the ticker", symbol)
	}
	result := response.Chart.Result[0]
	meta := result.Meta
	location, err := time.LoadLocation(meta.Timezone)
	if err != nil {
		location = time.UTC
	}
	history := []map[string]any{}
	if len(result.Indicators.Quote) > 0 {
		closes := result.Indicators.Quote[0].Close
		for i, stamp := range result.Timestamps {
			if i < len(closes) && closes[i] != nil {
				history = append(history, map[string]any{"date": time.Unix(stamp, 0).In(location).Format(time.DateOnly), "close": roundTo(*closes[i], 4)})
			}
		}
	}
	out := map[string]any{
		"symbol": meta.Symbol, "name": firstText(meta.LongName, meta.ShortName), "exchange": meta.Exchange, "currency": meta.Currency,
		"price": meta.Price, "day_low": meta.DayLow, "day_high": meta.DayHigh, "52_week_low": meta.YearLow, "52_week_high": meta.YearHigh,
		"as_of":                   time.Unix(meta.MarketTime, 0).In(location).Format("2006-01-02 15:04 MST"),
		"daily_closes_last_month": history,
		"note":                    "Quotes can be delayed; this is information, not investment advice.",
	}
	if len(history) >= 2 {
		if previous, _ := history[len(history)-2]["close"].(float64); previous != 0 {
			out["change_percent_vs_previous_close"] = roundTo((meta.Price-previous)/previous*100, 2)
		}
	}
	return out, nil
}

func (e *AssistantToolExecutor) resolveTicker(ctx context.Context, name string) (string, error) {
	var response struct {
		Quotes []struct {
			Symbol string `json:"symbol"`
		} `json:"quotes"`
	}
	params := url.Values{"q": {name}, "quotesCount": {"3"}, "newsCount": {"0"}}
	if err := getMarketJSON(ctx, e.httpClient(), quoteSearchEndpoint+"?"+params.Encode(), &response); err != nil {
		return "", err
	}
	if len(response.Quotes) == 0 || response.Quotes[0].Symbol == "" {
		return "", errors.New("no listed security matches that name")
	}
	return response.Quotes[0].Symbol, nil
}

// getMarketJSON is getPublicJSON with the browser-like headers the quote
// service expects.
func getMarketJSON(ctx context.Context, client *http.Client, endpoint string, out any) error {
	return getPublicJSONWith(ctx, client, endpoint, out, map[string]string{"User-Agent": "Mozilla/5.0 (compatible; rtylr-caret/1.0)"})
}

func (e *AssistantToolExecutor) cryptoPrice(ctx context.Context, input map[string]any) (map[string]any, error) {
	query := strings.ToLower(inputText(input, "coin", "symbol", "name", "query", "id"))
	if query == "" || len(query) > 60 {
		return nil, errors.New("give a coin such as bitcoin or ETH")
	}
	currency := strings.ToLower(inputText(input, "currency", "vs", "in"))
	if currency == "" {
		currency = "usd"
	}
	var search struct {
		Coins []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"coins"`
	}
	if err := getPublicJSON(ctx, e.httpClient(), cryptoSearchEndpoint+"?query="+url.QueryEscape(query), &search); err != nil {
		return nil, err
	}
	if len(search.Coins) == 0 {
		return nil, errors.New("no coin matches that name")
	}
	coin := search.Coins[0]
	for _, candidate := range search.Coins {
		if strings.EqualFold(candidate.Symbol, query) || strings.EqualFold(candidate.ID, query) {
			coin = candidate
			break
		}
	}
	var prices map[string]map[string]float64
	params := url.Values{"ids": {coin.ID}, "vs_currencies": {currency}, "include_24hr_change": {"true"}, "include_market_cap": {"true"}}
	if err := getPublicJSON(ctx, e.httpClient(), cryptoPriceEndpoint+"?"+params.Encode(), &prices); err != nil {
		return nil, err
	}
	price, ok := prices[coin.ID][currency]
	if !ok {
		return nil, fmt.Errorf("no %s price for %s", strings.ToUpper(currency), coin.Name)
	}
	return map[string]any{
		"coin": coin.Name, "symbol": strings.ToUpper(coin.Symbol), "currency": strings.ToUpper(currency), "price": price,
		"change_24h_percent": roundTo(prices[coin.ID][currency+"_24h_change"], 2),
		"market_cap":         prices[coin.ID][currency+"_market_cap"],
		"note":               "Crypto prices move constantly; this is information, not investment advice.",
	}, nil
}

func firstText(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
