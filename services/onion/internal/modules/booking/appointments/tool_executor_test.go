package booking_appointments

import (
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/booking/appointments/internal/domain"
)

func TestParseCreateAppointmentToolInput(t *testing.T) {
	parsed, err := parseCreateAppointmentToolInput(map[string]any{
		"service_id":     "42",
		"resource_id":    7,
		"staff_id":       9,
		"customer_name":  "Maya Haddad",
		"customer_email": "maya@example.com",
		"customer_phone": "+96170000000",
		"start_time":     "2026-07-15T10:00:00Z",
		"end_time":       "2026-07-15T10:45:00Z",
		"notes":          "Initial consultation",
	})
	if err != nil {
		t.Fatalf("parseCreateAppointmentToolInput returned error: %v", err)
	}
	if parsed.ServiceID != 42 {
		t.Fatalf("ServiceID = %d, want 42", parsed.ServiceID)
	}
	if parsed.ResourceID != 7 {
		t.Fatalf("ResourceID = %d, want 7", parsed.ResourceID)
	}
	if parsed.StaffID != 9 {
		t.Fatalf("StaffID = %d, want 9", parsed.StaffID)
	}
	if parsed.CustomerName != "Maya Haddad" {
		t.Fatalf("CustomerName = %q, want Maya Haddad", parsed.CustomerName)
	}
	if parsed.Status != "confirmed" {
		t.Fatalf("Status = %q, want confirmed", parsed.Status)
	}
	if parsed.StartTime != "2026-07-15T10:00:00Z" {
		t.Fatalf("StartTime = %q, want 2026-07-15T10:00:00Z", parsed.StartTime)
	}
	if parsed.EndTime != "2026-07-15T10:45:00Z" {
		t.Fatalf("EndTime = %q, want 2026-07-15T10:45:00Z", parsed.EndTime)
	}
}

func TestParseCreateAppointmentToolInputUsesDuration(t *testing.T) {
	parsed, err := parseCreateAppointmentToolInput(map[string]any{
		"service_id":       42,
		"client_name":      "Maya Haddad",
		"client_email":     "maya@example.com",
		"phone":            "+96170000000",
		"start_time":       "2026-07-15T10:00:00Z",
		"duration_minutes": "30",
		"status":           "in progress",
	})
	if err != nil {
		t.Fatalf("parseCreateAppointmentToolInput returned error: %v", err)
	}
	if parsed.CustomerEmail != "maya@example.com" {
		t.Fatalf("CustomerEmail = %q, want maya@example.com", parsed.CustomerEmail)
	}
	if parsed.CustomerPhone != "+96170000000" {
		t.Fatalf("CustomerPhone = %q, want +96170000000", parsed.CustomerPhone)
	}
	if parsed.EndTime != "2026-07-15T10:30:00Z" {
		t.Fatalf("EndTime = %q, want 2026-07-15T10:30:00Z", parsed.EndTime)
	}
	if parsed.Status != "in_progress" {
		t.Fatalf("Status = %q, want in_progress", parsed.Status)
	}
}

func TestParseCreateAppointmentToolInputRejectsMissingService(t *testing.T) {
	_, err := parseCreateAppointmentToolInput(map[string]any{
		"customer_name": "Maya Haddad",
		"start_time":    "2026-07-15T10:00:00Z",
		"end_time":      "2026-07-15T10:30:00Z",
	})
	if err == nil || !strings.Contains(err.Error(), "service_id") {
		t.Fatalf("error = %v, want service_id validation", err)
	}
}

func TestParseCreateAppointmentToolInputRejectsMissingCustomer(t *testing.T) {
	_, err := parseCreateAppointmentToolInput(map[string]any{
		"service_id": 42,
		"start_time": "2026-07-15T10:00:00Z",
		"end_time":   "2026-07-15T10:30:00Z",
	})
	if err == nil || !strings.Contains(err.Error(), "customer_name") {
		t.Fatalf("error = %v, want customer validation", err)
	}
}

func TestParseCreateAppointmentToolInputRejectsEndBeforeStart(t *testing.T) {
	_, err := parseCreateAppointmentToolInput(map[string]any{
		"service_id":    42,
		"customer_name": "Maya Haddad",
		"start_time":    "2026-07-15T10:00:00Z",
		"end_time":      "2026-07-15T09:30:00Z",
	})
	if err == nil || !strings.Contains(err.Error(), "before") {
		t.Fatalf("error = %v, want time-order validation", err)
	}
}

func TestParseCreateAppointmentToolInputRejectsBadStatus(t *testing.T) {
	_, err := parseCreateAppointmentToolInput(map[string]any{
		"service_id":    42,
		"customer_name": "Maya Haddad",
		"start_time":    "2026-07-15T10:00:00Z",
		"end_time":      "2026-07-15T10:30:00Z",
		"status":        "ready",
	})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestParseListAppointmentsToolInput(t *testing.T) {
	parsed, err := parseListAppointmentsToolInput(map[string]any{
		"status": "in progress",
		"today":  "yes",
	})
	if err != nil {
		t.Fatalf("parseListAppointmentsToolInput returned error: %v", err)
	}
	if parsed.Status != "in_progress" {
		t.Fatalf("Status = %q, want in_progress", parsed.Status)
	}
	if !parsed.Today {
		t.Fatal("Today = false, want true")
	}
}

func TestParseListAppointmentsToolInputRejectsBadStatus(t *testing.T) {
	_, err := parseListAppointmentsToolInput(map[string]any{"status": "ready"})
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestParseGetAppointmentToolInputAcceptsAliases(t *testing.T) {
	for _, input := range []map[string]any{
		{"appointment_id": "17"},
		{"appointment": "#17"},
		{"id": "appointment-17"},
		{"appointment_id": "appt-17"},
		{"appointment_id": 17},
	} {
		got, err := parseGetAppointmentToolInput(input)
		if err != nil {
			t.Fatalf("parseGetAppointmentToolInput(%v) returned error: %v", input, err)
		}
		if got != 17 {
			t.Fatalf("parseGetAppointmentToolInput(%v) = %d, want 17", input, got)
		}
	}
}

func TestParseGetAppointmentToolInputRejectsMissingAppointment(t *testing.T) {
	_, err := parseGetAppointmentToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "appointment_id is required") {
		t.Fatalf("error = %v, want missing appointment_id validation", err)
	}
}

func TestAppointmentToolPayloadFormatsFields(t *testing.T) {
	startTime := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	endTime := time.Date(2026, 7, 15, 10, 45, 0, 0, time.UTC)
	createdAt := time.Date(2026, 6, 18, 13, 30, 0, 0, time.FixedZone("EEST", 3*60*60))
	payload := appointmentToolPayload(domain.Appointment{
		ID:            11,
		ServiceID:     42,
		ResourceID:    7,
		StaffID:       9,
		CustomerName:  "Maya Haddad",
		CustomerEmail: "maya@example.com",
		StartTime:     startTime,
		EndTime:       endTime,
		Status:        "confirmed",
		CreatedAt:     createdAt,
	})

	if payload["id"] != int64(11) {
		t.Fatalf("id = %v", payload["id"])
	}
	if payload["customer_name"] != "Maya Haddad" {
		t.Fatalf("customer_name = %v", payload["customer_name"])
	}
	if payload["start_time"] != "2026-07-15T10:00:00Z" {
		t.Fatalf("start_time = %v", payload["start_time"])
	}
	if payload["created_at"] != "2026-06-18T10:30:00Z" {
		t.Fatalf("created_at = %v", payload["created_at"])
	}
}
