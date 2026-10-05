package workspaces

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

func int64Value(value any) (int64, error) {
	switch v := value.(type) {
	case nil:
		return 0, nil
	case int:
		return int64(v), nil
	case int64:
		return v, nil
	case int32:
		return int64(v), nil
	case float64:
		if v == 0 {
			return 0, nil
		}
		if math.Trunc(v) != v {
			return 0, errors.New("must be an integer")
		}
		return int64(v), nil
	case json.Number:
		if strings.TrimSpace(v.String()) == "" {
			return 0, nil
		}
		return v.Int64()
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" || raw == "0" {
			return 0, nil
		}
		return strconv.ParseInt(raw, 10, 64)
	default:
		return 0, fmt.Errorf("unsupported value %T", value)
	}
}

func boolValue(value any) (bool, error) {
	switch v := value.(type) {
	case nil:
		return false, nil
	case bool:
		return v, nil
	case string:
		raw := strings.ToLower(strings.TrimSpace(v))
		if raw == "" {
			return false, nil
		}
		switch raw {
		case "true", "1", "yes", "y":
			return true, nil
		case "false", "0", "no", "n":
			return false, nil
		default:
			return false, errors.New("must be boolean")
		}
	case int:
		if v == 0 {
			return false, nil
		}
		if v == 1 {
			return true, nil
		}
		return false, errors.New("must be 0 or 1")
	case int64:
		if v == 0 {
			return false, nil
		}
		if v == 1 {
			return true, nil
		}
		return false, errors.New("must be 0 or 1")
	case float64:
		if v == 0 {
			return false, nil
		}
		if v == 1 {
			return true, nil
		}
		return false, errors.New("must be 0 or 1")
	case json.Number:
		raw := strings.TrimSpace(v.String())
		if raw == "" || raw == "0" {
			return false, nil
		}
		if raw == "1" {
			return true, nil
		}
		return false, errors.New("must be 0 or 1")
	default:
		return false, fmt.Errorf("unsupported value %T", value)
	}
}

func optionalFloat64(value any) (*float64, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case float64:
		if v < 0 {
			return nil, errors.New("must be non-negative")
		}
		return &v, nil
	case int:
		f := float64(v)
		if f < 0 {
			return nil, errors.New("must be non-negative")
		}
		return &f, nil
	case int64:
		f := float64(v)
		if f < 0 {
			return nil, errors.New("must be non-negative")
		}
		return &f, nil
	case json.Number:
		f, err := v.Float64()
		if err != nil {
			return nil, err
		}
		if f < 0 {
			return nil, errors.New("must be non-negative")
		}
		return &f, nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return nil, nil
		}
		f, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, err
		}
		if f < 0 {
			return nil, errors.New("must be non-negative")
		}
		return &f, nil
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func optionalToolDate(value any) (*time.Time, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case time.Time:
		return &v, nil
	case string:
		raw := strings.TrimSpace(v)
		if raw == "" {
			return nil, nil
		}
		for _, layout := range []string{"2006-01-02", time.RFC3339, time.RFC3339Nano} {
			parsed, err := time.Parse(layout, raw)
			if err == nil {
				return &parsed, nil
			}
		}
		return nil, errors.New("must be YYYY-MM-DD or RFC3339")
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func optionalInt64Slice(value any) ([]int64, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []int64:
		return v, nil
	case []any:
		out := make([]int64, 0, len(v))
		for _, item := range v {
			id, err := int64Value(item)
			if err != nil {
				return nil, err
			}
			if id < 0 {
				return nil, errors.New("must be non-negative")
			}
			if id != 0 {
				out = append(out, id)
			}
		}
		return out, nil
	default:
		return nil, fmt.Errorf("unsupported value %T", value)
	}
}

func taskDetailToolPayload(detail domain.TaskDetail) map[string]any {
	subtasks := make([]map[string]any, 0, len(detail.Subtasks))
	for _, subtask := range detail.Subtasks {
		subtasks = append(subtasks, taskToolPayload(subtask))
	}

	comments := make([]map[string]any, 0, len(detail.Comments))
	for _, comment := range detail.Comments {
		comments = append(comments, taskCommentToolPayload(comment))
	}

	activity := make([]map[string]any, 0, len(detail.Activity))
	for _, item := range detail.Activity {
		activity = append(activity, taskActivityToolPayload(item))
	}

	return map[string]any{
		"task":     taskToolPayload(detail.Task),
		"subtasks": subtasks,
		"comments": comments,
		"activity": activity,
		"counts": map[string]any{
			"subtasks": len(subtasks),
			"comments": len(comments),
			"activity": len(activity),
		},
	}
}

func taskToolPayload(task domain.Task) map[string]any {
	labels := task.Labels
	if labels == nil {
		labels = []int64{}
	}

	return map[string]any{
		"id":                task.ID,
		"identifier":        task.Identifier,
		"title":             task.Title,
		"description":       task.Description,
		"status":            task.Status,
		"priority":          task.Priority,
		"project_id":        task.ProjectID,
		"project_name":      task.ProjectName,
		"workspace_id":      task.WorkspaceID,
		"parent_task_id":    task.ParentTaskID,
		"section_id":        task.SectionID,
		"cycle_id":          task.CycleID,
		"milestone_id":      task.MilestoneID,
		"assignee_id":       task.AssigneeID,
		"assignee_name":     task.AssigneeName,
		"assignee_initials": task.AssigneeInitials,
		"reporter_id":       task.ReporterID,
		"due_date":          formatToolDate(task.DueDate),
		"start_date":        formatToolDate(task.StartDate),
		"completed_at":      formatTaskToolTime(task.CompletedAt),
		"estimated_hours":   task.EstimatedHours,
		"actual_hours":      task.ActualHours,
		"subtask_count":     task.SubtaskCount,
		"labels":            labels,
		"created_at":        formatRequiredTaskToolTime(task.CreatedAt),
		"updated_at":        formatRequiredTaskToolTime(task.UpdatedAt),
	}
}

func taskCommentToolPayload(comment domain.TaskComment) map[string]any {
	return map[string]any{
		"id":         comment.ID,
		"task_id":    comment.TaskID,
		"user_id":    comment.UserID,
		"content":    comment.Content,
		"is_edited":  comment.IsEdited,
		"created_at": formatRequiredTaskToolTime(comment.CreatedAt),
		"updated_at": formatRequiredTaskToolTime(comment.UpdatedAt),
	}
}

func taskActivityToolPayload(activity domain.TaskActivity) map[string]any {
	return map[string]any{
		"id":         activity.ID,
		"task_id":    activity.TaskID,
		"user_id":    activity.UserID,
		"action":     activity.Action,
		"old_value":  activity.OldValue,
		"new_value":  activity.NewValue,
		"created_at": formatRequiredTaskToolTime(activity.CreatedAt),
	}
}

func formatTaskToolTime(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format(taskToolTimestampDate)
}

func formatRequiredTaskToolTime(value time.Time) string {
	if value.IsZero() {
		return ""
	}
	return value.Format(taskToolTimestampDate)
}

func formatToolDate(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format("2006-01-02")
}
