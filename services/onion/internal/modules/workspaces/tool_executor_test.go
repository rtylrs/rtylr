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
	if parsed.WorkspaceRaw != "" {
		t.Fatalf("WorkspaceRaw = %q, want empty when project provided", parsed.WorkspaceRaw)
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

func TestParseListTasksToolInputUsesPersonalWorkspaceDefault(t *testing.T) {
	parsed, err := parseListTasksToolInput(map[string]any{}, 99)
	if err != nil {
		t.Fatalf("parseListTasksToolInput returned error: %v", err)
	}
	if parsed.WorkspaceRaw != defaultListTasksWorkspaceID {
		t.Fatalf("WorkspaceRaw = %q, want personal", parsed.WorkspaceRaw)
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
