package workspaces

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/tenant"
)

const (
	listTasksToolName         = "flow.list_tasks"
	getTaskToolName           = "flow.get_task"
	createTaskToolName        = "flow.create_task"
	defaultListTasksToolLimit = int64(20)
	maxListTasksToolLimit     = int64(50)
	taskToolTimestampDate     = "2006-01-02T15:04:05Z"
)

type listTasksToolInput struct {
	ProjectID  int64
	Status     string
	AssigneeID *int64
	Priority   string
	ActiveOnly bool
	Compact    bool
	Page       int64
	Limit      int64
}

func (m *Module) ExecuteCaretTool(ctx context.Context, name string, input map[string]any) (map[string]any, bool, error) {
	required, dispatched := flowToolPermissions[name]
	if !dispatched {
		return nil, false, nil
	}
	if err := requireFlowToolPermission(ctx, required); err != nil {
		return nil, true, err
	}
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
		return nil, true, fmt.Errorf("flow tool %q has no executor", name)
	}
}

func (m *Module) executeListTasksTool(ctx context.Context, input map[string]any) (map[string]any, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok || authUser == nil || authUser.UserID <= 0 {
		return nil, errors.New("authentication is required")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, errors.New("business scope could not be resolved")
	}

	parsed, err := parseListTasksToolInput(input, authUser.UserID)
	if err != nil {
		return nil, err
	}

	if parsed.ProjectID != 0 && !m.service.IsProjectMember(ctx, authUser.UserID, parsed.ProjectID) {
		return nil, errors.New("permission denied")
	}

	// X2: Caret's context is the business, not a workspace. ListTasksV2 scopes
	// to the caller's business and filters row by row on what the caller can
	// actually open, which is stricter than the workspace listing this replaced.
	offset := (parsed.Page - 1) * parsed.Limit
	tasks, total, err := m.service.ListTasksV2(ctx, authUser.UserID, businessID, domain.ListTasksFilter{
		ProjectID:  parsed.ProjectID,
		Status:     parsed.Status,
		AssigneeID: parsed.AssigneeID,
		Priority:   parsed.Priority,
		ActiveOnly: parsed.ActiveOnly,
		Compact:    parsed.Compact,
		Limit:      parsed.Limit,
		Offset:     offset,
	})
	if err != nil {
		return nil, err
	}

	payloads := make([]map[string]any, 0, len(tasks))
	for _, task := range tasks {
		payloads = append(payloads, taskToolPayload(task))
	}

	return map[string]any{
		"tasks":     payloads,
		"total":     total,
		"returned":  len(payloads),
		"page":      parsed.Page,
		"page_size": parsed.Limit,
		"has_more":  offset+int64(len(payloads)) < total,
		"filters": map[string]any{
			"project_id":  parsed.ProjectID,
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

	projectID, err := int64Value(input["project_id"])
	if err != nil {
		return nil, fmt.Errorf("invalid project_id: %w", err)
	}
	if projectID < 0 {
		return nil, errors.New("project_id must be non-negative")
	}
	if projectID == 0 {
		// X2: no workspace to file into. Passing 0 is what the v2 task API does
		// for the same case, and EnsureMyTasksProject keys the bucket on
		// (user, business).
		resolvedProjectID, _, ensureErr := m.service.EnsureMyTasksProject(ctx, authUser.UserID, businessID, 0)
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
