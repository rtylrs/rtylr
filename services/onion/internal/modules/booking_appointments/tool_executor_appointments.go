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
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

func parseListAppointmentsToolInput(input map[string]any) (listAppointmentsToolInput, error) {
	status := normalizeAppointmentStatusFilter(input["status"])
	if strings.TrimSpace(appointmentToolString(input["status"])) != "" && status == "" {
		return listAppointmentsToolInput{}, errors.New("status must be one of pending, confirmed, in_progress, completed, cancelled, or no_show")
	}

	today, err := appointmentToolBool(firstAppointmentToolValue(input, "today", "is_today"), false)
	if err != nil {
		return listAppointmentsToolInput{}, fmt.Errorf("invalid today: %w", err)
	}

	return listAppointmentsToolInput{
		Status: status,
		Today:  today,
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

func normalizeAppointmentStatusFilter(value any) string {
	status := strings.ToLower(strings.TrimSpace(appointmentToolString(value)))
	if status == "" {
		return ""
	}
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	if _, ok := validAppointmentStatuses[status]; ok {
		return status
	}
	return ""
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

func appointmentToolPayload(appointment domain.Appointment) map[string]any {
	return map[string]any{
		"id":             appointment.ID,
		"service_id":     appointment.ServiceID,
		"resource_id":    appointment.ResourceID,
		"staff_id":       appointment.StaffID,
		"customer_name":  appointment.CustomerName,
		"customer_email": appointment.CustomerEmail,
		"customer_phone": appointment.CustomerPhone,
		"start_time":     appointmentToolFormatTime(appointment.StartTime),
		"end_time":       appointmentToolFormatTime(appointment.EndTime),
		"status":         appointment.Status,
		"notes":          appointment.Notes,
		"created_at":     appointmentToolFormatTime(appointment.CreatedAt),
		"updated_at":     appointmentToolFormatTime(appointment.UpdatedAt),
	}
}

func firstAppointmentToolValue(input map[string]any, keys ...string) any {
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

func appointmentIDToolValue(value any) (int64, error) {
	if raw, ok := value.(string); ok {
		text := strings.TrimSpace(raw)
		if text == "" {
			return 0, nil
		}
		normalized := strings.ToLower(text)
		normalized = strings.TrimPrefix(normalized, "#")
		normalized = strings.TrimPrefix(normalized, "appointment-")
		normalized = strings.TrimPrefix(normalized, "appt-")
		if normalized == "" {
			return 0, nil
		}
		return strconv.ParseInt(normalized, 10, 64)
	}
	return appointmentToolInt64(value)
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

func appointmentToolBool(value any, defaultValue bool) (bool, error) {
	switch v := value.(type) {
	case nil:
		return defaultValue, nil
	case bool:
		return v, nil
	case json.Number:
		n, err := strconv.ParseInt(v.String(), 10, 64)
		if err != nil {
			return false, err
		}
		switch n {
		case 0:
			return false, nil
		case 1:
			return true, nil
		default:
			return false, errors.New("must be true or false")
		}
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		if raw == "" {
			return defaultValue, nil
		}
		switch raw {
		case "true", "t", "yes", "y", "1", "today":
			return true, nil
		case "false", "f", "no", "n", "0":
			return false, nil
		default:
			return false, errors.New("must be true or false")
		}
	default:
		return false, fmt.Errorf("unsupported value %T", value)
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

func appointmentToolFormatTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.UTC().Format(time.RFC3339)
}

func invalidateAppointmentToolCache(ctx context.Context, ownerID int64) {
	cacher.InvalidateScope(ctx, cacheKeys.ScopedPrefix(appointmentCacheScope, ownerID))
}
