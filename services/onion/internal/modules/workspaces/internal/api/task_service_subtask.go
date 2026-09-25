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

func (a *TaskAPI) CreateSubtask(ctx context.Context, req *flowtaskv1.CreateSubtaskRequest) (*flowtaskv1.CreateSubtaskResponse, error) {
	if err := requireFlowPermission(ctx, "flow:tasks:create"); err != nil {
		return nil, err
	}
	authUser, businessID, err := a.resolveAuth(ctx)
	if err != nil {
		return nil, err
	}
	// Same elevated operation as CreateTask/UpdateTask: assigning another user
	// is gated on flow:tasks:assign, self-assignment is not.
	if err := requireTaskAssignmentPermission(ctx, req.AssigneeId, authUser.UserID); err != nil {
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
		// The delete is scoped to the comment's author, so a comment written by
		// someone else reports ErrUnauthorized: the caller may administer the
		// project but not this row. That is a permission answer, not an outage.
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
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
