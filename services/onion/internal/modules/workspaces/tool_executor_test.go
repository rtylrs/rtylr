package workspaces

import (
	"strings"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

func TestParseListTasksToolInput(t *testing.T) {
	parsed, err := parseListTasksToolInput(map[string]any{
		"project":     "42",
		"status":      "in progress",
		"assignee":    "me",
		"priority":    "high",
		"active_only": "true",
		"compact":     "false",
		"page":        "2",
		"page_size":   100,
	}, 99)
	if err != nil {
		t.Fatalf("parseListTasksToolInput returned error: %v", err)
	}
	if parsed.ProjectID != 42 {
		t.Fatalf("ProjectID = %d, want 42", parsed.ProjectID)
	}
	if parsed.Status != "in_progress" {
		t.Fatalf("Status = %q, want in_progress", parsed.Status)
	}
	if parsed.AssigneeID == nil || *parsed.AssigneeID != 99 {
		t.Fatalf("AssigneeID = %v, want 99", parsed.AssigneeID)
	}
	if parsed.Priority != "high" {
		t.Fatalf("Priority = %q, want high", parsed.Priority)
	}
	if !parsed.ActiveOnly {
		t.Fatalf("ActiveOnly = false, want true")
	}
	if parsed.Compact {
		t.Fatalf("Compact = true, want false")
	}
	if parsed.Page != 2 {
		t.Fatalf("Page = %d, want 2", parsed.Page)
	}
	if parsed.Limit != maxListTasksToolLimit {
		t.Fatalf("Limit = %d, want cap %d", parsed.Limit, maxListTasksToolLimit)
	}
}

// X2: with no project given the tool lists the caller's business rather than
// defaulting into a "personal" workspace, so there is no workspace default left
// to assert. The pagination defaults this also covered still matter.
func TestParseListTasksToolInputAppliesPaginationDefaults(t *testing.T) {
	parsed, err := parseListTasksToolInput(map[string]any{}, 99)
	if err != nil {
		t.Fatalf("parseListTasksToolInput returned error: %v", err)
	}
	if parsed.ProjectID != 0 {
		t.Fatalf("ProjectID = %d, want 0 for a business-wide listing", parsed.ProjectID)
	}
	if parsed.Page != 1 {
		t.Fatalf("Page = %d, want 1", parsed.Page)
	}
	if parsed.Limit != defaultListTasksToolLimit {
		t.Fatalf("Limit = %d, want %d", parsed.Limit, defaultListTasksToolLimit)
	}
	if !parsed.Compact {
		t.Fatalf("Compact = false, want true")
	}
}

func TestParseListTasksToolInputRejectsInvalidStatus(t *testing.T) {
	_, err := parseListTasksToolInput(map[string]any{"status": "stuck"}, 99)
	if err == nil || !strings.Contains(err.Error(), "status") {
		t.Fatalf("error = %v, want status validation", err)
	}
}

func TestParseListTasksToolInputRejectsInvalidBool(t *testing.T) {
	_, err := parseListTasksToolInput(map[string]any{"active_only": "sometimes"}, 99)
	if err == nil || !strings.Contains(err.Error(), "active_only") {
		t.Fatalf("error = %v, want active_only validation", err)
	}
}

func TestParseGetTaskToolInputAcceptsAliases(t *testing.T) {
	for _, input := range []map[string]any{
		{"task_id": "17"},
		{"task": "T-17"},
		{"id": "#17"},
		{"task_id": "task-17"},
	} {
		got, err := parseGetTaskToolInput(input)
		if err != nil {
			t.Fatalf("parseGetTaskToolInput(%v) returned error: %v", input, err)
		}
		if got != 17 {
			t.Fatalf("parseGetTaskToolInput(%v) = %d, want 17", input, got)
		}
	}
}

func TestParseGetTaskToolInputRejectsMissingTask(t *testing.T) {
	_, err := parseGetTaskToolInput(map[string]any{})
	if err == nil || !strings.Contains(err.Error(), "task_id is required") {
		t.Fatalf("error = %v, want missing task_id validation", err)
	}
}

func TestTaskToolPayload(t *testing.T) {
	dueDate := time.Date(2026, 6, 21, 0, 0, 0, 0, time.UTC)
	startDate := time.Date(2026, 6, 19, 0, 0, 0, 0, time.UTC)
	completedAt := time.Date(2026, 6, 20, 9, 30, 0, 0, time.UTC)
	assigneeID := int64(7)
	estimatedHours := 2.5
	actualHours := 3.0

	payload := taskToolPayload(domain.Task{
		ID:               17,
		Identifier:       "T-17",
		Title:            "Review invoices",
		Description:      "Check overdue invoices",
		Status:           "done",
		Priority:         "high",
		ProjectID:        42,
		ProjectName:      "Finance Ops",
		WorkspaceID:      3,
		AssigneeID:       &assigneeID,
		AssigneeName:     "Maya Haddad",
		AssigneeInitials: "MH",
		ReporterID:       9,
		DueDate:          &dueDate,
		StartDate:        &startDate,
		CompletedAt:      &completedAt,
		EstimatedHours:   &estimatedHours,
		ActualHours:      &actualHours,
		SubtaskCount:     2,
		Labels:           []int64{1, 2},
		CreatedAt:        time.Date(2026, 6, 19, 10, 0, 0, 0, time.UTC),
		UpdatedAt:        time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC),
	})
	if payload["identifier"] != "T-17" {
		t.Fatalf("identifier = %v, want T-17", payload["identifier"])
	}
	if payload["due_date"] != "2026-06-21" {
		t.Fatalf("due_date = %v, want 2026-06-21", payload["due_date"])
	}
	if payload["completed_at"] != "2026-06-20T09:30:00Z" {
		t.Fatalf("completed_at = %v, want formatted timestamp", payload["completed_at"])
	}
	if payload["subtask_count"] != int32(2) {
		t.Fatalf("subtask_count = %v, want 2", payload["subtask_count"])
	}
}

func TestTaskDetailToolPayloadIncludesSubtasksCommentsAndActivity(t *testing.T) {
	now := time.Date(2026, 6, 20, 10, 0, 0, 0, time.UTC)

	payload := taskDetailToolPayload(domain.TaskDetail{
		Task: domain.Task{
			ID:        17,
			Title:     "Review invoices",
			Status:    "todo",
			Priority:  "high",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Subtasks: []domain.Task{{
			ID:        18,
			Title:     "Pull overdue list",
			Status:    "done",
			CreatedAt: now,
			UpdatedAt: now,
		}},
		Comments: []domain.TaskComment{{
			ID:        5,
			TaskID:    17,
			UserID:    9,
			Content:   "Waiting on accounting.",
			IsEdited:  true,
			CreatedAt: now,
			UpdatedAt: now,
		}},
		Activity: []domain.TaskActivity{{
			ID:        7,
			TaskID:    17,
			UserID:    9,
			Action:    "status_changed",
			OldValue:  "todo",
			NewValue:  "done",
			CreatedAt: now,
		}},
	})

	if _, ok := payload["task"].(map[string]any); !ok {
		t.Fatalf("task payload missing: %#v", payload["task"])
	}
	gotSubtasks, ok := payload["subtasks"].([]map[string]any)
	if !ok {
		t.Fatalf("subtasks has unexpected type: %#v", payload["subtasks"])
	}
	if len(gotSubtasks) != 1 || gotSubtasks[0]["title"] != "Pull overdue list" {
		t.Fatalf("subtasks = %#v, want one subtask", gotSubtasks)
	}
	gotComments, ok := payload["comments"].([]map[string]any)
	if !ok {
		t.Fatalf("comments has unexpected type: %#v", payload["comments"])
	}
	if len(gotComments) != 1 || gotComments[0]["content"] != "Waiting on accounting." {
		t.Fatalf("comments = %#v, want one comment", gotComments)
	}
	gotActivity, ok := payload["activity"].([]map[string]any)
	if !ok {
		t.Fatalf("activity has unexpected type: %#v", payload["activity"])
	}
	if len(gotActivity) != 1 || gotActivity[0]["action"] != "status_changed" {
		t.Fatalf("activity = %#v, want one activity item", gotActivity)
	}
	counts, ok := payload["counts"].(map[string]any)
	if !ok {
		t.Fatalf("counts has unexpected type: %#v", payload["counts"])
	}
	if counts["subtasks"] != 1 || counts["comments"] != 1 || counts["activity"] != 1 {
		t.Fatalf("counts = %#v, want all counts to be 1", counts)
	}
}
