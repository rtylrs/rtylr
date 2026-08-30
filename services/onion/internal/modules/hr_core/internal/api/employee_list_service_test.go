package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_core/internal/domain"
)

func TestRegisterEmployeeListAPI_RegistersContractPath(t *testing.T) {
	mux := http.NewServeMux()
	RegisterEmployeeListAPI(mux, NewEmployeeListAPI(nil))

	req := httptest.NewRequest(http.MethodGet, employeeListPath, nil)
	res := httptest.NewRecorder()
	mux.ServeHTTP(res, req)

	if res.Code != http.StatusForbidden {
		t.Fatalf("contract list route status = %d, want %d; body=%s", res.Code, http.StatusForbidden, res.Body.String())
	}
	if strings.Contains(res.Body.String(), "404") {
		t.Fatalf("contract list route was not registered: %s", res.Body.String())
	}
}

func TestParseEmployeeListRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, employeeListPath+"?department_id=12&location_id=9&status=active&limit=50&offset=10", nil)
	parsed, err := parseEmployeeListRequest(req)
	if err != nil {
		t.Fatalf("parseEmployeeListRequest: %v", err)
	}
	if parsed.GetDepartmentId() != 12 || parsed.GetLocationId() != 9 || parsed.GetStatus() != "active" || parsed.GetLimit() != 50 || parsed.GetOffset() != 10 {
		t.Fatalf("unexpected parsed request: %+v", parsed)
	}
}

func TestParseEmployeeListRequestRejectsInvalidFilter(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, employeeListPath+"?department_id=nope", nil)
	if _, err := parseEmployeeListRequest(req); err == nil {
		t.Fatal("expected invalid department_id to fail")
	}
}

func TestEmployeeToPayload(t *testing.T) {
	created := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	employee := domain.EmployeeRecord{
		ID:              42,
		OwnerID:         7,
		UserID:          21,
		DepartmentID:    12,
		LocationID:      9,
		WorkScheduleID:  5,
		EmploymentType:  "part_time",
		Status:          "active",
		StartDate:       created,
		EndDate:         created.AddDate(0, 1, 0),
		SalaryCents:     125000,
		HourlyRateCents: 2500,
		CreatedAt:       created,
		UpdatedAt:       created,
	}
	payload := employeeToPayload(employee)
	if payload.GetId() != 42 || payload.GetUserId() != 21 || payload.GetWorkScheduleId() != 5 {
		t.Fatalf("identity fields not mapped: %+v", payload)
	}
	if payload.GetContractType() != "part_time" || payload.GetHireDate() != "2026-08-30" {
		t.Fatalf("employment fields not mapped: %+v", payload)
	}
	if payload.HourlyRate == nil || *payload.HourlyRate != 25 || payload.MonthlySalary == nil || *payload.MonthlySalary != 1250 {
		t.Fatalf("compensation fields not mapped: %+v", payload)
	}
	if payload.TerminationDate == nil || *payload.TerminationDate != "2026-09-30" {
		t.Fatalf("termination date not mapped: %+v", payload)
	}
}
