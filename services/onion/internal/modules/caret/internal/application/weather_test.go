package application

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func fakePublicAPI(t *testing.T, routes map[string]string) *http.Client {
	t.Helper()
	return &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		for prefix, body := range routes {
			if strings.HasPrefix(r.URL.String(), prefix) {
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Request: r, Body: io.NopCloser(strings.NewReader(body))}, nil
			}
		}
		t.Errorf("unexpected request %s", r.URL)
		return &http.Response{StatusCode: 404, Request: r, Body: io.NopCloser(strings.NewReader(""))}, nil
	})}
}

func TestWeatherForecastGeocodesAndDescribesConditions(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		geocodingEndpoint: `{"results":[{"name":"Paris","latitude":48.85,"longitude":2.35,"country":"France","admin1":"Île-de-France","timezone":"Europe/Paris"},{"name":"Paris","latitude":33.66,"longitude":-95.55,"country":"United States","admin1":"Texas","timezone":"America/Chicago"}]}`,
		forecastEndpoint:  `{"timezone":"America/Chicago","current":{"temperature_2m":30.5,"weather_code":2,"interval":900},"daily":{"time":["2026-10-11"],"weather_code":[61],"temperature_2m_max":[31.2],"temperature_2m_min":[19.8],"precipitation_sum":[4.2],"precipitation_probability_max":[70],"wind_speed_10m_max":[18],"sunrise":["2026-10-11T07:20"],"sunset":["2026-10-11T18:55"]}}`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "weather.forecast", map[string]any{"location": "Paris, Texas"})
	if err != nil {
		t.Fatal(err)
	}
	if output["location"] != "Paris, Texas, United States" {
		t.Fatalf("location = %v", output["location"])
	}
	current := output["current"].(map[string]any)
	if current["conditions"] != "partly cloudy" || current["interval"] != nil {
		t.Fatalf("current = %#v", current)
	}
	day := output["daily"].([]map[string]any)[0]
	if day["conditions"] != "rain" || day["high"] != "31°C" || day["precipitation_chance_percent"] != 70.0 {
		t.Fatalf("day = %#v", day)
	}
}

func TestGeocodeRejectsEmptyAndReportsNoMatch(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{geocodingEndpoint: `{}`})}
	if _, _, err := executor.ExecuteCaretTool(context.Background(), "places.geocode", map[string]any{}); err == nil {
		t.Fatal("empty place accepted")
	}
	if _, _, err := executor.ExecuteCaretTool(context.Background(), "places.geocode", map[string]any{"name": "Nowhereville"}); err == nil || !strings.Contains(err.Error(), "no place") {
		t.Fatalf("err = %v", err)
	}
}
