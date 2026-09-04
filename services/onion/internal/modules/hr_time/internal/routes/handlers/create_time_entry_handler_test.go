package handlers

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
