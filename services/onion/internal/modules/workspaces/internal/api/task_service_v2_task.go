package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

func (a *TaskAPIV2) UpdateTaskStatus(ctx context.Context, req *flowtaskv2.UpdateTaskStatusRequest) (*flowtaskv2.UpdateTaskStatusResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	body := req.GetBody()
	if body == nil || !domain.IsValidTaskStatus(body.GetStatus()) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid task status")
	}

	input := domain.UpdateTaskStatusInput{
		ID:         req.GetId(),
		OwnerID:    caller.UserID,
		BusinessID: businessID,
		Status:     body.GetStatus(),
	}
	if err := a.Service.UpdateTaskStatus(ctx, input, caller.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update task status")
	}

	if updated, fetchErr := a.Service.GetTaskDetail(ctx, req.GetId()); fetchErr == nil {
		publishTaskEventByProject("task.status_changed", updated.Task.ProjectID, req.GetId(), caller.UserID)
		publishBusinessEvent("business.task_updated", businessID, req.GetId(), caller.UserID)
	}
	return &flowtaskv2.UpdateTaskStatusResponse{Message: "Task status updated"}, nil
}

func (a *TaskAPIV2) DeleteTask(ctx context.Context, req *flowtaskv2.DeleteTaskRequest) (*flowtaskv2.DeleteTaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	var projectID int64
	if existing, fetchErr := a.Service.GetTaskDetail(ctx, req.GetId()); fetchErr == nil {
		projectID = existing.Task.ProjectID
	}
	if err := a.Service.DeleteTask(ctx, req.GetId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete task")
	}
	publishTaskEventByProject("task.deleted", projectID, req.GetId(), caller.UserID)
	return &flowtaskv2.DeleteTaskResponse{Message: "Task deleted successfully"}, nil
}

func (a *TaskAPIV2) CreateSubtask(ctx context.Context, req *flowtaskv2.CreateSubtaskRequest) (*flowtaskv2.CreateSubtaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:create"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	// Handing the subtask to someone else is the same elevated operation as on
	// create/update; without this a member holding only flow:tasks:create could
	// assign a colleague and grant them project membership.
	if err := requireTaskAssignmentPermission(ctx, req.AssigneeId, caller.UserID); err != nil {
		return nil, err
	}

	parentID := req.GetParentTaskId()
	parent, err := a.Service.GetTaskDetail(ctx, parentID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusNotFound, "Parent task not found")
	}
	// v2 derives everything from the trusted parent row; no client echo of
	// project or workspace ids to mismatch.
	parentProjectID := parent.Task.ProjectID

	if roleErr := a.Service.RequireTaskRole(ctx, parentID, caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	dueDate, err := parseTaskDate(req.GetDueDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid due_date")
	}
	startDate, err := parseTaskDate(req.GetStartDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date")
	}

	task, createErr := a.Service.CreateSubtask(ctx, domain.TaskInput{
		ProjectID:      parentProjectID,
		WorkspaceID:    parent.Task.WorkspaceID,
		OwnerID:        caller.UserID,
		BusinessID:     businessID,
		ParentTaskID:   &parentID,
		Title:          req.GetTitle(),
		Description:    req.GetDescription(),
		Priority:       req.GetPriority(),
		AssigneeID:     req.AssigneeId,
		ReporterID:     caller.UserID,
		DueDate:        dueDate,
		StartDate:      startDate,
		EstimatedHours: req.EstimatedHours,
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create subtask")
	}
	return &flowtaskv2.CreateSubtaskResponse{TaskId: task.ID, Message: "Subtask created successfully"}, nil
}

func (a *TaskAPIV2) GetSubtasks(ctx context.Context, req *flowtaskv2.GetSubtasksRequest) (*flowtaskv2.GetSubtasksResponse, error) {
	caller, _, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, caller.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	subtasks, total, err := a.Service.GetSubtasks(ctx, req.GetTaskId(), int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get subtasks")
	}
	return &flowtaskv2.GetSubtasksResponse{
		Subtasks:   subtasksToCompactV2(subtasks),
		Pagination: buildPagination(page, limit, total, len(subtasks)),
	}, nil
}

func (a *TaskAPIV2) AddComment(ctx context.Context, req *flowtaskv2.AddCommentRequest) (*flowtaskv2.AddCommentResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, caller.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	comment, err := a.Service.AddComment(ctx, req.GetTaskId(), caller.UserID, req.GetContent())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add comment")
	}
	publishCommentEvent("comment.added", req.GetTaskId(), caller.UserID)
	return &flowtaskv2.AddCommentResponse{Comment: toCommentV2(comment)}, nil
}

func (a *TaskAPIV2) DeleteComment(ctx context.Context, req *flowtaskv2.DeleteCommentRequest) (*flowtaskv2.DeleteCommentResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireCommentRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.DeleteComment(ctx, req.GetId(), caller.UserID); err != nil {
		// The delete is scoped to the comment's author, so a comment written by
		// someone else reports ErrUnauthorized: the caller may administer the
		// project but not this row. That is a permission answer, not an outage.
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete comment")
	}
	return &flowtaskv2.DeleteCommentResponse{Message: "Comment deleted successfully"}, nil
}

func (a *TaskAPIV2) GetTaskActivity(ctx context.Context, req *flowtaskv2.GetTaskActivityRequest) (*flowtaskv2.GetTaskActivityResponse, error) {
	caller, _, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, caller.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}
	activities, err := a.Service.GetTaskActivity(ctx, req.GetTaskId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get task activity")
	}
	out := make([]*flowtaskv2.TaskActivity, 0, len(activities))
	for _, act := range activities {
		out = append(out, toActivityV2(act))
	}
	return &flowtaskv2.GetTaskActivityResponse{Activity: out}, nil
}
