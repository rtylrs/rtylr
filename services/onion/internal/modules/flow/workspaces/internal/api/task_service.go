package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
)

type TaskAPI struct {
	Service *application.Service
}

func NewTaskAPI(service *application.Service) *TaskAPI {
	return &TaskAPI{Service: service}
}

func (a *TaskAPI) resolveAuth(ctx context.Context) (*authPkg.AuthUser, int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return nil, 0, onekiterrors.PermissionDenied()
	}
	return authUser, businessID, nil
}

func (a *TaskAPI) ListTasks(ctx context.Context, req *flowtaskv1.ListTasksRequest) (*flowtaskv1.ListTasksResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	var workspaceID int64
	if req.GetWorkspaceId() != "" {
		resolved, resolveErr := a.Service.ResolveWorkspaceID(ctx, authUser.UserID, businessID, req.GetWorkspaceId())
		if resolveErr != nil {
			return nil, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
		}
		workspaceID = resolved
	}

	if req.GetProjectId() != 0 && !a.Service.IsProjectMember(ctx, authUser.UserID, req.GetProjectId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	filter := domain.ListTasksFilter{
		ProjectID:   req.GetProjectId(),
		BusinessID:  businessID,
		WorkspaceID: workspaceID,
		Status:      req.GetStatus(),
		AssigneeID:  req.AssigneeId,
		Priority:    req.GetPriority(),
		Limit:       int64(limit),
		Offset:      offset,
	}

	tasks, total, err := a.Service.ListTasks(ctx, filter)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list tasks")
	}

	return &flowtaskv1.ListTasksResponse{
		Tasks:      tasksToProto(tasks),
		Pagination: buildPagination(page, limit, total, len(tasks)),
	}, nil
}

func (a *TaskAPI) GetMyTasks(ctx context.Context, req *flowtaskv1.GetMyTasksRequest) (*flowtaskv1.GetMyTasksResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, _ := resolvePagination(req.GetPage(), req.GetLimit())
	tasks, total, err := a.Service.GetMyTasks(ctx, authUser.UserID, int64(page), int64(limit))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get my tasks")
	}

	return &flowtaskv1.GetMyTasksResponse{
		Tasks:      tasksToProto(tasks),
		Pagination: buildPagination(page, limit, total, len(tasks)),
	}, nil
}

func (a *TaskAPI) CreateTask(ctx context.Context, req *flowtaskv1.CreateTaskRequest) (*flowtaskv1.CreateTaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:create"); err != nil {
		return nil, err
	}
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if err := requireTaskAssignmentPermission(ctx, req.AssigneeId, authUser.UserID); err != nil {
		return nil, err
	}

	workspaceID, err := a.Service.ResolveWorkspaceID(ctx, authUser.UserID, businessID, formatWorkspaceID(req.GetWorkspaceId()))
	if err != nil || workspaceID == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
	}

	projectID := req.GetProjectId()
	if projectID == 0 {
		myTasksID, _, ensureErr := a.Service.EnsureMyTasksProject(ctx, authUser.UserID, businessID, workspaceID)
		if ensureErr != nil || myTasksID == 0 {
			return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to resolve My Tasks project")
		}
		projectID = myTasksID
	}

	if roleErr := a.Service.RequireProjectRole(ctx, projectID, authUser.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
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

	input := domain.TaskInput{
		ProjectID:      projectID,
		WorkspaceID:    workspaceID,
		OwnerID:        authUser.UserID,
		BusinessID:     businessID,
		ParentTaskID:   req.ParentTaskId,
		Title:          req.GetTitle(),
		Description:    req.GetDescription(),
		Priority:       req.GetPriority(),
		AssigneeID:     req.AssigneeId,
		ReporterID:     authUser.UserID,
		DueDate:        dueDate,
		StartDate:      startDate,
		EstimatedHours: req.EstimatedHours,
		Labels:         req.GetLabels(),
	}

	task, createErr := a.Service.CreateTask(ctx, input)
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create task")
	}

	payload := toTask(task)
	publishTaskEvent("task.created", payload, authUser.UserID)
	return &flowtaskv1.CreateTaskResponse{Task: payload}, nil
}

func (a *TaskAPI) GetTaskDetail(ctx context.Context, req *flowtaskv1.GetTaskDetailRequest) (*flowtaskv1.GetTaskDetailResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:view"); err != nil {
		return nil, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if !a.Service.IsTaskMember(ctx, authUser.UserID, req.GetId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	detail, err := a.Service.GetTaskDetail(ctx, req.GetId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get task detail")
	}

	comments := make([]*flowtaskv1.TaskComment, 0, len(detail.Comments))
	for _, c := range detail.Comments {
		comments = append(comments, toComment(c))
	}
	activity := make([]*flowtaskv1.TaskActivity, 0, len(detail.Activity))
	for _, act := range detail.Activity {
		activity = append(activity, toActivity(act))
	}

	return &flowtaskv1.GetTaskDetailResponse{
		Task:     toTask(detail.Task),
		Subtasks: tasksToProto(detail.Subtasks),
		Comments: comments,
		Activity: activity,
	}, nil
}

func (a *TaskAPI) UpdateTask(ctx context.Context, req *flowtaskv1.UpdateTaskRequest) (*flowtaskv1.UpdateTaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), authUser.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	body := req.GetBody()
	// AssigneeId and Archived below are read as raw pointers, not through the
	// nil-safe GetX() accessors, so a request that omitted the body entirely
	// would dereference nil and panic into a 500. The request is malformed.
	if body == nil {
		return nil, onekiterrors.New(http.StatusBadRequest, "Missing update body")
	}
	if err := requireTaskAssignmentPermission(ctx, body.AssigneeId, authUser.UserID); err != nil {
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

	input := domain.UpdateTaskInput{
		ID:          req.GetId(),
		OwnerID:     authUser.UserID,
		BusinessID:  businessID,
		Title:       body.GetTitle(),
		Description: body.GetDescription(),
		Status:      body.GetStatus(),
		Priority:    body.GetPriority(),
		AssigneeID:  body.AssigneeId,
		DueDate:     dueDate,
		StartDate:   startDate,
		Labels:      body.GetLabels(),
		// Read the pointer, not GetArchived(): the accessor flattens an absent
		// field to false, which would archive-or-restore a task the request
		// never mentioned. nil here is what makes COALESCE(?, archived) hold.
		Archived: body.Archived,
	}

	if err := a.Service.UpdateTask(ctx, input, authUser.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update task")
	}

	if updated, fetchErr := a.Service.GetTaskDetail(ctx, input.ID); fetchErr == nil {
		publishTaskEventByProject("task.updated", updated.Task.ProjectID, input.ID, authUser.UserID)
	}

	return &flowtaskv1.UpdateTaskResponse{Message: "Task updated successfully"}, nil
}

func (a *TaskAPI) UpdateTaskStatus(ctx context.Context, req *flowtaskv1.UpdateTaskStatusRequest) (*flowtaskv1.UpdateTaskStatusResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), authUser.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	body := req.GetBody()
	if body == nil || !domain.IsValidTaskStatus(body.GetStatus()) {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid task status")
	}

	input := domain.UpdateTaskStatusInput{
		ID:         req.GetId(),
		OwnerID:    authUser.UserID,
		BusinessID: businessID,
		Status:     body.GetStatus(),
	}

	if err := a.Service.UpdateTaskStatus(ctx, input, authUser.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update task status")
	}

	if updated, fetchErr := a.Service.GetTaskDetail(ctx, req.GetId()); fetchErr == nil {
		publishTaskEventByProject("task.status_changed", updated.Task.ProjectID, req.GetId(), authUser.UserID)
	}

	return &flowtaskv1.UpdateTaskStatusResponse{Message: "Task status updated"}, nil
}

func (a *TaskAPI) DeleteTask(ctx context.Context, req *flowtaskv1.DeleteTaskRequest) (*flowtaskv1.DeleteTaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:edit"); err != nil {
		return nil, err
	}
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	if roleErr := a.Service.RequireTaskRole(ctx, req.GetId(), authUser.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	var projectID int64
	if existing, fetchErr := a.Service.GetTaskDetail(ctx, req.GetId()); fetchErr == nil {
		projectID = existing.Task.ProjectID
	}

	if err := a.Service.DeleteTask(ctx, req.GetId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete task")
	}

	publishTaskEventByProject("task.deleted", projectID, req.GetId(), authUser.UserID)
	return &flowtaskv1.DeleteTaskResponse{Message: "Task deleted successfully"}, nil
}
