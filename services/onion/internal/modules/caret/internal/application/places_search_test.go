package application

import (
	"context"
	"strings"
	"testing"
)

func TestSearchPlacesSortsByDistance(t *testing.T) {
	previous := overpassEndpoints
	overpassEndpoints = []string{"https://overpass.test/api/interpreter"}
	defer func() { overpassEndpoints = previous }()
	executor := &AssistantToolExecutor{HTTP: fakePublicAPI(t, map[string]string{
		"https://overpass.test": `{"elements":[{"lat":33.9050,"lon":35.4820,"tags":{"name":"Far Cafe"}},{"lat":33.8962,"lon":35.4821,"tags":{"name":"Near Cafe","opening_hours":"08:00-23:00","addr:street":"Bliss"}}]}`,
	})}
	output, _, err := executor.ExecuteCaretTool(context.Background(), "places.search", map[string]any{"category": "Cafes", "lat": 33.896, "lon": 35.482})
	if err != nil {
		t.Fatal(err)
	}
	places := output["places"].([]map[string]any)
	if places[0]["name"] != "Near Cafe" || places[0]["street"] != "Bliss" || places[0]["distance_m"].(float64) > 50 {
		t.Fatalf("places = %#v", places)
	}
	if _, _, err := executor.ExecuteCaretTool(context.Background(), "places.search", map[string]any{"category": "dragons", "lat": 1.0, "lon": 1.0}); err == nil || !strings.Contains(err.Error(), "cafe") {
		t.Fatalf("unknown category err = %v", err)
	}
}
