package api

import (
	"encoding/json"
	"testing"

	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
	flowtimev1 "github.com/voxire/voxie/gen/go/flow/time/v1"
)

func TestCreateTaskRequestUnmarshalsOptionalInt64Strings(t *testing.T) {
	var req flowtaskv1.CreateTaskRequest
	if err := json.Unmarshal([]byte(`{
		"project_id":"23",
		"workspace_id":"13",
		"title":"meow",
		"description":"-",
		"priority":"medium",
		"assignee_id":"10",
		"due_date":"-",
		"start_date":"-",
		"estimated_hours":0.5,
		"labels":["382"]
	}`), &req); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if req.GetAssigneeId() != 10 || len(req.GetLabels()) != 1 || req.GetLabels()[0] != 382 {
		t.Fatalf("optional int64 fields were not decoded: assignee=%d labels=%v", req.GetAssigneeId(), req.GetLabels())
	}
}

func TestCreateWorkDayLineRequestUnmarshalsOptionalInt64Strings(t *testing.T) {
	var req flowtimev1.CreateWorkDayLineRequest
	if err := json.Unmarshal([]byte(`{
		"work_day_id":"7",
		"client_key":"401774b4-9508-4de4-9887-c4d713c6ba28",
		"description":"-",
		"duration_minutes":60,
		"project_id":"23",
		"task_id":"1602",
		"work_mode":"onsite",
		"is_billable":true
	}`), &req); err != nil {
		t.Fatalf("UnmarshalJSON: %v", err)
	}
	if err := req.Validate(); err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if req.GetProjectId() != 23 || req.GetTaskId() != 1602 {
		t.Fatalf("optional int64 fields were not decoded: project=%d task=%d", req.GetProjectId(), req.GetTaskId())
	}
}
