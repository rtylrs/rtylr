package booking_calendar

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const getBookingCalendarToolName = "booking.get_calendar"

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != getBookingCalendarToolName {
		return nil, false, nil
	}

	businessID, err := bookingCalendarToolBusinessID(ctx)
	if err != nil {
		return nil, true, err
	}
	from, to, err := bookingCalendarToolRange(input)
	if err != nil {
		return nil, true, err
	}

	days, err := m.service.GetCalendar(ctx, businessID, from, to)
	if err != nil {
		return nil, true, err
	}
	sort.Slice(days, func(i, j int) bool {
		return days[i].Date < days[j].Date
	})

	totalAppointments := 0
	for _, day := range days {
		totalAppointments += len(day.Appointments)
	}

	return map[string]any{
		"days":               days,
		"returned":           len(days),
		"total_appointments": totalAppointments,
		"filters": map[string]any{
			"from": from,
			"to":   to,
		},
	}, true, nil
}

func bookingCalendarToolBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return 0, errors.New("authentication is required")
	}
	businessID, err := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if err != nil {
		return 0, errors.New("authentication is required")
	}
	return businessID, nil
}

func bookingCalendarToolRange(input map[string]any) (string, string, error) {
	from, err := bookingCalendarToolDate(firstBookingCalendarToolValue(input, "from", "date_from", "start_date"), "from")
	if err != nil {
		return "", "", err
	}
	to, err := bookingCalendarToolDate(firstBookingCalendarToolValue(input, "to", "date_to", "end_date"), "to")
	if err != nil {
		return "", "", err
	}
	if from == "" {
		from = time.Now().Format("2006-01-02")
	}
	start, err := time.Parse("2006-01-02", from)
	if err != nil {
		return "", "", errors.New("from must be in YYYY-MM-DD format")
	}
	if to == "" {
		to = start.AddDate(0, 0, 7).Format("2006-01-02")
	}
	end, err := time.Parse("2006-01-02", to)
	if err != nil {
		return "", "", errors.New("to must be in YYYY-MM-DD format")
	}
	if end.Before(start) {
		return "", "", errors.New("to must be on or after from")
	}
	if end.Sub(start) > 90*24*time.Hour {
		return "", "", errors.New("calendar range cannot exceed 90 days")
	}
	return from, to, nil
}

func firstBookingCalendarToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok || value == nil {
			continue
		}
		if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
			continue
		}
		return value
	}
	return nil
}

func bookingCalendarToolDate(value any, field string) (string, error) {
	date := strings.TrimSpace(bookingCalendarToolString(value))
	if date == "" {
		return "", nil
	}
	if strings.EqualFold(date, "today") {
		return time.Now().Format("2006-01-02"), nil
	}
	if _, err := time.Parse("2006-01-02", date); err != nil {
		return "", fmt.Errorf("%s must be in YYYY-MM-DD format", field)
	}
	return date, nil
}

func bookingCalendarToolString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}
