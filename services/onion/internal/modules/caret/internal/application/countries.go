package application

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strings"
)

var countriesEndpoint = "https://restcountries.com/v3.1/"

func init() {
	registerGeneralTool("knowledge.country", (*AssistantToolExecutor).countryFacts)
}

func (e *AssistantToolExecutor) countryFacts(ctx context.Context, input map[string]any) (map[string]any, error) {
	query := inputText(input, "country", "name", "code", "query")
	if query == "" || len([]rune(query)) > 80 {
		return nil, errors.New("give a country name or ISO code")
	}
	fields := "?fields=name,cca2,capital,currencies,languages,population,region,subregion,timezones,idd"
	endpoint := countriesEndpoint + "name/" + url.PathEscape(query) + fields
	if len(query) == 2 || len(query) == 3 {
		endpoint = countriesEndpoint + "alpha/" + url.PathEscape(query) + fields
	}
	var countries []struct {
		Name struct {
			Common   string `json:"common"`
			Official string `json:"official"`
		} `json:"name"`
		Code       string   `json:"cca2"`
		Capital    []string `json:"capital"`
		Currencies map[string]struct {
			Name   string `json:"name"`
			Symbol string `json:"symbol"`
		} `json:"currencies"`
		Languages  map[string]string `json:"languages"`
		Population int64             `json:"population"`
		Region     string            `json:"region"`
		Subregion  string            `json:"subregion"`
		Timezones  []string          `json:"timezones"`
		IDD        struct {
			Root     string   `json:"root"`
			Suffixes []string `json:"suffixes"`
		} `json:"idd"`
	}
	if err := decodeOneOrMany(ctx, e, endpoint, &countries); err != nil {
		if errors.Is(err, errNotFound) {
			return nil, errors.New("no country matches that name")
		}
		return nil, err
	}
	if len(countries) == 0 {
		return nil, errors.New("no country matches that name")
	}
	// A name search can match several countries ("Guinea"); prefer an exact name.
	best := countries[0]
	for _, country := range countries {
		if strings.EqualFold(country.Name.Common, query) || strings.EqualFold(country.Name.Official, query) {
			best = country
			break
		}
	}
	currencies := []string{}
	for code, currency := range best.Currencies {
		currencies = append(currencies, code+" ("+currency.Name+", "+currency.Symbol+")")
	}
	languages := []string{}
	for _, language := range best.Languages {
		languages = append(languages, language)
	}
	sort.Strings(currencies)
	sort.Strings(languages)
	dialing := best.IDD.Root
	if len(best.IDD.Suffixes) == 1 {
		dialing += best.IDD.Suffixes[0]
	}
	return map[string]any{
		"name": best.Name.Common, "official_name": best.Name.Official, "country_code": best.Code,
		"capital": strings.Join(best.Capital, ", "), "currencies": currencies, "languages": languages,
		"population": best.Population, "region": best.Region, "subregion": best.Subregion,
		"timezones": best.Timezones, "dialing_code": dialing,
	}, nil
}

// decodeOneOrMany reads an endpoint that answers with an object for one
// match (alpha lookups) or an array (name searches).
func decodeOneOrMany[T any](ctx context.Context, e *AssistantToolExecutor, endpoint string, out *[]T) error {
	var raw any
	if err := getPublicJSON(ctx, e.httpClient(), endpoint, &raw); err != nil {
		return err
	}
	if _, isList := raw.([]any); !isList {
		raw = []any{raw}
	}
	return remarshal(raw, out)
}

func remarshal(in, out any) error {
	data, err := json.Marshal(in)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, out)
}
