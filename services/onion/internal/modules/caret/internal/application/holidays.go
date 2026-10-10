package application

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var holidaysEndpoint = "https://date.nager.at/api/v3/PublicHolidays/"

func init() {
	registerGeneralTool("calendar.public_holidays", (*AssistantToolExecutor).publicHolidays)
}

func (e *AssistantToolExecutor) publicHolidays(ctx context.Context, input map[string]any) (map[string]any, error) {
	country := strings.ToUpper(inputText(input, "country_code", "country", "code"))
	if len(country) != 2 {
		return nil, errors.New("use a two-letter ISO country code, e.g. \"LB\" or \"AE\"")
	}
	year := time.Now().Year()
	if n, ok := inputNumber(input, "year"); ok {
		year = int(n)
	}
	if year < 1975 || year > 2075 {
		return nil, errors.New("year must be between 1975 and 2075")
	}
	var holidays []struct {
		Date      string   `json:"date"`
		LocalName string   `json:"localName"`
		Name      string   `json:"name"`
		Global    bool     `json:"global"`
		Counties  []string `json:"counties"`
		Types     []string `json:"types"`
	}
	endpoint := holidaysEndpoint + strconv.Itoa(year) + "/" + country
	if err := getPublicJSON(ctx, e.httpClient(), endpoint, &holidays); err != nil {
		if errors.Is(err, errNotFound) {
			return nil, fmt.Errorf("no public holiday data for %s; search the web instead", country)
		}
		return nil, err
	}
	if len(holidays) == 0 {
		return nil, fmt.Errorf("no public holiday data for %s; search the web instead", country)
	}
	upcomingOnly := inputBool(input, "upcoming")
	today := time.Now().UTC().Format(time.DateOnly)
	out := make([]map[string]any, 0, len(holidays))
	for _, holiday := range holidays {
		if upcomingOnly && holiday.Date < today {
			continue
		}
		entry := map[string]any{"date": holiday.Date, "name": holiday.Name}
		if holiday.LocalName != holiday.Name {
			entry["local_name"] = holiday.LocalName
		}
		if parsed, err := time.Parse(time.DateOnly, holiday.Date); err == nil {
			entry["weekday"] = parsed.Weekday().String()
		}
		if !holiday.Global {
			entry["regional"] = true
		}
		out = append(out, entry)
	}
	return map[string]any{"country_code": country, "year": year, "holidays": out,
		"note": "Movable religious holidays can shift by a day once officially announced."}, nil
}
