package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

// convertTime shows one moment in several timezones, for "what time is 3pm
// Beirut in New York" and for scheduling across offices.
func (e *AssistantToolExecutor) convertTime(_ context.Context, input map[string]any) (map[string]any, error) {
	fromZone := inputText(input, "from", "from_timezone", "timezone", "zone")
	if fromZone == "" {
		fromZone = "UTC"
	}
	from, err := loadToolLocation(fromZone)
	if err != nil {
		return nil, err
	}
	moment := time.Now().In(from)
	if raw := inputText(input, "time", "datetime", "at"); raw != "" && !strings.EqualFold(raw, "now") {
		moment, err = parseLocalTime(raw, from)
		if err != nil {
			return nil, err
		}
	}
	targets := strings.FieldsFunc(inputText(input, "to", "to_timezone", "targets"), func(r rune) bool { return r == ',' || r == ' ' })
	if list, ok := input["to"].([]any); ok {
		for _, item := range list {
			if zone, ok := item.(string); ok {
				targets = append(targets, zone)
			}
		}
	}
	if len(targets) == 0 || len(targets) > 12 {
		return nil, errors.New("list one to twelve IANA timezones to convert to")
	}
	converted := make([]map[string]any, 0, len(targets))
	for _, zone := range targets {
		location, err := loadToolLocation(zone)
		if err != nil {
			return nil, err
		}
		local := moment.In(location)
		_, offset := local.Zone()
		converted = append(converted, map[string]any{
			"timezone": location.String(), "local_time": local.Format("Mon 2 Jan 2006, 15:04"),
			"utc_offset": formatUTCOffset(offset), "day_shift": dayShift(moment, local),
		})
	}
	_, offset := moment.Zone()
	return map[string]any{
		"from": map[string]any{"timezone": from.String(), "local_time": moment.Format("Mon 2 Jan 2006, 15:04"), "utc_offset": formatUTCOffset(offset)},
		"to":   converted,
	}, nil
}

func loadToolLocation(zone string) (*time.Location, error) {
	zone = strings.TrimSpace(zone)
	if len(zone) > 64 {
		return nil, errors.New("timezone is too long")
	}
	if strings.EqualFold(zone, "utc") || strings.EqualFold(zone, "gmt") {
		return time.UTC, nil
	}
	location, err := time.LoadLocation(zone)
	if err != nil {
		return nil, fmt.Errorf("%q is not an IANA timezone such as Asia/Beirut", zone)
	}
	return location, nil
}

func parseLocalTime(raw string, location *time.Location) (time.Time, error) {
	raw = strings.TrimSpace(raw)
	for _, layout := range []string{"2006-01-02 15:04", "2006-01-02T15:04", "2006-01-02T15:04:05", time.RFC3339} {
		if parsed, err := time.ParseInLocation(layout, raw, location); err == nil {
			return parsed, nil
		}
	}
	today := time.Now().In(location)
	for _, layout := range []string{"15:04", "3:04pm", "3pm", "3:04 pm", "3 pm"} {
		if parsed, err := time.ParseInLocation(layout, strings.ToLower(raw), location); err == nil {
			return time.Date(today.Year(), today.Month(), today.Day(), parsed.Hour(), parsed.Minute(), 0, 0, location), nil
		}
	}
	return time.Time{}, fmt.Errorf("%q is not a time; use \"2026-10-11 15:00\" or \"15:00\"", raw)
}

func formatUTCOffset(seconds int) string {
	sign := "+"
	if seconds < 0 {
		sign, seconds = "-", -seconds
	}
	return fmt.Sprintf("UTC%s%02d:%02d", sign, seconds/3600, seconds%3600/60)
}

func dayShift(source, target time.Time) string {
	a := time.Date(source.Year(), source.Month(), source.Day(), 0, 0, 0, 0, time.UTC)
	b := time.Date(target.Year(), target.Month(), target.Day(), 0, 0, 0, 0, time.UTC)
	switch {
	case b.After(a):
		return "next day"
	case b.Before(a):
		return "previous day"
	default:
		return "same day"
	}
}
