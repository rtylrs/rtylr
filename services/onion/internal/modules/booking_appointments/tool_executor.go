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
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	appointmentCacheScope     = "booking_appointments"
	createAppointmentToolName = "booking.create_appointment"
	getAppointmentToolName    = "booking.get_appointment"
	listAppointmentsToolName  = "booking.list_appointments"
)

type listAppointmentsToolInput struct {
	Status string
	Today  bool
}

var validAppointmentStatuses = map[string]struct{}{
	"pending":     {},
	"confirmed":   {},
	"in_progress": {},
	"completed":   {},
	"cancelled":   {},
	"no_show":     {},
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listAppointmentsToolName:
		output, err := m.executeListAppointmentsTool(ctx, input)
		return output, true, err
	case getAppointmentToolName:
		output, err := m.executeGetAppointmentTool(ctx, input)
		return output, true, err
	case createAppointmentToolName:
		output, err := m.executeCreateAppointmentTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListAppointmentsTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListAppointmentsToolInput(input)
	if err != nil {
		return nil, err
	}

	var appointments []domain.Appointment
	if parsed.Today {
		appointments, err = m.service.ListToday(ctx, businessID)
	} else {
		appointments, err = m.service.List(ctx, businessID)
	}
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(appointments))
	for _, appointment := range appointments {
		if parsed.Status != "" && appointment.Status != parsed.Status {
			continue
		}
		payloads = append(payloads, appointmentToolPayload(appointment))
	}

	return map[string]any{
		"appointments": payloads,
		"returned":     len(payloads),
		"filters": map[string]any{
			"status": parsed.Status,
			"today":  parsed.Today,
		},
	}, nil
}

func (m *Module) executeGetAppointmentTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("authentication is required")
	}

	appointmentID, err := parseGetAppointmentToolInput(input)
	if err != nil {
		return nil, err
	}

	appointment, err := m.service.GetByID(ctx, businessID, appointmentID)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"appointment": appointmentToolPayload(appointment),
	}, nil
}

func (m *Module) executeCreateAppointmentTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseCreateAppointmentToolInput(input)
	if err != nil {
		return nil, err
	}

	id, err := m.service.Create(ctx, businessID, parsed)
	if err != nil {
		return nil, err
	}
	invalidateAppointmentToolCache(ctx, businessID)

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

func parseGetAppointmentToolInput(input map[string]any) (int64, error) {
	appointmentID, err := appointmentIDToolValue(firstAppointmentToolValue(input, "appointment_id", "appointment", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid appointment_id: %w", err)
	}
	if appointmentID <= 0 {
		return 0, errors.New("appointment_id is required")
	}
	return appointmentID, nil
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
