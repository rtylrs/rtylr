package hr_core

import (
	"strings"
	"testing"
)

func TestParseCreateEmployeeToolInput(t *testing.T) {
	parsed, err := parseCreateEmployeeToolInput(map[string]any{
		"name":            "Maya Haddad",
		"email":           "maya@example.com",
		"phone":           "+96170000000",
		"employment_type": "part time",
		"salary":          "1200.50",
		"hourly_rate":     12,
		"start_date":      "2026-06-18",
		"notes":           "Morning shift",
	})
	if err != nil {
		t.Fatalf("parseCreateEmployeeToolInput returned error: %v", err)
	}
	if parsed.FirstName != "Maya" || parsed.LastName != "Haddad" {
		t.Fatalf("name = %q %q, want Maya Haddad", parsed.FirstName, parsed.LastName)
	}
	if parsed.Email != "maya@example.com" {
		t.Fatalf("Email = %q, want maya@example.com", parsed.Email)
	}
	if parsed.EmploymentType != "part_time" {
		t.Fatalf("EmploymentType = %q, want part_time", parsed.EmploymentType)
	}
	if parsed.Salary != 1200.50 {
		t.Fatalf("Salary = %v, want 1200.50", parsed.Salary)
	}
	if got := parsed.StartDate.Format("2006-01-02"); got != "2026-06-18" {
		t.Fatalf("StartDate = %q, want 2026-06-18", got)
	}
}

func TestParseCreateEmployeeToolInputRejectsMissingEmail(t *testing.T) {
	_, err := parseCreateEmployeeToolInput(map[string]any{
		"first_name": "Maya",
		"last_name":  "Haddad",
	})
	if err == nil || !strings.Contains(err.Error(), "email") {
		t.Fatalf("error = %v, want email validation", err)
	}
}

func TestParseCreateEmployeeToolInputRejectsNegativeSalary(t *testing.T) {
	_, err := parseCreateEmployeeToolInput(map[string]any{
		"first_name": "Maya",
		"last_name":  "Haddad",
		"email":      "maya@example.com",
		"salary":     -1,
	})
	if err == nil || !strings.Contains(err.Error(), "invalid salary") {
		t.Fatalf("error = %v, want salary validation", err)
	}
}

func TestParseCreateEmployeeToolInputRejectsBadEmploymentType(t *testing.T) {
	_, err := parseCreateEmployeeToolInput(map[string]any{
		"first_name":      "Maya",
		"last_name":       "Haddad",
		"email":           "maya@example.com",
		"employment_type": "temporary",
	})
	if err == nil || !strings.Contains(err.Error(), "employment_type") {
		t.Fatalf("error = %v, want employment type validation", err)
	}
}
