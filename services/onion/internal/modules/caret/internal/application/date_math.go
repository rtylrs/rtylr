package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

func init() {
	registerGeneralTool("utilities.date_math", (*AssistantToolExecutor).dateMath)
}

// dateMath answers the date questions models get wrong: the span between two
// dates (with working days), a date plus or minus an offset, and weekdays.
func (e *AssistantToolExecutor) dateMath(ctx context.Context, input map[string]any) (map[string]any, error) {
	today := todayIn(ctx)
	weekend, err := weekendDays(inputText(input, "weekend"))
	if err != nil {
		return nil, err
	}
	start, err := parseToolDateFrom(inputText(input, "from", "start", "date", "start_date"), today)
	if err != nil {
		return nil, err
	}
	if end := inputText(input, "to", "end", "end_date"); end != "" {
		finish, err := parseToolDateFrom(end, today)
		if err != nil {
			return nil, err
		}
		return dateSpan(start, finish, weekend), nil
	}
	days, _ := inputNumber(input, "days", "add_days")
	weeks, _ := inputNumber(input, "weeks")
	months, _ := inputNumber(input, "months")
	years, _ := inputNumber(input, "years")
	workdays, hasWorkdays := inputNumber(input, "business_days", "working_days", "workdays")
	result := addMonthsClamped(start, 12*int(years)+int(months)).AddDate(0, 0, int(days)+7*int(weeks))
	if hasWorkdays {
		result = addWorkingDays(result, int(workdays), weekend)
	}
	return map[string]any{
		"start": describeDate(start), "result": describeDate(result),
		"days_between": int(result.Sub(start).Hours() / 24),
	}, nil
}

func dateSpan(start, end time.Time, weekend map[time.Weekday]bool) map[string]any {
	sign := 1
	if end.Before(start) {
		start, end, sign = end, start, -1
	}
	days := int(end.Sub(start).Hours() / 24)
	working := 0
	for day := start; day.Before(end); day = day.AddDate(0, 0, 1) {
		if !weekend[day.Weekday()] {
			working++
		}
	}
	years, months, rest := calendarDifference(start, end)
	return map[string]any{
		"from": describeDate(start), "to": describeDate(end),
		"days": sign * days, "weeks": fmt.Sprintf("%d weeks and %d days", days/7, days%7),
		"calendar":     fmt.Sprintf("%d years, %d months, %d days", years, months, rest),
		"working_days": sign * working,
		"note":         "Working days count from the start date up to (not including) the end date and skip weekends only, not public holidays.",
	}
}

func calendarDifference(start, end time.Time) (years, months, days int) {
	years = end.Year() - start.Year()
	months = int(end.Month()) - int(start.Month())
	days = end.Day() - start.Day()
	if days < 0 {
		months--
		days += time.Date(end.Year(), end.Month(), 0, 0, 0, 0, 0, time.UTC).Day()
	}
	if months < 0 {
		years--
		months += 12
	}
	return years, months, days
}

// addMonthsClamped moves by whole months and keeps the day inside the target
// month, so Jan 31 plus one month is Feb 28, not Mar 3.
func addMonthsClamped(date time.Time, months int) time.Time {
	first := time.Date(date.Year(), date.Month()+time.Month(months), 1, 0, 0, 0, 0, time.UTC)
	last := first.AddDate(0, 1, -1).Day()
	return time.Date(first.Year(), first.Month(), min(date.Day(), last), 0, 0, 0, 0, time.UTC)
}

func addWorkingDays(date time.Time, count int, weekend map[time.Weekday]bool) time.Time {
	step := 1
	if count < 0 {
		step, count = -1, -count
	}
	for count > 0 {
		date = date.AddDate(0, 0, step)
		if !weekend[date.Weekday()] {
			count--
		}
	}
	return date
}

func describeDate(date time.Time) map[string]any {
	_, week := date.ISOWeek()
	return map[string]any{"date": date.Format(time.DateOnly), "weekday": date.Weekday().String(), "iso_week": week, "day_of_year": date.YearDay()}
}

// parseToolDate reads YYYY-MM-DD, an RFC 3339 timestamp, or "today" (UTC).
func parseToolDate(raw string) (time.Time, error) {
	return parseToolDateFrom(raw, time.Now().UTC().Truncate(24*time.Hour))
}

// parseToolDateFrom resolves "today", "tomorrow" and "yesterday" against the
// given date (midnight UTC of the person's local day).
func parseToolDateFrom(raw string, today time.Time) (time.Time, error) {
	raw = strings.TrimSpace(strings.ToLower(raw))
	switch raw {
	case "", "today", "now":
		return today, nil
	case "tomorrow":
		return today.AddDate(0, 0, 1), nil
	case "yesterday":
		return today.AddDate(0, 0, -1), nil
	}
	for _, layout := range []string{time.DateOnly, time.RFC3339, "2006/01/02", "02/01/2006"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC), nil
		}
	}
	return time.Time{}, fmt.Errorf("%q is not a date; use YYYY-MM-DD", raw)
}

// weekendDays parses "saturday,sunday" or "fri-sat"; the default is
// Saturday and Sunday.
func weekendDays(raw string) (map[time.Weekday]bool, error) {
	if strings.TrimSpace(raw) == "" {
		return map[time.Weekday]bool{time.Saturday: true, time.Sunday: true}, nil
	}
	if strings.EqualFold(strings.TrimSpace(raw), "none") {
		return map[time.Weekday]bool{}, nil
	}
	out := map[time.Weekday]bool{}
	for _, part := range strings.FieldsFunc(strings.ToLower(raw), func(r rune) bool { return r == ',' || r == '-' || r == ' ' || r == '/' }) {
		found := false
		for day := time.Sunday; day <= time.Saturday; day++ {
			if strings.HasPrefix(strings.ToLower(day.String()), part) && len(part) >= 2 {
				out[day], found = true, true
			}
		}
		if !found {
			return nil, errors.New("weekend must list weekdays, e.g. \"friday,saturday\"")
		}
	}
	return out, nil
}

type userLocationKey struct{}

// WithUserLocation records the person's timezone for tools that resolve
// "today".
func WithUserLocation(ctx context.Context, location *time.Location) context.Context {
	return context.WithValue(ctx, userLocationKey{}, location)
}

// todayIn is the person's current date, as midnight UTC of that day.
func todayIn(ctx context.Context) time.Time {
	location, ok := ctx.Value(userLocationKey{}).(*time.Location)
	if !ok || location == nil {
		location = time.UTC
	}
	local := time.Now().In(location)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}
