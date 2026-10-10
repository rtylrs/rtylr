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

func TestWeatherHistoryReadsPastDays(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		weatherArchiveEndpoint: `{"daily":{"time":["2026-10-02"],"weather_code":[63],"temperature_2m_max":[24.1],"temperature_2m_min":[18.0],"precipitation_sum":[12.5],"wind_speed_10m_max":[30]}}`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "weather.history", map[string]any{"latitude": 33.9, "longitude": 35.5, "date": "2026-10-02"})
	if err != nil {
		t.Fatal(err)
	}
	day := output["days"].([]map[string]any)[0]
	if day["conditions"] != "rain" || day["precipitation_mm"] != 12.5 {
		t.Fatalf("day = %#v", day)
	}
	if _, _, err := executor.ExecuteCaretTool(context.Background(), "weather.history", map[string]any{"latitude": 1.0, "longitude": 1.0, "date": "2099-01-01"}); err == nil {
		t.Fatal("a future date was accepted")
	}
}

func TestAirQualityAddsCategory(t *testing.T) {
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{airQualityEndpoint: `{"current":{"us_aqi":120,"pm2_5":40}}`})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "weather.air_quality", map[string]any{"lat": 1.0, "lon": 2.0})
	if err != nil || output["current"].(map[string]any)["us_aqi_category"] != "unhealthy for sensitive groups" {
		t.Fatalf("output = %#v, %v", output, err)
	}
}
