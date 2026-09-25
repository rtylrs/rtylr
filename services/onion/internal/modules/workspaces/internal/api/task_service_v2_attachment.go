package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	flowtaskv2 "github.com/voxire/voxie/gen/go/flow/task/v2"
)

func (a *TaskAPIV2) CreateAttachment(ctx context.Context, req *flowtaskv2.CreateAttachmentRequest) (*flowtaskv2.CreateAttachmentResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	id, createErr := a.Service.CreateAttachment(ctx, domain.TaskAttachmentInput{
		TaskID: req.GetTaskId(), OwnerID: caller.UserID, BusinessID: businessID,
		Name: req.GetName(), URL: req.GetUrl(), Size: req.GetSize(), UploadedBy: caller.UserID,
	})
	if createErr != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create attachment")
	}
	return &flowtaskv2.CreateAttachmentResponse{AttachmentId: id, Message: "Attachment created successfully"}, nil
}

func (a *TaskAPIV2) GetAttachments(ctx context.Context, req *flowtaskv2.GetAttachmentsRequest) (*flowtaskv2.GetAttachmentsResponse, error) {
	caller, _, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, caller.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}
	attachments, err := a.Service.GetAttachments(ctx, req.GetTaskId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get attachments")
	}
	out := make([]*flowtaskv2.Attachment, 0, len(attachments))
	for _, att := range attachments {
		out = append(out, toAttachmentV2(att))
	}
	return &flowtaskv2.GetAttachmentsResponse{Attachments: out}, nil
}

func (a *TaskAPIV2) DeleteAttachment(ctx context.Context, req *flowtaskv2.DeleteAttachmentRequest) (*flowtaskv2.DeleteAttachmentResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireAttachmentRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.DeleteAttachment(ctx, req.GetId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete attachment")
	}
	return &flowtaskv2.DeleteAttachmentResponse{Message: "Attachment deleted successfully"}, nil
}

func (a *TaskAPIV2) CreateTimeLog(ctx context.Context, req *flowtaskv2.CreateTimeLogRequest) (*flowtaskv2.CreateTimeLogResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	taskID := req.GetTaskId()
	if taskID == 0 {
		return nil, onekiterrors.New(http.StatusBadRequest, "task_id is required")
	}
	if roleErr := a.Service.RequireTaskRole(ctx, taskID, caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	id, err := a.Service.CreateTimeLog(ctx, domain.TimeLogInput{
		OwnerID: caller.UserID, BusinessID: businessID, TaskID: taskID,
		UserID: caller.UserID, DurationMinutes: int(req.GetDurationMinutes()),
		Description: req.GetDescription(),
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create time log")
	}
	return &flowtaskv2.CreateTimeLogResponse{TimeLogId: id, Message: "Time log created successfully"}, nil
}

func (a *TaskAPIV2) GetTimeLogs(ctx context.Context, req *flowtaskv2.GetTimeLogsRequest) (*flowtaskv2.GetTimeLogsResponse, error) {
	caller, _, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if !a.Service.IsTaskMember(ctx, caller.UserID, req.GetTaskId()) {
		return nil, onekiterrors.PermissionDenied()
	}
	logs, err := a.Service.GetTimeLogs(ctx, req.GetTaskId())
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to get time logs")
	}
	out := make([]*flowtaskv2.TimeLog, 0, len(logs))
	for _, l := range logs {
		out = append(out, toTimeLogV2(l))
	}
	return &flowtaskv2.GetTimeLogsResponse{TimeLogs: out}, nil
}

func (a *TaskAPIV2) DeleteTimeLog(ctx context.Context, req *flowtaskv2.DeleteTimeLogRequest) (*flowtaskv2.DeleteTimeLogResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTimeLogRole(ctx, req.GetId(), caller.UserID, businessID, domain.FlowRoleAdmin); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.DeleteTimeLogFlat(ctx, req.GetId(), caller.UserID); err != nil {
		// Same shape as DeleteComment: a log written by someone else is a
		// permission answer from the row-scoped delete, not an outage.
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete time log")
	}
	return &flowtaskv2.DeleteTimeLogResponse{Message: "Time log deleted successfully"}, nil
}

func (a *TaskAPIV2) AddTaskDependency(ctx context.Context, req *flowtaskv2.AddTaskDependencyRequest) (*flowtaskv2.AddTaskDependencyResponse, error) {
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
	return &flowtaskv2.AddTaskDependencyResponse{Message: "Dependency added"}, nil
}

func (a *TaskAPIV2) RemoveTaskDependency(ctx context.Context, req *flowtaskv2.RemoveTaskDependencyRequest) (*flowtaskv2.RemoveTaskDependencyResponse, error) {
	caller, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	if roleErr := a.Service.RequireTaskRole(ctx, req.GetTaskId(), caller.UserID, businessID, domain.FlowRoleMember); roleErr != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	if err := a.Service.RemoveTaskDependency(ctx, req.GetTaskId(), req.GetDependsOnTaskId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to remove dependency")
	}
	return &flowtaskv2.RemoveTaskDependencyResponse{Message: "Dependency removed"}, nil
}
