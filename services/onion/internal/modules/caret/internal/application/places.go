package application

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

var geocodingEndpoint = "https://geocoding-api.open-meteo.com/v1/search"

type place struct {
	Name       string  `json:"name"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Country    string  `json:"country"`
	Code       string  `json:"country_code"`
	Admin1     string  `json:"admin1"`
	Timezone   string  `json:"timezone"`
	Population int64   `json:"population"`
	Elevation  float64 `json:"elevation"`
}

func (p place) label() string {
	parts := []string{p.Name}
	if p.Admin1 != "" && p.Admin1 != p.Name {
		parts = append(parts, p.Admin1)
	}
	if p.Country != "" {
		parts = append(parts, p.Country)
	}
	return strings.Join(parts, ", ")
}

func (p place) output() map[string]any {
	out := map[string]any{
		"name": p.label(), "latitude": p.Latitude, "longitude": p.Longitude,
		"country": p.Country, "country_code": p.Code, "timezone": p.Timezone,
	}
	if p.Population > 0 {
		out["population"] = p.Population
	}
	return out
}

// geocode finds places by name. "Paris, Texas" narrows to matches whose
// region or country contains the part after the comma.
func (e *AssistantToolExecutor) geocode(ctx context.Context, name string, count int) ([]place, error) {
	name = strings.TrimSpace(name)
	if name == "" || len([]rune(name)) > 120 {
		return nil, errors.New("give a city or place name")
	}
	query, qualifier, _ := strings.Cut(name, ",")
	params := url.Values{"name": {strings.TrimSpace(query)}, "count": {strconv.Itoa(10)}, "language": {"en"}, "format": {"json"}}
	var response struct {
		Results []place `json:"results"`
	}
	if err := getPublicJSON(ctx, e.httpClient(), geocodingEndpoint+"?"+params.Encode(), &response); err != nil {
		return nil, err
	}
	results := response.Results
	if qualifier = strings.ToLower(strings.TrimSpace(qualifier)); qualifier != "" {
		filtered := results[:0:0]
		for _, candidate := range results {
			haystack := strings.ToLower(candidate.Admin1 + " " + candidate.Country + " " + candidate.Code)
			if strings.Contains(haystack, qualifier) {
				filtered = append(filtered, candidate)
			}
		}
		if len(filtered) > 0 {
			results = filtered
		} else if broader, err := e.geocode(ctx, qualifier, count); err == nil {
			// "Hamra, Beirut": a neighbourhood the index lacks resolves to
			// the city rather than a same-named place in another country.
			return broader, nil
		}
	}
	if len(results) == 0 {
		return nil, errors.New("no place matches that name")
	}
	if len(results) > count {
		results = results[:count]
	}
	return results, nil
}

func (e *AssistantToolExecutor) geocodeTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	places, err := e.geocode(ctx, inputText(input, "name", "place", "city", "location", "query"), 5)
	if err != nil {
		return nil, err
	}
	out := make([]map[string]any, len(places))
	for i, p := range places {
		out[i] = p.output()
	}
	return map[string]any{"places": out}, nil
}
