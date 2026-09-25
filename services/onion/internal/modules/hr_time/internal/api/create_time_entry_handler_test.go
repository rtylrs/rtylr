package api

import (
	"encoding/json"
	"testing"
)

func TestCreateTimeEntryRequestUnmarshalJSONAcceptsNumericAndStringIDs(t *testing.T) {
	var req CreateTimeEntryRequest
	if err := json.Unmarshal([]byte(`{
		"employee_id": "4101",
		"project_id": 12,
		"task_id": "34",
		"category_id": 56,
		"date": "2026-09-04",
		"duration_minutes": 30,
		"description": "Planning",
		"source": "manual"
	}`), &req); err != nil {
		t.Fatalf("unmarshal request: %v", err)
	}

	if req.EmployeeID != 4101 {
		t.Errorf("employee_id = %d, want 4101", req.EmployeeID)
	}
	if req.ProjectID != 12 {
		t.Errorf("project_id = %d, want 12", req.ProjectID)
	}
	if req.TaskID != 34 {
		t.Errorf("task_id = %d, want 34", req.TaskID)
	}
	if req.CategoryID != 56 {
		t.Errorf("category_id = %d, want 56", req.CategoryID)
	}
	if req.DurationMinutes != 30 || req.Description != "Planning" || req.Source != "manual" {
		t.Errorf("non-ID fields were not decoded: %+v", req)
	}
}

func TestCreateTimeEntryRequestUnmarshalJSONRejectsInvalidID(t *testing.T) {
	var req CreateTimeEntryRequest
	if err := json.Unmarshal([]byte(`{"employee_id":"not-a-number"}`), &req); err == nil {
		t.Fatal("expected invalid employee_id to fail decoding")
	}
}

func TestCreateTimeEntryValidateRejectsNonPositiveEmployeeID(t *testing.T) {
	handler := &CreateTimeEntryHandler{}
	for _, employeeID := range []int64{0, -1} {
		err := handler.Validate(t.Context(), &CreateTimeEntryRequest{
			EmployeeID:      employeeID,
			Date:            "2026-11-02",
			DurationMinutes: 30,
		})
		if err == nil {
			t.Fatalf("employee_id %d unexpectedly passed validation", employeeID)
		}
	}
}

func TestUpdateTimeEntryRequestTracksOmittedFields(t *testing.T) {
	var req UpdateTimeEntryRequest
	if err := json.Unmarshal([]byte(`{"description":"updated","is_billable":false}`), &req); err != nil {
		t.Fatalf("unmarshal update request: %v", err)
	}
	if req.Description == nil || *req.Description != "updated" {
		t.Fatalf("description = %v, want updated", req.Description)
	}
	if req.IsBillable == nil || *req.IsBillable {
		t.Fatalf("is_billable = %v, want present false", req.IsBillable)
	}
	if req.EmployeeID != nil || req.Date != nil || req.DurationMinutes != nil || req.ProjectID != nil {
		t.Fatalf("omitted fields were materialized: %+v", req)
	}
}
