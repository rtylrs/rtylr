package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

var importPriorities = map[string]bool{"none": true, "low": true, "medium": true, "high": true, "urgent": true}

// DuplicateTask needs create on Flow and member on the original's project.
// The copy keeps the assignee only when the caller may assign work to others.
func (a *TaskAPIV2) DuplicateTask(ctx context.Context, req *flowtaskv2.DuplicateTaskRequest) (*flowtaskv2.DuplicateTaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:create"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	keepAssignee := requireFlowPermission(ctx, "flow:tasks:assign") == nil
	task, err := a.Service.DuplicateTask(ctx, businessID, caller.UserID, req.GetId(), strings.TrimSpace(req.GetTitle()), keepAssignee)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		if isFieldInputError(err) {
			return nil, fieldError(err, "Failed to duplicate task")
		}
		if mapped := hierarchyError(err); mapped != nil {
			return nil, mapped
		}
		return nil, workspaceInternalError(ctx, "DuplicateTask failed", err)
	}
	publishTaskEventByProject("task.created", task.ProjectID, task.ID, caller.UserID)
	publishBusinessEvent("business.task_created", businessID, task.ID, caller.UserID)
	return &flowtaskv2.DuplicateTaskResponse{Task: toTaskV2(task)}, nil
}

// ImportTasks checks every row before writing any. A refused row comes back
// in Errors with nothing created, so the client can fix the file and resend.
func (a *TaskAPIV2) ImportTasks(ctx context.Context, req *flowtaskv2.ImportTasksRequest) (*flowtaskv2.ImportTasksResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:create"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	projectID := req.GetProjectId()
	if roleErr := a.Service.RequireProjectRole(ctx, projectID, caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.CheckImportProject(ctx, businessID, projectID); err != nil {
		return nil, fieldError(err, "This project cannot take imported tasks")
	}

	rows := make([]application.ImportTaskRow, 0, len(req.GetRows()))
	var rowErrors []*flowtaskv2.ImportRowError
	refuse := func(index int, message string) {
		rowErrors = append(rowErrors, &flowtaskv2.ImportRowError{Row: int32(index + 1), Message: message}) // #nosec G115 -- at most 500 rows.
	}
	for i, row := range req.GetRows() {
		title := strings.TrimSpace(row.GetTitle())
		if title == "" {
			refuse(i, "Title is empty")
			continue
		}
		priority := strings.ToLower(strings.TrimSpace(row.GetPriority()))
		if priority != "" && !importPriorities[priority] {
			refuse(i, fmt.Sprintf("Priority %q is not none, low, medium, high or urgent", row.GetPriority()))
			continue
		}
		dueDate, dateErr := parseTaskDate(strings.TrimSpace(row.GetDueDate()))
		if dateErr != nil {
			refuse(i, "Due date must be YYYY-MM-DD")
			continue
		}
		if row.SectionId != nil && a.Service.GetColumnProject(ctx, *row.SectionId) != projectID {
			refuse(i, "Section is not a status of this project")
			continue
		}
		if row.AssigneeId != nil && requireTaskAssignmentPermission(ctx, row.AssigneeId, caller.UserID) != nil {
			refuse(i, "You cannot assign tasks to other people")
			continue
		}
		subtasks := make([]string, 0, len(row.GetSubtasks()))
		for _, sub := range row.GetSubtasks() {
			if sub = strings.TrimSpace(sub); sub != "" {
				subtasks = append(subtasks, sub)
			}
		}
		rows = append(rows, application.ImportTaskRow{
			Input: domain.TaskInput{
				ProjectID: projectID, OwnerID: caller.UserID, BusinessID: businessID, ReporterID: caller.UserID,
				SectionID: row.SectionId, Title: title, Description: row.GetDescription(), Priority: priority,
				AssigneeID: row.AssigneeId, DueDate: dueDate, EffortMinutes: positiveMinutes(row.EffortMinutes),
			},
			Subtasks: subtasks,
		})
	}
	if len(rowErrors) > 0 {
		return &flowtaskv2.ImportTasksResponse{Errors: rowErrors, TaskIds: []int64{}}, nil
	}

	ids, err := a.Service.ImportTasks(ctx, rows)
	for _, id := range ids {
		publishTaskEventByProject("task.created", projectID, id, caller.UserID)
	}
	if err != nil {
		logger.Ctx(ctx).Error("ImportTasks stopped part way", logger.Int("created", len(ids)), logger.Error(err))
		return nil, onekiterrors.New(http.StatusInternalServerError,
			fmt.Sprintf("Import stopped after %d tasks; the rest were not created", len(ids)))
	}
	return &flowtaskv2.ImportTasksResponse{
		Created: int32(len(ids)), // #nosec G115 -- at most 500 rows.
		TaskIds: ids,
		Errors:  []*flowtaskv2.ImportRowError{},
	}, nil
}

func positiveMinutes(minutes *int32) *int32 {
	if minutes == nil || *minutes <= 0 {
		return nil
	}
	return minutes
}
