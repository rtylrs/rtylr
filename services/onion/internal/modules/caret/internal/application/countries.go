package application

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"strings"
	"sync"
)

// countriesData is a compact country table generated from mledoze/countries
// (ODbL): names, codes, capitals, currencies, languages, region, dialing
// code, area, borders and internet domain. These change rarely, so they ship
// with the server instead of depending on an API.
//
//go:embed countries_data.json
var countriesData []byte

type countryFacts struct {
	Name        string   `json:"name"`
	Official    string   `json:"official"`
	CCA2        string   `json:"cca2"`
	CCA3        string   `json:"cca3"`
	Capital     []string `json:"capital"`
	Currencies  []string `json:"currencies"`
	Languages   []string `json:"languages"`
	Region      string   `json:"region"`
	Subregion   string   `json:"subregion"`
	DialingCode string   `json:"dialing_code"`
	AreaKm2     float64  `json:"area_km2"`
	Borders     []string `json:"borders"`
	TLD         []string `json:"tld"`
	Alt         []string `json:"alt"`
}

var (
	countriesOnce sync.Once
	countries     []countryFacts
)

func init() {
	registerGeneralTool("knowledge.country", (*AssistantToolExecutor).countryFactsTool)
}

func loadCountries() []countryFacts {
	countriesOnce.Do(func() {
		_ = json.Unmarshal(countriesData, &countries)
	})
	return countries
}

// findCountry matches an ISO code, a common or official name, or a known
// alternative spelling; a unique partial name match is accepted too.
func findCountry(query string) (countryFacts, bool) {
	q := strings.ToLower(strings.TrimSpace(query))
	var partial []countryFacts
	for _, country := range loadCountries() {
		if q == strings.ToLower(country.CCA2) || q == strings.ToLower(country.CCA3) ||
			q == strings.ToLower(country.Name) || q == strings.ToLower(country.Official) {
			return country, true
		}
		for _, alt := range country.Alt {
			if q == strings.ToLower(alt) {
				return country, true
			}
		}
		if len(q) >= 4 && strings.Contains(strings.ToLower(country.Name), q) {
			partial = append(partial, country)
		}
	}
	if len(partial) == 1 {
		return partial[0], true
	}
	return countryFacts{}, false
}

func (e *AssistantToolExecutor) countryFactsTool(_ context.Context, input map[string]any) (map[string]any, error) {
	query := inputText(input, "country", "name", "code", "query")
	if query == "" || len([]rune(query)) > 80 {
		return nil, errors.New("give a country name or ISO code")
	}
	country, ok := findCountry(query)
	if !ok {
		return nil, errors.New("no single country matches that name; try its full name or ISO code")
	}
	borders := make([]string, 0, len(country.Borders))
	for _, code := range country.Borders {
		if neighbour, ok := findCountry(code); ok {
			borders = append(borders, neighbour.Name)
		}
	}
	return map[string]any{
		"name": country.Name, "official_name": country.Official, "country_code": country.CCA2,
		"capital": strings.Join(country.Capital, ", "), "currencies": country.Currencies, "languages": country.Languages,
		"region": country.Region, "subregion": country.Subregion, "dialing_code": country.DialingCode,
		"area_km2": country.AreaKm2, "borders": borders, "internet_domain": strings.Join(country.TLD, ", "),
		"note": "For population and other figures that change, check knowledge.wikipedia or the web.",
	}, nil
}
