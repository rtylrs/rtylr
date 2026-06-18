package booking_appointments

import (
	"strings"
	"testing"
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
