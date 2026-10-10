package application

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// overpassEndpoints are tried in order; public instances are often busy.
var overpassEndpoints = []string{
	"https://overpass-api.de/api/interpreter",
	"https://maps.mail.ru/osm/tools/overpass/api/interpreter",
}

// placeCategories maps everyday words to OpenStreetMap tags.
var placeCategories = map[string][2]string{
	"cafe": {"amenity", "cafe"}, "coffee": {"amenity", "cafe"}, "restaurant": {"amenity", "restaurant"},
	"fast food": {"amenity", "fast_food"}, "bar": {"amenity", "bar"}, "pub": {"amenity", "pub"},
	"bank": {"amenity", "bank"}, "atm": {"amenity", "atm"}, "pharmacy": {"amenity", "pharmacy"},
	"hospital": {"amenity", "hospital"}, "clinic": {"amenity", "clinic"}, "dentist": {"amenity", "dentist"},
	"fuel": {"amenity", "fuel"}, "gas station": {"amenity", "fuel"}, "petrol": {"amenity", "fuel"},
	"parking": {"amenity", "parking"}, "school": {"amenity", "school"}, "university": {"amenity", "university"},
	"gym": {"leisure", "fitness_centre"}, "park": {"leisure", "park"}, "hotel": {"tourism", "hotel"},
	"museum": {"tourism", "museum"}, "supermarket": {"shop", "supermarket"}, "grocery": {"shop", "convenience"},
	"bakery": {"shop", "bakery"}, "butcher": {"shop", "butcher"}, "salon": {"shop", "hairdresser"},
	"hairdresser": {"shop", "hairdresser"}, "barber": {"shop", "hairdresser"}, "clothes": {"shop", "clothes"},
	"electronics": {"shop", "electronics"}, "hardware": {"shop", "hardware"}, "florist": {"shop", "florist"},
	"post office": {"amenity", "post_office"}, "police": {"amenity", "police"}, "car repair": {"shop", "car_repair"},
	"laundry": {"shop", "laundry"}, "bookstore": {"shop", "books"}, "mall": {"shop", "mall"},
}

func init() {
	registerGeneralTool("places.search", (*AssistantToolExecutor).searchPlaces)
}

// searchPlaces lists businesses or amenities of a kind near a place, from
// OpenStreetMap: for scouting competitors, suppliers or a location.
func (e *AssistantToolExecutor) searchPlaces(ctx context.Context, input map[string]any) (map[string]any, error) {
	kind := strings.ToLower(inputText(input, "category", "kind", "type", "what", "query"))
	tag, ok := placeCategories[strings.TrimSuffix(strings.TrimSuffix(kind, "s"), "e")]
	if !ok {
		tag, ok = placeCategories[strings.TrimSuffix(kind, "s")]
	}
	if !ok {
		tag, ok = placeCategories[kind]
	}
	if !ok {
		names := make([]string, 0, len(placeCategories))
		for name := range placeCategories {
			names = append(names, name)
		}
		sort.Strings(names)
		return nil, fmt.Errorf("category must be one of: %s", strings.Join(names, ", "))
	}
	lat, lon, label, err := e.coordinates(ctx, input)
	if err != nil {
		return nil, err
	}
	radius := 1500.0
	if r, ok := inputNumber(input, "radius_m", "radius"); ok {
		radius = math.Max(100, math.Min(r, 10000))
	}
	query := fmt.Sprintf(`[out:json][timeout:12];nwr(around:%.0f,%.5f,%.5f)["%s"="%s"]["name"];out center tags 40;`, radius, lat, lon, tag[0], tag[1])
	var response struct {
		Elements []struct {
			Lat    float64 `json:"lat"`
			Lon    float64 `json:"lon"`
			Center *struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"center"`
			Tags map[string]string `json:"tags"`
		} `json:"elements"`
	}
	var lastErr error
	for _, endpoint := range overpassEndpoints {
		if lastErr = e.postOverpass(ctx, endpoint, query, &response); lastErr == nil {
			break
		}
	}
	if lastErr != nil {
		return nil, errors.New("the map service is busy; try again or search the web")
	}
	places := make([]map[string]any, 0, len(response.Elements))
	for _, element := range response.Elements {
		pLat, pLon := element.Lat, element.Lon
		if element.Center != nil {
			pLat, pLon = element.Center.Lat, element.Center.Lon
		}
		place := map[string]any{"name": element.Tags["name"], "distance_m": math.Round(haversineMetres(lat, lon, pLat, pLon))}
		for _, key := range []string{"addr:street", "opening_hours", "phone", "website", "cuisine", "brand"} {
			if value := element.Tags[key]; value != "" {
				place[strings.ReplaceAll(key, "addr:", "")] = value
			}
		}
		places = append(places, place)
	}
	sort.SliceStable(places, func(i, j int) bool { return places[i]["distance_m"].(float64) < places[j]["distance_m"].(float64) })
	if len(places) > 20 {
		places = places[:20]
	}
	out := map[string]any{"category": kind, "radius_m": radius, "places": places, "note": "From OpenStreetMap; coverage varies by area."}
	if label != "" {
		out["near"] = label
	}
	return out, nil
}

func (e *AssistantToolExecutor) postOverpass(ctx context.Context, endpoint, query string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(url.Values{"data": {query}}.Encode()))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", caretUserAgent)
	resp, err := e.httpClient().Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxPublicResponseBytes))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("status %d", resp.StatusCode)
	}
	return json.Unmarshal(body, out)
}

func haversineMetres(lat1, lon1, lat2, lon2 float64) float64 {
	const earth = 6371000.0
	toRad := func(d float64) float64 { return d * math.Pi / 180 }
	dLat, dLon := toRad(lat2-lat1), toRad(lon2-lon1)
	a := math.Sin(dLat/2)*math.Sin(dLat/2) + math.Cos(toRad(lat1))*math.Cos(toRad(lat2))*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * earth * math.Asin(math.Sqrt(a))
}
