package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

func (a *TaskAPIV2) GetMyTasks(ctx context.Context, req *flowtaskv2.GetMyTasksRequest) (*flowtaskv2.GetMyTasksResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	page, limit, _ := resolvePagination(req.GetPage(), req.GetLimit())
	tasks, total, err := a.Service.GetMyTasksForBusiness(ctx, caller.UserID, businessID, int64(page), int64(limit))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get my tasks")
	}
	return &flowtaskv2.GetMyTasksResponse{
		Tasks:      tasksToProtoV2(tasks),
		Pagination: buildPagination(page, limit, total, len(tasks)),
	}, nil
}

func (a *TaskAPIV2) GetMyTasksProject(ctx context.Context, _ *flowtaskv2.GetMyTasksProjectRequest) (*flowtaskv2.GetMyTasksProjectResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	projectID, _, err := a.Service.EnsureMyTasksProject(ctx, caller.UserID, businessID, 0)
	if err != nil || projectID == 0 {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to resolve My Tasks project")
	}
	return &flowtaskv2.GetMyTasksProjectResponse{ProjectId: projectID}, nil
}

func (a *TaskAPIV2) SearchTasks(ctx context.Context, req *flowtaskv2.SearchTasksRequest) (*flowtaskv2.SearchTasksResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	tasks, total, err := a.Service.SearchTasksV2(ctx, caller.UserID, businessID, req.GetQ(), int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to search tasks")
	}
	return &flowtaskv2.SearchTasksResponse{
		Tasks:      tasksToProtoV2(tasks),
		Pagination: buildPagination(page, limit, total, len(tasks)),
	}, nil
}

func (a *TaskAPIV2) CreateTask(ctx context.Context, req *flowtaskv2.CreateTaskRequest) (*flowtaskv2.CreateTaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:create"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTaskAssignmentPermission(ctx, req.AssigneeId, caller.UserID); err != nil {
		return nil, err
	}

	projectID := req.GetProjectId()

	// When project_id is omitted the service provisions the My Tasks bucket and
	// files the task there; the role gate runs against the resolved project.
	if projectID != 0 {
		if roleErr := a.Service.RequireProjectRole(ctx, projectID, caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
			return nil, onekiterrors.PermissionDenied()
		}
	}

	if req.SectionId != nil && (projectID == 0 || a.Service.GetColumnProject(ctx, *req.SectionId) != projectID) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Column must belong to the selected project")
	}
	dueDate, err := parseTaskDate(req.GetDueDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid due_date")
	}
	startDate, err := parseTaskDate(req.GetStartDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date")
	}

	if req.GetStatus() != "" && !domain.IsValidTaskStatus(req.GetStatus()) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid status")
	}
	input := domain.TaskInput{
		ProjectID:      projectID,
		OwnerID:        caller.UserID,
		BusinessID:     businessID,
		ParentTaskID:   req.ParentTaskId,
		SectionID:      req.SectionId,
		Title:          req.GetTitle(),
		Description:    req.GetDescription(),
		Status:         req.GetStatus(),
		EffortMinutes:  req.EffortMinutes,
		Priority:       req.GetPriority(),
		AssigneeID:     req.AssigneeId,
		ReporterID:     caller.UserID,
		DueDate:        dueDate,
		StartDate:      startDate,
		EstimatedHours: req.EstimatedHours,
		Labels:         req.GetLabels(),
	}

	task, createErr := a.Service.CreateTask(ctx, input)
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create task")
	}

	payload := toTaskV2(task)
	publishTaskEventByProject("task.created", task.ProjectID, task.ID, caller.UserID)
	publishBusinessEvent("business.task_created", businessID, task.ID, caller.UserID)
	return &flowtaskv2.CreateTaskResponse{Task: payload}, nil
}

func (a *TaskAPIV2) GetTaskDetail(ctx context.Context, req *flowtaskv2.GetTaskDetailRequest) (*flowtaskv2.GetTaskDetailResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleViewer); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	detail, err := a.Service.GetTaskDetailForViewer(ctx, req.GetId(), caller.UserID, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get task detail")
	}

	comments := make([]*flowtaskv2.TaskComment, 0, len(detail.Comments))
	for _, c := range detail.Comments {
		comments = append(comments, toCommentV2(c))
	}
	activity := make([]*flowtaskv2.TaskActivity, 0, len(detail.Activity))
	for _, act := range detail.Activity {
		activity = append(activity, toActivityV2(act))
	}
	return &flowtaskv2.GetTaskDetailResponse{
		Task:     toTaskV2(detail.Task),
		Subtasks: tasksToProtoV2(detail.Subtasks),
		Comments: comments,
		Activity: activity,
	}, nil
}

func (a *TaskAPIV2) UpdateTask(ctx context.Context, req *flowtaskv2.UpdateTaskRequest) (*flowtaskv2.UpdateTaskResponse, error) {
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
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Missing update body")
	}
	if err := requireTaskAssignmentPermission(ctx, body.AssigneeId, caller.UserID); err != nil {
		return nil, err
	}

	dueDate, err := parseTaskDate(body.GetDueDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid due_date")
	}
	startDate, err := parseTaskDate(body.GetStartDate())
	if err != nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid start_date")
	}

	if body.GetStatus() != "" && !domain.IsValidTaskStatus(body.GetStatus()) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid status")
	}
	if body.EffortMinutes != nil && *body.EffortMinutes < 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid effort_minutes")
	}
	if body.SectionId != nil && *body.SectionId != 0 {
		detail, err := a.Service.GetTaskDetail(ctx, req.GetId())
		if err != nil || a.Service.GetColumnProject(ctx, *body.SectionId) != detail.Task.ProjectID {
			return nil, onekiterrors.New(http.StatusBadRequest, "Column must belong to the task project")
		}
	}
	input := domain.UpdateTaskInput{
		SectionID:        body.SectionId,
		ID:               req.GetId(),
		OwnerID:          caller.UserID,
		BusinessID:       businessID,
		Title:            body.GetTitle(),
		Description:      body.GetDescription(),
		Status:           body.GetStatus(),
		Priority:         body.GetPriority(),
		AssigneeID:       body.AssigneeId,
		DueDate:          dueDate,
		StartDate:        startDate,
		Labels:           body.GetLabels(),
		Archived:         body.Archived,
		EffortMinutes:    body.EffortMinutes,
		EstimatedHours:   body.EstimatedHours,
		FocusPin:         body.FocusPin,
		WaitingOn:        body.WaitingOn,
		DecisionNote:     body.DecisionNote,
		ClearDescription: body.Description != nil && body.GetDescription() == "",
		ClearDueDate:     body.DueDate != nil && dueDate == nil,
		ClearStartDate:   body.StartDate != nil && startDate == nil,
	}

	if err := a.Service.UpdateTask(ctx, input, caller.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update task")
	}

	updated, fetchErr := a.Service.GetTaskDetail(ctx, input.ID)
	if fetchErr != nil {
		return &flowtaskv2.UpdateTaskResponse{Message: "Task updated successfully"}, nil
	}

	publishTaskEventByProject("task.updated", updated.Task.ProjectID, input.ID, caller.UserID)
	publishBusinessEvent("business.task_updated", businessID, input.ID, caller.UserID)
	return &flowtaskv2.UpdateTaskResponse{Message: "Task updated successfully"}, nil
}
