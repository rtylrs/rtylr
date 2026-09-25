package workspaces

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
)

func parseGetTaskToolInput(input map[string]any) (int64, error) {
	taskID, err := taskIDToolValue(firstToolValue(input, "task_id", "task", "id"))
	if err != nil {
		return 0, fmt.Errorf("invalid task_id: %w", err)
	}
	if taskID <= 0 {
		return 0, errors.New("task_id is required")
	}
	return taskID, nil
}

func parseListTasksToolInput(input map[string]any, userID int64) (listTasksToolInput, error) {
	projectID, err := int64Value(firstToolValue(input, "project_id", "project"))
	if err != nil {
		return listTasksToolInput{}, fmt.Errorf("invalid project_id: %w", err)
	}
	if projectID < 0 {
		return listTasksToolInput{}, errors.New("project_id must be non-negative")
	}

	status := normalizeTaskToolStatus(input["status"])
	if strings.TrimSpace(stringValue(input["status"])) != "" && status == "" {
		return listTasksToolInput{}, errors.New("status must be one of todo, in_progress, done, or cancelled")
	}

	priority := normalizeTaskFilterPriority(input["priority"])
	if strings.TrimSpace(stringValue(input["priority"])) != "" && priority == "" {
		return listTasksToolInput{}, errors.New("priority must be one of none, low, medium, high, or urgent")
	}

	assigneeID, err := optionalAssigneeID(firstToolValue(input, "assignee_id", "assignee"), userID)
	if err != nil {
		return listTasksToolInput{}, fmt.Errorf("invalid assignee_id: %w", err)
	}

	activeOnly, err := boolValue(input["active_only"])
	if err != nil {
		return listTasksToolInput{}, fmt.Errorf("invalid active_only: %w", err)
	}

	compact := true
	if _, ok := input["compact"]; ok {
		compact, err = boolValue(input["compact"])
		if err != nil {
			return listTasksToolInput{}, fmt.Errorf("invalid compact: %w", err)
		}
	}

	page, err := int64Value(input["page"])
	if err != nil {
		return listTasksToolInput{}, fmt.Errorf("invalid page: %w", err)
	}
	if page < 0 {
		return listTasksToolInput{}, errors.New("page must be non-negative")
	}
	if page == 0 {
		page = 1
	}

	limit, err := int64Value(firstToolValue(input, "limit", "page_size"))
	if err != nil {
		return listTasksToolInput{}, fmt.Errorf("invalid limit: %w", err)
	}
	if limit < 0 {
		return listTasksToolInput{}, errors.New("limit must be non-negative")
	}
	if limit == 0 {
		limit = defaultListTasksToolLimit
	}
	if limit > maxListTasksToolLimit {
		limit = maxListTasksToolLimit
	}

	return listTasksToolInput{
		ProjectID:  projectID,
		Status:     status,
		AssigneeID: assigneeID,
		Priority:   priority,
		ActiveOnly: activeOnly,
		Compact:    compact,
		Page:       page,
		Limit:      limit,
	}, nil
}

func stringValue(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func firstToolValue(input map[string]any, keys ...string) any {
	for _, key := range keys {
		value, ok := input[key]
		if !ok {
			continue
		}
		if raw, ok := value.(string); ok && strings.TrimSpace(raw) == "" {
			continue
		}
		if number, ok := value.(float64); ok && math.IsNaN(number) {
			continue
		}
		return value
	}
	return nil
}

func normalizeTaskToolStatus(value any) string {
	status := strings.ToLower(strings.TrimSpace(stringValue(value)))
	if status == "" {
		return ""
	}
	status = strings.ReplaceAll(status, "-", "_")
	status = strings.ReplaceAll(status, " ", "_")
	if domain.IsValidTaskStatus(status) {
		return status
	}
	return ""
}

func normalizeTaskFilterPriority(value any) string {
	priority := strings.ToLower(strings.TrimSpace(stringValue(value)))
	if priority == "" {
		return ""
	}
	switch priority {
	case "none", "low", "medium", "high", "urgent":
		return priority
	default:
		return ""
	}
}

func normalizeToolPriority(value any) string {
	priority := strings.ToLower(strings.TrimSpace(stringValue(value)))
	if priority == "" {
		return "none"
	}
	switch priority {
	case "none", "low", "medium", "high", "urgent":
		return priority
	default:
		return ""
	}
}

func optionalAssigneeID(value any, userID int64) (*int64, error) {
	raw := strings.ToLower(strings.TrimSpace(stringValue(value)))
	if raw == "" {
		return nil, nil
	}
	if raw == "me" || raw == "self" {
		return &userID, nil
	}
	return optionalInt64(value)
}

func taskIDToolValue(value any) (int64, error) {
	if raw, ok := value.(string); ok {
		text := strings.TrimSpace(raw)
		if text == "" {
			return 0, nil
		}
		normalized := strings.ToLower(text)
		normalized = strings.TrimPrefix(normalized, "#")
		normalized = strings.TrimPrefix(normalized, "task-")
		normalized = strings.TrimPrefix(normalized, "t-")
		if normalized == "" {
			return 0, nil
		}
		return strconv.ParseInt(normalized, 10, 64)
	}
	return int64Value(value)
}

func optionalInt64(value any) (*int64, error) {
	id, err := int64Value(value)
	if err != nil {
		return nil, err
	}
	if id < 0 {
		return nil, errors.New("must be non-negative")
	}
	if id == 0 {
		return nil, nil
	}
	return &id, nil
}
