package application

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	latestRatesEndpoint     = "https://open.er-api.com/v6/latest/"
	historicalRatesEndpoint = "https://api.frankfurter.dev/v1/"
	currencyCode            = regexp.MustCompile(`^[A-Z]{3}$`)
)

// convertCurrency converts with today's rates (about 160 currencies) or, for
// a past date, the European Central Bank reference rates.
func (e *AssistantToolExecutor) convertCurrency(ctx context.Context, input map[string]any) (map[string]any, error) {
	from := strings.ToUpper(inputText(input, "from", "base", "source"))
	targets := strings.FieldsFunc(strings.ToUpper(inputText(input, "to", "target", "targets", "symbols")), func(r rune) bool { return r == ',' || r == ' ' })
	if list, ok := input["to"].([]any); ok {
		for _, item := range list {
			if code, ok := item.(string); ok {
				targets = append(targets, strings.ToUpper(strings.TrimSpace(code)))
			}
		}
	}
	if !currencyCode.MatchString(from) || len(targets) == 0 || len(targets) > 10 {
		return nil, errors.New("use ISO currency codes, e.g. {\"amount\":100,\"from\":\"USD\",\"to\":\"EUR\"}")
	}
	for _, code := range targets {
		if !currencyCode.MatchString(code) {
			return nil, fmt.Errorf("%q is not an ISO currency code", code)
		}
	}
	amount, ok := inputNumber(input, "amount", "value")
	if !ok {
		amount = 1
	}
	date := inputText(input, "date", "on")
	var rates map[string]float64
	var asOf, source string
	if date != "" && date != time.Now().UTC().Format(time.DateOnly) {
		if _, err := time.Parse(time.DateOnly, date); err != nil {
			return nil, errors.New("date must be YYYY-MM-DD")
		}
		var response struct {
			Date  string             `json:"date"`
			Rates map[string]float64 `json:"rates"`
		}
		params := url.Values{"base": {from}, "symbols": {strings.Join(targets, ",")}}
		if err := getPublicJSON(ctx, e.httpClient(), historicalRatesEndpoint+date+"?"+params.Encode(), &response); err != nil {
			if errors.Is(err, errNotFound) {
				return nil, errors.New("historical rates cover the currencies the European Central Bank publishes, from 1999")
			}
			return nil, err
		}
		rates, asOf, source = response.Rates, response.Date, "European Central Bank reference rates"
	} else {
		var response struct {
			Result     string             `json:"result"`
			ErrorType  string             `json:"error-type"`
			UpdatedUTC string             `json:"time_last_update_utc"`
			Rates      map[string]float64 `json:"rates"`
		}
		if err := getPublicJSON(ctx, e.httpClient(), latestRatesEndpoint+url.PathEscape(from), &response); err != nil {
			return nil, err
		}
		if response.Result != "success" {
			return nil, fmt.Errorf("%s is not a supported currency", from)
		}
		rates, asOf, source = response.Rates, response.UpdatedUTC, "daily market rates (open.er-api.com)"
	}
	conversions := make([]map[string]any, 0, len(targets))
	for _, code := range targets {
		rate, ok := rates[code]
		if !ok {
			return nil, fmt.Errorf("no rate for %s from %s", code, from)
		}
		conversions = append(conversions, map[string]any{
			"currency": code, "rate": rate, "amount": roundTo(amount*rate, 4),
		})
	}
	return map[string]any{
		"from": from, "amount": amount, "conversions": conversions,
		"as_of": asOf, "source": source,
		"note": "Reference rates; a bank or card rate will differ by its spread.",
	}, nil
}

func roundTo(value float64, places int) float64 {
	scale := math.Pow(10, float64(places))
	return math.Round(value*scale) / scale
}
