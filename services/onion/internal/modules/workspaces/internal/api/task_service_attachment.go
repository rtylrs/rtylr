package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
)

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
		// Same shape as DeleteComment: a log written by someone else is a
		// permission answer from the row-scoped delete, not an outage.
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
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
