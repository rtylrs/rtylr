package booking_appointments

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/booking_appointments/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
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
