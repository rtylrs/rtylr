package booking_appointments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/booking_appointments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

const appointmentCacheScope = "booking_appointments"

var validAppointmentStatuses = map[string]struct{}{
	"pending":     {},
	"confirmed":   {},
	"in_progress": {},
	"completed":   {},
	"cancelled":   {},
	"no_show":     {},
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != "booking.create_appointment" {
		return nil, false, nil
	}
	output, err := m.executeCreateAppointmentTool(ctx, input)
	return output, true, err
}

func (m *Module) executeCreateAppointmentTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateAppointmentToolInput(input)
	if err != nil {
		return nil, err
	}

	id, err := m.service.Create(ctx, authUser.UserID, parsed)
	if err != nil {
		return nil, err
	}
	invalidateAppointmentToolCache(ctx, authUser.UserID)

	return map[string]any{
		"appointment": map[string]any{
			"id":             id,
			"service_id":     parsed.ServiceID,
			"resource_id":    parsed.ResourceID,
			"staff_id":       parsed.StaffID,
			"customer_name":  parsed.CustomerName,
			"customer_email": parsed.CustomerEmail,
			"customer_phone": parsed.CustomerPhone,
			"start_time":     parsed.StartTime,
			"end_time":       parsed.EndTime,
			"status":         parsed.Status,
			"notes":          parsed.Notes,
		},
	}, nil
}

func parseCreateAppointmentToolInput(input map[string]any) (domain.AppointmentInput, error) {
	serviceID, err := appointmentToolInt64(input["service_id"])
	if err != nil {
		return domain.AppointmentInput{}, fmt.Errorf("invalid service_id: %w", err)
	}
	if serviceID <= 0 {
		return domain.AppointmentInput{}, errors.New("service_id is required")
	}

	resourceID, err := appointmentToolNonNegativeInt64(input["resource_id"])
	if err != nil {
		return domain.AppointmentInput{}, fmt.Errorf("invalid resource_id: %w", err)
	}
	staffID, err := appointmentToolNonNegativeInt64(input["staff_id"])
	if err != nil {
		return domain.AppointmentInput{}, fmt.Errorf("invalid staff_id: %w", err)
	}

	customerName := strings.TrimSpace(appointmentToolString(input["customer_name"]))
	if customerName == "" {
		customerName = strings.TrimSpace(appointmentToolString(input["client_name"]))
	}
	if customerName == "" {
		return domain.AppointmentInput{}, errors.New("customer_name is required")
	}

	startTime, err := appointmentToolTime(input["start_time"])
	if err != nil {
		return domain.AppointmentInput{}, fmt.Errorf("invalid start_time: %w", err)
	}
	if startTime.IsZero() {
		return domain.AppointmentInput{}, errors.New("start_time is required")
	}

	endTime, err := appointmentToolEndTime(input, startTime)
	if err != nil {
		return domain.AppointmentInput{}, err
	}
	if !startTime.Before(endTime) {
		return domain.AppointmentInput{}, errors.New("start_time must be before end_time")
	}

	status := normalizeAppointmentStatus(input["status"])
	if status == "" {
		return domain.AppointmentInput{}, errors.New("status must be one of pending, confirmed, in_progress, completed, cancelled, or no_show")
	}

	customerEmail := strings.TrimSpace(appointmentToolString(input["customer_email"]))
	if customerEmail == "" {
		customerEmail = strings.TrimSpace(appointmentToolString(input["client_email"]))
	}
	customerPhone := strings.TrimSpace(appointmentToolString(input["customer_phone"]))
	if customerPhone == "" {
		customerPhone = strings.TrimSpace(appointmentToolString(input["phone"]))
	}

	return domain.AppointmentInput{
		ServiceID:     serviceID,
		ResourceID:    resourceID,
		StaffID:       staffID,
		CustomerName:  customerName,
		CustomerEmail: customerEmail,
		CustomerPhone: customerPhone,
		StartTime:     startTime.Format(time.RFC3339),
		EndTime:       endTime.Format(time.RFC3339),
		Status:        status,
		Notes:         strings.TrimSpace(appointmentToolString(input["notes"])),
	}, nil
}

func appointmentToolEndTime(input map[string]any, startTime time.Time) (time.Time, error) {
	endTime, err := appointmentToolTime(input["end_time"])
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid end_time: %w", err)
	}
	if !endTime.IsZero() {
		return endTime, nil
	}

	durationValue := input["duration_minutes"]
	if durationValue == nil {
		durationValue = input["duration"]
	}
	durationMinutes, err := appointmentToolFloat64(durationValue)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid duration_minutes: %w", err)
	}
	if durationMinutes <= 0 {
		return time.Time{}, errors.New("end_time or positive duration_minutes is required")
	}
	return startTime.Add(time.Duration(durationMinutes * float64(time.Minute))), nil
}

func normalizeAppointmentStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(appointmentToolString(value)))
	if status == "" {
		return "confirmed"
	}
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	if _, ok := validAppointmentStatuses[status]; ok {
		return status
	}
	return ""
}

func appointmentToolTime(value any) (time.Time, error) {
	switch v := value.(type) {
	case nil:
		return time.Time{}, nil
	case time.Time:
		return v, nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return time.Time{}, nil
		}
		for _, layout := range []string{time.RFC3339, time.RFC3339Nano, "2006-01-02T15:04:05Z07:00"} {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return parsed, nil
			}
		}
		return time.Time{}, errors.New("must be RFC3339")
	default:
		return time.Time{}, fmt.Errorf("unsupported value %T", value)
	}
}

func appointmentToolNonNegativeInt64(value any) (int64, error) {
	out, err := appointmentToolInt64(value)
	if err != nil {
		return 0, err
	}
	if out < 0 {
		return 0, errors.New("must be non-negative")
	}
	return out, nil
}

func appointmentToolInt64(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		if math.Trunc(v) != v {
			return 0, errors.New("must be an integer")
		}
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseInt(raw, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func appointmentToolFloat64(value any) (float64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case float64:
		return v, nil
	case float32:
		return float64(v), nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case json.Number:
		return v.Float64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return 0, nil
		}
		return strconv.ParseFloat(raw, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func appointmentToolString(value any) string {
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

func invalidateAppointmentToolCache(ctx context.Context, ownerID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(appointmentCacheScope, ownerID))
	}
}
