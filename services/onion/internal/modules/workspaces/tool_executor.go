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
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listTasksToolName           = "flow.list_tasks"
	getTaskToolName             = "flow.get_task"
	createTaskToolName          = "flow.create_task"
	defaultListTasksToolLimit   = int64(20)
	maxListTasksToolLimit       = int64(50)
	taskToolTimestampDate       = "2006-01-02T15:04:05Z"
	defaultListTasksWorkspaceID = "personal"
)

type listTasksToolInput struct {
	ProjectID    int64
	WorkspaceRaw string
	Status       string
	AssigneeID   *int64
	Priority     string
	ActiveOnly   bool
	Compact      bool
	Page         int64
	Limit        int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	switch name {
	case listTasksToolName:
		output, err := m.executeListTasksTool(ctx, input)
		return output, true, err
	case getTaskToolName:
		output, err := m.executeGetTaskTool(ctx, input)
		return output, true, err
	case createTaskToolName:
		output, err := m.executeCreateTaskTool(ctx, input)
		return output, true, err
	default:
		return nil, false, nil
	}
}

func (m *Module) executeListTasksTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	parsed, err := parseListTasksToolInput(input, authUser.UserID)
	if err != nil {
		return nil, err
	}

	workspaceID := int64(0)
	if parsed.WorkspaceRaw != "" {
		resolved, resolveErr := m.service.ResolveWorkspaceID(ctx, authUser.UserID, parsed.WorkspaceRaw)
		if resolveErr != nil || resolved <= 0 {
			return nil, errors.New("invalid workspace_id")
		}
		workspaceID = resolved
	}

	if parsed.ProjectID != 0 && !m.service.IsProjectMember(ctx, authUser.UserID, parsed.ProjectID) {
		return nil, errors.New("permission denied")
	}

	tasks, err := m.service.ListTasks(ctx, domain.ListTasksFilter{
		ProjectID:   parsed.ProjectID,
		WorkspaceID: workspaceID,
		Status:      parsed.Status,
		AssigneeID:  parsed.AssigneeID,
		Priority:    parsed.Priority,
		ActiveOnly:  parsed.ActiveOnly,
		Compact:     parsed.Compact,
	})
	if err != nil {
		return nil, err
	}

	start := int((parsed.Page - 1) * parsed.Limit)
	if start > len(tasks) {
		start = len(tasks)
	}
	end := start + int(parsed.Limit)
	if end > len(tasks) {
		end = len(tasks)
	}

	payloads := make([]map[string]any, 0, end-start)
	for _, task := range tasks[start:end] {
		payloads = append(payloads, taskToolPayload(task))
	}

	return map[string]any{
		"tasks":     payloads,
		"total":     len(tasks),
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  end < len(tasks),
		"filters": map[string]any{
			"project_id":  parsed.ProjectID,
			"workspace":   parsed.WorkspaceRaw,
			"status":      parsed.Status,
			"assignee_id": parsed.AssigneeID,
			"priority":    parsed.Priority,
			"active_only": parsed.ActiveOnly,
		},
	}, nil
}

func (m *Module) executeGetTaskTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}

	taskID, err := parseGetTaskToolInput(input)
	if err != nil {
		return nil, err
	}

	if !m.service.IsTaskMember(ctx, authUser.UserID, taskID) {
		return nil, errors.New("permission denied")
	}

	detail, err := m.service.GetTaskDetail(ctx, taskID)
	if err != nil {
		return nil, err
	}

	return taskDetailToolPayload(detail), nil
}

func (m *Module) executeCreateTaskTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("business scope could not be resolved")
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
		BusinessID:     businessID,
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

	workspaceRaw := strings.TrimSpace(stringValue(firstToolValue(input, "workspace_id", "workspace")))
	if projectID == 0 && workspaceRaw == "" {
		workspaceRaw = defaultListTasksWorkspaceID
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
		ProjectID:    projectID,
		WorkspaceRaw: workspaceRaw,
		Status:       status,
		AssigneeID:   assigneeID,
		Priority:     priority,
		ActiveOnly:   activeOnly,
		Compact:      compact,
		Page:         page,
		Limit:        limit,
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
	switch status {
	case "todo", "in_progress", "done", "cancelled":
		return status
	default:
		return ""
	}
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
