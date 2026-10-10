package application

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"time"
)

var (
	weatherArchiveEndpoint = "https://archive-api.open-meteo.com/v1/archive"
	airQualityEndpoint     = "https://air-quality-api.open-meteo.com/v1/air-quality"
)

// weatherHistory reports the observed daily weather for past dates, e.g. to
// check whether rain explains a slow sales day.
func (e *AssistantToolExecutor) weatherHistory(ctx context.Context, input map[string]any) (map[string]any, error) {
	lat, lon, label, err := e.coordinates(ctx, input)
	if err != nil {
		return nil, err
	}
	start, err := parseToolDate(inputText(input, "start_date", "from", "date", "start"))
	if err != nil {
		return nil, err
	}
	end := start
	if raw := inputText(input, "end_date", "to", "end"); raw != "" {
		if end, err = parseToolDate(raw); err != nil {
			return nil, err
		}
	}
	if end.Before(start) {
		start, end = end, start
	}
	if end.Sub(start) > 92*24*time.Hour {
		return nil, errors.New("ask for at most about three months at a time")
	}
	if !end.Before(time.Now().UTC().Truncate(24 * time.Hour)) {
		return nil, errors.New("history covers past days; use weather.forecast for today and later")
	}
	params := url.Values{
		"latitude": {strconv.FormatFloat(lat, 'f', 4, 64)}, "longitude": {strconv.FormatFloat(lon, 'f', 4, 64)},
		"start_date": {start.Format(time.DateOnly)}, "end_date": {end.Format(time.DateOnly)},
		"daily":    {"weather_code,temperature_2m_max,temperature_2m_min,precipitation_sum,wind_speed_10m_max"},
		"timezone": {"auto"},
	}
	var response struct {
		Daily struct {
			Time []string   `json:"time"`
			Code []*float64 `json:"weather_code"`
			Max  []*float64 `json:"temperature_2m_max"`
			Min  []*float64 `json:"temperature_2m_min"`
			Rain []*float64 `json:"precipitation_sum"`
			Wind []*float64 `json:"wind_speed_10m_max"`
		} `json:"daily"`
	}
	if err := getPublicJSON(ctx, e.httpClient(), weatherArchiveEndpoint+"?"+params.Encode(), &response); err != nil {
		return nil, err
	}
	days := make([]map[string]any, 0, len(response.Daily.Time))
	value := func(list []*float64, i int) any {
		if i < len(list) && list[i] != nil {
			return *list[i]
		}
		return nil
	}
	for i, day := range response.Daily.Time {
		entry := map[string]any{
			"date": day, "high_c": value(response.Daily.Max, i), "low_c": value(response.Daily.Min, i),
			"precipitation_mm": value(response.Daily.Rain, i), "max_wind_kmh": value(response.Daily.Wind, i),
		}
		if code, ok := value(response.Daily.Code, i).(float64); ok {
			entry["conditions"] = weatherCodeText(int(code))
		}
		days = append(days, entry)
	}
	out := map[string]any{"days": days, "note": "The last few days may be provisional."}
	if label != "" {
		out["location"] = label
	}
	return out, nil
}

func (e *AssistantToolExecutor) airQuality(ctx context.Context, input map[string]any) (map[string]any, error) {
	lat, lon, label, err := e.coordinates(ctx, input)
	if err != nil {
		return nil, err
	}
	params := url.Values{
		"latitude": {strconv.FormatFloat(lat, 'f', 4, 64)}, "longitude": {strconv.FormatFloat(lon, 'f', 4, 64)},
		"current":  {"us_aqi,european_aqi,pm2_5,pm10,ozone,nitrogen_dioxide,uv_index"},
		"timezone": {"auto"},
	}
	var response struct {
		Current map[string]any `json:"current"`
	}
	if err := getPublicJSON(ctx, e.httpClient(), airQualityEndpoint+"?"+params.Encode(), &response); err != nil {
		return nil, err
	}
	current := map[string]any{}
	for key, v := range response.Current {
		if key != "interval" {
			current[key] = v
		}
	}
	if aqi, ok := current["us_aqi"].(float64); ok {
		current["us_aqi_category"] = usAQICategory(aqi)
	}
	out := map[string]any{"current": current}
	if label != "" {
		out["location"] = label
	}
	return out, nil
}

func usAQICategory(aqi float64) string {
	switch {
	case aqi <= 50:
		return "good"
	case aqi <= 100:
		return "moderate"
	case aqi <= 150:
		return "unhealthy for sensitive groups"
	case aqi <= 200:
		return "unhealthy"
	case aqi <= 300:
		return "very unhealthy"
	default:
		return "hazardous"
	}
}
