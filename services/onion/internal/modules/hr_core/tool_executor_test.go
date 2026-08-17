package hr_core

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
)

func TestParseListEmployeesToolInput(t *testing.T) {
	parsed, err := parseListEmployeesToolInput(map[string]any{
		"limit": json.Number("10"),
		"page":  "3",
	})
	if err != nil {
		t.Fatalf("parseListEmployeesToolInput: %v", err)
	}
	if parsed.Limit != 10 {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
	if parsed.Offset != 20 {
		t.Fatalf("Offset = %d", parsed.Offset)
	}
	if parsed.Page != 3 {
		t.Fatalf("Page = %d", parsed.Page)
	}
}

func TestParseListEmployeesToolInputOffsetOverridesPage(t *testing.T) {
	parsed, err := parseListEmployeesToolInput(map[string]any{
		"limit":  10,
		"page":   3,
		"offset": 5,
	})
	if err != nil {
		t.Fatalf("parseListEmployeesToolInput: %v", err)
	}
	if parsed.Offset != 5 {
		t.Fatalf("Offset = %d", parsed.Offset)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d", parsed.Page)
	}
}

func TestParseListEmployeesToolInputDefaultsAndCapsLimit(t *testing.T) {
	parsed, err := parseListEmployeesToolInput(map[string]any{
		"limit": 500,
	})
	if err != nil {
		t.Fatalf("parseListEmployeesToolInput: %v", err)
	}
	if parsed.Limit != int32(maxListEmployeesToolLimit) {
		t.Fatalf("Limit = %d", parsed.Limit)
	}
	if parsed.Offset != 0 {
		t.Fatalf("Offset = %d", parsed.Offset)
	}
}

func TestParseListEmployeesToolInputRejectsInvalidLimit(t *testing.T) {
	_, err := parseListEmployeesToolInput(map[string]any{
		"limit": 1.5,
	})
	if err == nil {
		t.Fatal("expected invalid limit error")
	}
}

func TestParseGetEmployeeToolInputAcceptsAliases(t *testing.T) {
	for _, input := range []map[string]any{
		{"employee_id": "7"},
		{"employee": json.Number("7")},
		{"id": float64(7)},
	} {
		employeeID, err := parseGetEmployeeToolInput(input)
		if err != nil {
			t.Fatalf("parseGetEmployeeToolInput(%v): %v", input, err)
		}
		if employeeID != 7 {
			t.Fatalf("employeeID = %d, want 7", employeeID)
		}
	}
}

func TestParseGetEmployeeToolInputRejectsMissingEmployee(t *testing.T) {
	_, err := parseGetEmployeeToolInput(map[string]any{"name": "Maya Haddad"})
	if err == nil || !strings.Contains(err.Error(), "employee_id") {
		t.Fatalf("error = %v, want employee_id validation", err)
	}
}

func TestEmployeeToolPayload(t *testing.T) {
	payload := employeeToolPayload(domain.EmployeeRecord{
		ID:             7,
		FirstName:      "Maya",
		LastName:       "Haddad",
		Email:          "maya@example.com",
		EmploymentType: domain.EmploymentTypePartTime,
		Status:         "active",
		DepartmentID:   3,
		StartDate:      time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC),
	})
	if payload["full_name"] != "Maya Haddad" {
		t.Fatalf("full_name = %#v", payload["full_name"])
	}
	if payload["start_date"] != "2026-06-18" {
		t.Fatalf("start_date = %#v", payload["start_date"])
	}
	if payload["end_date"] != nil {
		t.Fatalf("end_date = %#v", payload["end_date"])
	}
}

func TestEmployeeDetailToolPayloadIncludesProfileFields(t *testing.T) {
	payload := employeeDetailToolPayload(domain.EmployeeRecord{
		ID:                    7,
		BusinessID:            99,
		UserID:                1001,
		FirstName:             "Maya",
		LastName:              "Haddad",
		Email:                 "maya@example.com",
		Phone:                 "+96170000000",
		EmploymentType:        domain.EmploymentTypeFullTime,
		Status:                domain.EmploymentStatusActive,
		DepartmentID:          3,
		LocationID:            4,
		WorkScheduleID:        5,
		SalaryCents:           150000,
		HourlyRateCents:       1200,
		OvertimeRateCents:     1800,
		StartDate:             time.Date(2026, 6, 18, 0, 0, 0, 0, time.UTC),
		Address:               "Beirut",
		EmergencyContactName:  "Rami Haddad",
		EmergencyContactPhone: "+96171111111",
		Notes:                 "Morning shift lead",
		CreatedAt:             time.Date(2026, 6, 18, 10, 0, 0, 0, time.UTC),
		UpdatedAt:             time.Date(2026, 6, 19, 11, 0, 0, 0, time.UTC),
	})

	if payload["full_name"] != "Maya Haddad" {
		t.Fatalf("full_name = %#v", payload["full_name"])
	}
	if payload["overtime_rate"] != 18.0 {
		t.Fatalf("overtime_rate = %#v", payload["overtime_rate"])
	}
	if payload["emergency_contact_name"] != "Rami Haddad" {
		t.Fatalf("emergency_contact_name = %#v", payload["emergency_contact_name"])
	}
	if payload["created_at"] != "2026-06-18T10:00:00Z" {
		t.Fatalf("created_at = %#v", payload["created_at"])
	}
	if _, ok := payload["business_id"]; ok {
		t.Fatalf("business_id should not be exposed in employee detail payload")
	}
	if _, ok := payload["user_id"]; ok {
		t.Fatalf("user_id should not be exposed in employee detail payload")
	}
}

func TestParseListHRDepartmentsToolInput(t *testing.T) {
	parsed, err := parseListHRDepartmentsToolInput(map[string]any{
		"location_id": json.Number("9"),
	})
	if err != nil {
		t.Fatalf("parseListHRDepartmentsToolInput: %v", err)
	}
	if parsed.LocationID != 9 {
		t.Fatalf("LocationID = %d", parsed.LocationID)
	}
}

func TestParseListHRDepartmentsToolInputRejectsInvalidLocation(t *testing.T) {
	_, err := parseListHRDepartmentsToolInput(map[string]any{
		"location_id": -1,
	})
	if err == nil || !strings.Contains(err.Error(), "location_id") {
		t.Fatalf("error = %v, want location_id validation", err)
	}
}

func TestHRReferenceToolPayloads(t *testing.T) {
	schedule := workScheduleToolPayload(domain.WorkScheduleRecord{
		ID:           12,
		Name:         "Default",
		ScheduleData: `{"days":["mon","tue"],"hours":8}`,
		IsDefault:    true,
	})
	scheduleData, ok := schedule["schedule_data"].(map[string]any)
	if !ok {
		t.Fatalf("schedule_data = %#v, want decoded object", schedule["schedule_data"])
	}
	if scheduleData["hours"] != float64(8) {
		t.Fatalf("schedule hours = %#v", scheduleData["hours"])
	}

	role := roleToolPayload(domain.RoleRecord{
		ID:          4,
		Name:        "Manager",
		Permissions: `{"view_team":true,"manage_roles":true,"disabled":false}`,
	})
	if !reflect.DeepEqual(role["permission_keys"], []string{"manage_roles", "view_team"}) {
		t.Fatalf("permission_keys = %#v", role["permission_keys"])
	}

	department := departmentToolPayload(domain.DepartmentRecord{
		ID:         6,
		Name:       "Operations",
		LocationID: 0,
	})
	if department["location_id"] != nil {
		t.Fatalf("location_id = %#v, want nil for zero id", department["location_id"])
	}
}

func TestHRSetupAndConfigToolPayloads(t *testing.T) {
	completedAt := time.Date(2026, 6, 18, 10, 0, 0, 0, time.FixedZone("EEST", 3*60*60))
	step := setupStepToolPayload(domain.SetupStepRecord{
		ID:          1,
		StepKey:     "company",
		StepLabel:   "Company profile",
		IsCompleted: true,
		CompletedAt: completedAt,
	})
	if step["completed_at"] != "2026-06-18T07:00:00Z" {
		t.Fatalf("completed_at = %#v", step["completed_at"])
	}

	config := hrConfigToolPayload(&domain.HRCompanyConfigRecord{
		ID:                   3,
		CompanyName:          "Suplab",
		FiscalYearStartMonth: 1,
		DefaultCurrency:      "USD",
		DefaultTimezone:      "Asia/Beirut",
		OvertimeEnabled:      true,
		OvertimeMultiplier:   1.5,
	})
	if config["company_name"] != "Suplab" {
		t.Fatalf("company_name = %#v", config["company_name"])
	}
	if config["default_work_schedule_id"] != nil {
		t.Fatalf("default_work_schedule_id = %#v, want nil for zero id", config["default_work_schedule_id"])
	}
}

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
