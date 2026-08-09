package domain

import "testing"

func TestIsValidTaskStatusMatchesFlowEnum(t *testing.T) {
	for _, status := range []string{
		TaskStatusBacklog,
		TaskStatusTodo,
		TaskStatusInProgress,
		TaskStatusInReview,
		TaskStatusDone,
		TaskStatusCancelled,
	} {
		if !IsValidTaskStatus(status) {
			t.Errorf("IsValidTaskStatus(%q) = false, want true", status)
		}
	}

	for _, status := range []string{"", "review", "completed", "in progress", "archived"} {
		if IsValidTaskStatus(status) {
			t.Errorf("IsValidTaskStatus(%q) = true, want false", status)
		}
	}
}
