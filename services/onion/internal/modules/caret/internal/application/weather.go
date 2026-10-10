package application

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
)

var forecastEndpoint = "https://api.open-meteo.com/v1/forecast"

func init() {
	registerGeneralTool("weather.forecast", (*AssistantToolExecutor).weatherForecast)
}

func (e *AssistantToolExecutor) weatherForecast(ctx context.Context, input map[string]any) (map[string]any, error) {
	lat, lon, label, err := e.coordinates(ctx, input)
	if err != nil {
		return nil, err
	}
	days := 3
	if n, ok := inputNumber(input, "days"); ok {
		days = int(n)
	}
	days = max(1, min(days, 16))
	units := "celsius"
	if unit := inputText(input, "units", "unit", "temperature_unit"); unit == "fahrenheit" || unit == "imperial" || unit == "f" {
		units = "fahrenheit"
	}
	params := url.Values{
		"latitude": {strconv.FormatFloat(lat, 'f', 4, 64)}, "longitude": {strconv.FormatFloat(lon, 'f', 4, 64)},
		"current":          {"temperature_2m,apparent_temperature,relative_humidity_2m,precipitation,weather_code,wind_speed_10m"},
		"daily":            {"weather_code,temperature_2m_max,temperature_2m_min,precipitation_sum,precipitation_probability_max,wind_speed_10m_max,sunrise,sunset"},
		"timezone":         {"auto"},
		"forecast_days":    {strconv.Itoa(days)},
		"temperature_unit": {units},
	}
	var response struct {
		Timezone     string         `json:"timezone"`
		CurrentUnits map[string]any `json:"current_units"`
		Current      map[string]any `json:"current"`
		Daily        struct {
			Time       []string   `json:"time"`
			Code       []float64  `json:"weather_code"`
			Max        []float64  `json:"temperature_2m_max"`
			Min        []float64  `json:"temperature_2m_min"`
			Rain       []float64  `json:"precipitation_sum"`
			RainChance []*float64 `json:"precipitation_probability_max"`
			Wind       []float64  `json:"wind_speed_10m_max"`
			Sunrise    []string   `json:"sunrise"`
			Sunset     []string   `json:"sunset"`
		} `json:"daily"`
	}
	if err := getPublicJSON(ctx, e.httpClient(), forecastEndpoint+"?"+params.Encode(), &response); err != nil {
		return nil, err
	}
	degree := "°C"
	if units == "fahrenheit" {
		degree = "°F"
	}
	current := map[string]any{}
	for key, value := range response.Current {
		if key == "weather_code" {
			if code, ok := value.(float64); ok {
				current["conditions"] = weatherCodeText(int(code))
			}
			continue
		}
		if key != "interval" {
			current[key] = value
		}
	}
	daily := make([]map[string]any, 0, len(response.Daily.Time))
	for i, day := range response.Daily.Time {
		entry := map[string]any{"date": day}
		if i < len(response.Daily.Code) {
			entry["conditions"] = weatherCodeText(int(response.Daily.Code[i]))
		}
		if i < len(response.Daily.Max) && i < len(response.Daily.Min) {
			entry["high"] = fmt.Sprintf("%.0f%s", response.Daily.Max[i], degree)
			entry["low"] = fmt.Sprintf("%.0f%s", response.Daily.Min[i], degree)
		}
		if i < len(response.Daily.Rain) {
			entry["precipitation_mm"] = response.Daily.Rain[i]
		}
		if i < len(response.Daily.RainChance) && response.Daily.RainChance[i] != nil {
			entry["precipitation_chance_percent"] = *response.Daily.RainChance[i]
		}
		if i < len(response.Daily.Wind) {
			entry["max_wind_kmh"] = response.Daily.Wind[i]
		}
		if i < len(response.Daily.Sunrise) && i < len(response.Daily.Sunset) {
			entry["sunrise"], entry["sunset"] = response.Daily.Sunrise[i], response.Daily.Sunset[i]
		}
		daily = append(daily, entry)
	}
	out := map[string]any{"timezone": response.Timezone, "units": degree, "current": current, "daily": daily}
	if label != "" {
		out["location"] = label
	}
	return out, nil
}

// coordinates reads latitude and longitude from the input, or geocodes a
// place name.
func (e *AssistantToolExecutor) coordinates(ctx context.Context, input map[string]any) (lat, lon float64, label string, err error) {
	lat, hasLat := inputNumber(input, "latitude", "lat")
	lon, hasLon := inputNumber(input, "longitude", "lon", "lng")
	if !hasLat || !hasLon {
		places, err := e.geocode(ctx, inputText(input, "location", "city", "place", "name"), 1)
		if err != nil {
			return 0, 0, "", err
		}
		lat, lon, label = places[0].Latitude, places[0].Longitude, places[0].label()
	}
	if lat < -90 || lat > 90 || lon < -180 || lon > 180 {
		return 0, 0, "", errors.New("latitude or longitude is out of range")
	}
	return lat, lon, label, nil
}

// weatherCodeText names a WMO weather interpretation code.
func weatherCodeText(code int) string {
	switch {
	case code == 0:
		return "clear sky"
	case code == 1:
		return "mainly clear"
	case code == 2:
		return "partly cloudy"
	case code == 3:
		return "overcast"
	case code == 45 || code == 48:
		return "fog"
	case code >= 51 && code <= 57:
		return "drizzle"
	case code >= 61 && code <= 67:
		return "rain"
	case code >= 71 && code <= 77:
		return "snow"
	case code >= 80 && code <= 82:
		return "rain showers"
	case code == 85 || code == 86:
		return "snow showers"
	case code == 95:
		return "thunderstorm"
	case code == 96 || code == 99:
		return "thunderstorm with hail"
	default:
		return "unknown"
	}
}
