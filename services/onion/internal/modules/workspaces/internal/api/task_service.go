package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
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

func (a *TaskAPI) CreateSubtask(ctx context.Context, req *flowtaskv1.CreateSubtaskRequest) (*flowtaskv1.CreateSubtaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:create"); err != nil {
		return nil, err
	}
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	parentID := req.GetParentTaskId()
	parent, err := a.Service.GetTaskDetail(ctx, parentID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusNotFound, "Parent task not found")
	}
	parentProjectID := parent.Task.ProjectID
	parentWorkspaceID := parent.Task.WorkspaceID

	// Gate on the parent task's scope. Project and workspace IDs are derived
	// from the trusted parent row so a child cannot be persisted into a
	// different tenant or project hierarchy.
	if roleErr := a.Service.RequireTaskRole(ctx, parentID, authUser.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if req.GetProjectId() != 0 && req.GetProjectId() != parentProjectID {
		return nil, onekiterrors.New(http.StatusBadRequest, "project_id must match the parent task")
	}
	if req.GetWorkspaceId() != 0 && req.GetWorkspaceId() != parentWorkspaceID {
		return nil, onekiterrors.New(http.StatusBadRequest, "workspace_id must match the parent task")
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
		WorkspaceID:    parentWorkspaceID,
		OwnerID:        authUser.UserID,
		BusinessID:     businessID,
		ParentTaskID:   &parentID,
		Title:          req.GetTitle(),
		Description:    req.GetDescription(),
		Priority:       req.GetPriority(),
		AssigneeID:     req.AssigneeId,
		ReporterID:     authUser.UserID,
		DueDate:        dueDate,
		StartDate:      startDate,
		EstimatedHours: req.EstimatedHours,
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create subtask")
	}

	return &flowtaskv1.CreateSubtaskResponse{TaskId: task.ID, Message: "Subtask created successfully"}, nil
}

func (a *TaskAPI) CreateAttachment(ctx context.Context, req *flowtaskv1.CreateAttachmentRequest) (*flowtaskv1.CreateAttachmentResponse, error) {
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), authUser.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, createErr := a.Service.CreateAttachment(ctx, domain.TaskAttachmentInput{
		TaskID:     req.GetTaskId(),
		OwnerID:    authUser.UserID,
		BusinessID: businessID,
		Name:       req.GetName(),
		URL:        req.GetUrl(),
		Size:       req.GetSize(),
		UploadedBy: authUser.UserID,
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create attachment")
	}

	return &flowtaskv1.CreateAttachmentResponse{AttachmentId: id, Message: "Attachment created successfully"}, nil
}

func (a *TaskAPI) GetSubtasks(ctx context.Context, req *flowtaskv1.GetSubtasksRequest) (*flowtaskv1.GetSubtasksResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if !a.Service.IsTaskMember(ctx, authUser.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	subtasks, total, err := a.Service.GetSubtasks(ctx, req.GetTaskId(), int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get subtasks")
	}

	return &flowtaskv1.GetSubtasksResponse{
		Subtasks:   subtasksToCompact(subtasks),
		Pagination: buildPagination(page, limit, total, len(subtasks)),
	}, nil
}

func (a *TaskAPI) AddComment(ctx context.Context, req *flowtaskv1.AddCommentRequest) (*flowtaskv1.AddCommentResponse, error) {
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, authUser.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), authUser.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	comment, err := a.Service.AddComment(ctx, req.GetTaskId(), authUser.UserID, req.GetContent())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add comment")
	}

	publishCommentEvent("comment.added", req.GetTaskId(), authUser.UserID)
	return &flowtaskv1.AddCommentResponse{Comment: toComment(comment)}, nil
}

func (a *TaskAPI) DeleteComment(ctx context.Context, req *flowtaskv1.DeleteCommentRequest) (*flowtaskv1.DeleteCommentResponse, error) {
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	if roleErr := a.Service.RequireCommentRole(ctx, req.GetId(), authUser.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.DeleteComment(ctx, req.GetId(), authUser.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete comment")
	}

	return &flowtaskv1.DeleteCommentResponse{Message: "Comment deleted successfully"}, nil
}

func (a *TaskAPI) GetTaskActivity(ctx context.Context, req *flowtaskv1.GetTaskActivityRequest) (*flowtaskv1.GetTaskActivityResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if !a.Service.IsTaskMember(ctx, authUser.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	activities, err := a.Service.GetTaskActivity(ctx, req.GetTaskId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get task activity")
	}

	out := make([]*flowtaskv1.TaskActivity, 0, len(activities))
	for _, act := range activities {
		out = append(out, toActivity(act))
	}
	return &flowtaskv1.GetTaskActivityResponse{Activity: out}, nil
}

func (a *TaskAPI) GetAttachments(ctx context.Context, req *flowtaskv1.GetAttachmentsRequest) (*flowtaskv1.GetAttachmentsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if !a.Service.IsTaskMember(ctx, authUser.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	attachments, err := a.Service.GetAttachments(ctx, req.GetTaskId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get attachments")
	}

	out := make([]*flowtaskv1.Attachment, 0, len(attachments))
	for _, att := range attachments {
		out = append(out, toAttachment(att))
	}
	return &flowtaskv1.GetAttachmentsResponse{Attachments: out}, nil
}

func (a *TaskAPI) DeleteAttachment(ctx context.Context, req *flowtaskv1.DeleteAttachmentRequest) (*flowtaskv1.DeleteAttachmentResponse, error) {
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	if roleErr := a.Service.RequireAttachmentRole(ctx, req.GetId(), authUser.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.DeleteAttachment(ctx, req.GetId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete attachment")
	}

	return &flowtaskv1.DeleteAttachmentResponse{Message: "Attachment deleted successfully"}, nil
}

func (a *TaskAPI) CreateTimeLog(ctx context.Context, req *flowtaskv1.CreateTimeLogRequest) (*flowtaskv1.CreateTimeLogResponse, error) {
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	taskID := req.GetTaskId()
	if taskID == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "task_id is required")
	}

	if roleErr := a.Service.RequireTaskRole(ctx, taskID, authUser.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	id, err := a.Service.CreateTimeLog(ctx, domain.TimeLogInput{
		OwnerID:         authUser.UserID,
		BusinessID:      businessID,
		TaskID:          taskID,
		UserID:          authUser.UserID,
		DurationMinutes: int(req.GetDurationMinutes()),
		Description:     req.GetDescription(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create time log")
	}

	return &flowtaskv1.CreateTimeLogResponse{TimeLogId: id, Message: "Time log created successfully"}, nil
}

func (a *TaskAPI) GetTimeLogs(ctx context.Context, req *flowtaskv1.GetTimeLogsRequest) (*flowtaskv1.GetTimeLogsResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if !a.Service.IsTaskMember(ctx, authUser.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	logs, err := a.Service.GetTimeLogs(ctx, req.GetTaskId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get time logs")
	}

	out := make([]*flowtaskv1.TimeLog, 0, len(logs))
	for _, l := range logs {
		out = append(out, toTimeLog(l))
	}
	return &flowtaskv1.GetTimeLogsResponse{TimeLogs: out}, nil
}

func (a *TaskAPI) DeleteTimeLog(ctx context.Context, req *flowtaskv1.DeleteTimeLogRequest) (*flowtaskv1.DeleteTimeLogResponse, error) {
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	if roleErr := a.Service.RequireTimeLogRole(ctx, req.GetId(), authUser.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.DeleteTimeLogFlat(ctx, req.GetId(), authUser.UserID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete time log")
	}

	return &flowtaskv1.DeleteTimeLogResponse{Message: "Time log deleted successfully"}, nil
}

func (a *TaskAPI) SearchTasks(ctx context.Context, req *flowtaskv1.SearchTasksRequest) (*flowtaskv1.SearchTasksResponse, error) {
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	workspaceID, err := a.Service.ResolveWorkspaceID(ctx, authUser.UserID, businessID, req.GetWorkspaceId())
	if err != nil || workspaceID == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	tasks, total, err := a.Service.SearchTasks(ctx, workspaceID, req.GetQ(), int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to search tasks")
	}

	return &flowtaskv1.SearchTasksResponse{
		Tasks:      tasksToProto(tasks),
		Pagination: buildPagination(page, limit, total, len(tasks)),
	}, nil
}

func (a *TaskAPI) GetMyTasksProject(ctx context.Context, _ *flowtaskv1.GetMyTasksProjectRequest) (*flowtaskv1.GetMyTasksProjectResponse, error) {
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}

	projectID, workspaceID, err := a.Service.EnsureMyTasksProject(ctx, authUser.UserID, businessID, 0)
	if err != nil || projectID == 0 {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to resolve My Tasks project")
	}

	return &flowtaskv1.GetMyTasksProjectResponse{
		ProjectId:   projectID,
		WorkspaceId: workspaceID,
	}, nil
}

// AddTaskDependency and RemoveTaskDependency mirror their v2 counterparts. The
// v1 contract has declared both since voxie added task dependencies; without
// them TaskAPI no longer satisfies flowtaskv1.TaskServiceServer and module.go
// panics while registering the service.
func (a *TaskAPI) AddTaskDependency(ctx context.Context, req *flowtaskv1.AddTaskDependencyRequest) (*flowtaskv1.AddTaskDependencyResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	body := req.GetBody()
	if body == nil || body.GetDependsOnTaskId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "depends_on_task_id is required")
	}
	if err := a.Service.AddTaskDependency(ctx, req.GetTaskId(), body.GetDependsOnTaskId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add dependency")
	}
	return &flowtaskv1.AddTaskDependencyResponse{Message: "Dependency added"}, nil
}

func (a *TaskAPI) RemoveTaskDependency(ctx context.Context, req *flowtaskv1.RemoveTaskDependencyRequest) (*flowtaskv1.RemoveTaskDependencyResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if req.GetDependsOnTaskId() == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "depends_on_task_id is required")
	}
	if err := a.Service.RemoveTaskDependency(ctx, req.GetTaskId(), req.GetDependsOnTaskId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to remove dependency")
	}
	return &flowtaskv1.RemoveTaskDependencyResponse{Message: "Dependency removed"}, nil
}
