package workspaces

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	if name != "flow.create_task" {
		return nil, false, nil
	}
	output, err := m.executeCreateTaskTool(ctx, input)
	return output, true, err
}

func (m *Module) executeCreateTaskTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	title := strings.TrimSpace(stringValue(input["title"]))
	if title == "" {
		return nil, errors.New("title is required")
	}
	if len(title) > 500 {
		return nil, errors.New("title must be 500 characters or fewer")
	}

	priority := normalizeToolPriority(input["priority"])
	if priority == "" {
		return nil, errors.New("priority must be one of none, low, medium, high, or urgent")
	}

	workspaceRaw := strings.TrimSpace(stringValue(input["workspace_id"]))
	if workspaceRaw == "" {
		workspaceRaw = "personal"
	}
	workspaceID, err := m.service.ResolveWorkspaceID(ctx, authUser.UserID, workspaceRaw)
	if err != nil || workspaceID <= 0 {
		return nil, errors.New("invalid workspace_id")
	}

	projectID, err := int64Value(input["project_id"])
	if err != nil {
		return nil, fmt.Errorf("invalid project_id: %w", err)
	}
	if projectID < 0 {
		return nil, errors.New("project_id must be non-negative")
	}
	if projectID == 0 {
		resolvedProjectID, _, ensureErr := m.service.EnsureMyTasksProject(ctx, authUser.UserID, workspaceID)
		if ensureErr != nil || resolvedProjectID <= 0 {
			return nil, errors.New("failed to resolve My Tasks project")
		}
		projectID = resolvedProjectID
	}

	parentTaskID, err := optionalInt64(input["parent_task_id"])
	if err != nil {
		return nil, fmt.Errorf("invalid parent_task_id: %w", err)
	}
	assigneeID, err := optionalAssigneeID(input["assignee_id"], authUser.UserID)
	if err != nil {
		return nil, fmt.Errorf("invalid assignee_id: %w", err)
	}
	dueDate, err := optionalToolDate(input["due_date"])
	if err != nil {
		return nil, fmt.Errorf("invalid due_date: %w", err)
	}
	startDate, err := optionalToolDate(input["start_date"])
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}
	estimatedHours, err := optionalFloat64(input["estimated_hours"])
	if err != nil {
		return nil, fmt.Errorf("invalid estimated_hours: %w", err)
	}
	labels, err := optionalInt64Slice(input["labels"])
	if err != nil {
		return nil, fmt.Errorf("invalid labels: %w", err)
	}

	task, err := m.service.CreateTask(ctx, domain.TaskInput{
		ProjectID:      projectID,
		WorkspaceID:    workspaceID,
		OwnerID:        authUser.UserID,
		ParentTaskID:   parentTaskID,
		Title:          title,
		Description:    strings.TrimSpace(stringValue(input["description"])),
		Priority:       priority,
		AssigneeID:     assigneeID,
		ReporterID:     authUser.UserID,
		DueDate:        dueDate,
		StartDate:      startDate,
		EstimatedHours: estimatedHours,
		Labels:         labels,
	})
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"task": map[string]any{
			"id":              task.ID,
			"title":           task.Title,
			"description":     task.Description,
			"status":          task.Status,
			"priority":        task.Priority,
			"project_id":      task.ProjectID,
			"workspace_id":    task.WorkspaceID,
			"parent_task_id":  task.ParentTaskID,
			"assignee_id":     task.AssigneeID,
			"reporter_id":     task.ReporterID,
			"due_date":        formatToolDate(task.DueDate),
			"start_date":      formatToolDate(task.StartDate),
			"estimated_hours": task.EstimatedHours,
			"labels":          task.Labels,
		},
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

func formatToolDate(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.Format("2006-01-02")
}
