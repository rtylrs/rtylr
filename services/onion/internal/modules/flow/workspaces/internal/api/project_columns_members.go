package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/wsbus"
	flowprojectv1 "github.com/voxire/voxie/gen/go/flow/project/v1"
)

func publishColumnEvent(eventType string, projectID, columnID, actorID int64) {
	if projectID == 0 {
		return
	}
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("project:%d", projectID),
		Type:     eventType,
		EntityID: columnID,
		ActorID:  actorID,
	})
}

func toProjectColumn(c domain.ProjectColumn) *flowprojectv1.ProjectColumn {
	return &flowprojectv1.ProjectColumn{
		Id:          c.ID,
		ProjectId:   c.ProjectID,
		WorkspaceId: c.WorkspaceID,
		Name:        c.Name,
		Color:       c.Color,
		SortOrder:   int32(c.SortOrder), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toProjectMember(m domain.ProjectMember) *flowprojectv1.ProjectMember {
	role, ok := domain.CanonicalFlowMemberRole(m.Role)
	if !ok {
		role = m.Role
	}
	out := &flowprojectv1.ProjectMember{
		Id:        m.ID,
		ProjectId: m.ProjectID,
		UserId:    m.UserID,
		Name:      m.Name,
		Email:     m.Email,
		Role:      role,
		AddedAt:   m.AddedAt.Format("2006-01-02T15:04:05Z"),
	}
	if m.AddedBy != nil {
		out.AddedBy = *m.AddedBy
	}
	return out
}

func (a *ProjectAPI) ListProjectColumns(ctx context.Context, req *flowprojectv1.ListProjectColumnsRequest) (*flowprojectv1.ListProjectColumnsResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:view"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleViewer); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	cols, total, err := a.Service.GetColumnsByProject(ctx, req.GetProjectId(), int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list columns")
	}
	out := make([]*flowprojectv1.ProjectColumn, 0, len(cols))
	for _, c := range cols {
		out = append(out, toProjectColumn(c))
	}
	return &flowprojectv1.ListProjectColumnsResponse{
		Columns:    out,
		Pagination: buildPagination(page, limit, total, len(cols)),
	}, nil
}

func (a *ProjectAPI) CreateProjectColumn(ctx context.Context, req *flowprojectv1.CreateProjectColumnRequest) (*flowprojectv1.CreateProjectColumnResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleMember); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	workspaceID := a.Service.GetProjectWorkspace(ctx, req.GetProjectId())
	if workspaceID == 0 || req.GetWorkspaceId() != workspaceID {
		return nil, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
	}

	color := req.GetColor()
	if color == "" {
		color = "#6b7280"
	}

	col, err := a.Service.CreateColumn(ctx, domain.ProjectColumnInput{
		ProjectID:   req.GetProjectId(),
		WorkspaceID: workspaceID,
		Name:        req.GetName(),
		Color:       color,
		SortOrder:   int(req.GetSortOrder()),
		CreatedBy:   authUser.UserID,
	})
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create column")
	}
	publishColumnEvent("column.created", col.ProjectID, col.ID, authUser.UserID)
	return &flowprojectv1.CreateProjectColumnResponse{Column: toProjectColumn(col)}, nil
}

func (a *ProjectAPI) UpdateProjectColumn(ctx context.Context, req *flowprojectv1.UpdateProjectColumnRequest) (*flowprojectv1.UpdateProjectColumnResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	err := a.Service.UpdateColumn(ctx, domain.UpdateProjectColumnInput{
		ID:        req.GetId(),
		ActorID:   authUser.UserID,
		Name:      req.GetName(),
		Color:     req.GetColor(),
		SortOrder: int(req.GetSortOrder()),
	})
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update column")
	}
	publishColumnEvent(
		"column.updated",
		a.Service.GetColumnProject(ctx, req.GetId()),
		req.GetId(),
		authUser.UserID,
	)
	return &flowprojectv1.UpdateProjectColumnResponse{Message: "Column updated"}, nil
}

func (a *ProjectAPI) DeleteProjectColumn(ctx context.Context, req *flowprojectv1.DeleteProjectColumnRequest) (*flowprojectv1.DeleteProjectColumnResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:edit"); err != nil {
		return nil, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	projectID := a.Service.GetColumnProject(ctx, req.GetId())
	if err := a.Service.DeleteColumn(ctx, req.GetId(), authUser.UserID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete column")
	}
	publishColumnEvent("column.deleted", projectID, req.GetId(), authUser.UserID)
	return &flowprojectv1.DeleteProjectColumnResponse{Message: "Column deleted"}, nil
}

func (a *ProjectAPI) MoveTaskToColumn(ctx context.Context, req *flowprojectv1.MoveTaskToColumnRequest) (*flowprojectv1.MoveTaskToColumnResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.MoveTaskToColumn(ctx, req.GetTaskId(), req.GetColumnId(), authUser.UserID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return nil, onekiterrors.PermissionDenied()
		}
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to move task")
	}
	projectID := a.Service.GetColumnProject(ctx, req.GetColumnId())
	publishTaskEventByProject("task.updated", projectID, req.GetTaskId(), authUser.UserID)
	return &flowprojectv1.MoveTaskToColumnResponse{Message: "Task moved"}, nil
}

func (a *ProjectAPI) ListProjectMembers(ctx context.Context, req *flowprojectv1.ListProjectMembersRequest) (*flowprojectv1.ListProjectMembersResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:view"); err != nil {
		return nil, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if !a.Service.IsProjectMember(ctx, authUser.UserID, req.GetProjectId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	members, total, err := a.Service.ListProjectMembers(ctx, req.GetProjectId(), int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list project members")
	}
	out := make([]*flowprojectv1.ProjectMember, 0, len(members))
	for _, m := range members {
		out = append(out, toProjectMember(m))
	}
	return &flowprojectv1.ListProjectMembersResponse{
		Members:    out,
		Pagination: buildPagination(page, limit, total, len(members)),
	}, nil
}

func (a *ProjectAPI) AddProjectMember(ctx context.Context, req *flowprojectv1.AddProjectMemberRequest) (*flowprojectv1.AddProjectMemberResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:assign"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	role := req.GetRole()
	if role == "" {
		role = "member"
	}
	canonicalRole, ok := domain.CanonicalFlowMemberRole(role)
	if !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "role must be admin, member, or viewer")
	}
	if err := a.Service.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: req.GetProjectId(),
		UserID:    req.GetUserId(),
		Role:      canonicalRole,
		AddedBy:   authUser.UserID,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add project member")
	}
	return &flowprojectv1.AddProjectMemberResponse{Message: "Member added"}, nil
}

func (a *ProjectAPI) RemoveProjectMember(ctx context.Context, req *flowprojectv1.RemoveProjectMemberRequest) (*flowprojectv1.RemoveProjectMemberResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:assign"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.RemoveProjectMember(ctx, domain.RemoveProjectMemberInput{
		ProjectID: req.GetProjectId(),
		UserID:    req.GetUserId(),
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to remove project member")
	}
	return &flowprojectv1.RemoveProjectMemberResponse{Message: "Member removed"}, nil
}

func (a *ProjectAPI) UpdateProjectMemberRole(ctx context.Context, req *flowprojectv1.UpdateProjectMemberRoleRequest) (*flowprojectv1.UpdateProjectMemberRoleResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:assign"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	role, ok := domain.CanonicalFlowMemberRole(req.GetRole())
	if !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "role must be admin, member, or viewer")
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.UpdateProjectMemberRole(ctx, domain.UpdateProjectMemberRoleInput{
		ProjectID: req.GetProjectId(),
		UserID:    req.GetUserId(),
		Role:      role,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update role")
	}
	return &flowprojectv1.UpdateProjectMemberRoleResponse{Message: "Role updated"}, nil
}

func (a *ProjectAPI) InviteProjectMember(ctx context.Context, req *flowprojectv1.InviteProjectMemberRequest) (*flowprojectv1.InviteProjectMemberResponse, error) {
	if err := requireFlowPermission(ctx, "flow:projects:assign"); err != nil {
		return nil, err
	}
	businessID, authUser, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireProjectRole(ctx, req.GetProjectId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	role := req.GetRole()
	if role == "" {
		role = "member"
	}
	canonicalRole, ok := domain.CanonicalFlowMemberRole(role)
	if !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "role must be admin, member, or viewer")
	}
	found, err := a.Service.InviteProjectMemberByEmail(ctx, req.GetProjectId(), req.GetEmail(), canonicalRole, authUser.UserID)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to invite member")
	}
	if !found {
		return nil, onekiterrors.New(http.StatusNotFound, "No user found with that email")
	}
	return &flowprojectv1.InviteProjectMemberResponse{Message: "Member invited"}, nil
}

func (a *ProjectAPI) TransferProjectOwnership(ctx context.Context, req *flowprojectv1.TransferProjectOwnershipRequest) (*flowprojectv1.TransferProjectOwnershipResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.TransferProjectOwnership(ctx, domain.TransferProjectOwnershipInput{
		ProjectID:  req.GetProjectId(),
		NewOwnerID: req.GetNewOwnerId(),
		CallerID:   authUser.UserID,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to transfer ownership")
	}
	return &flowprojectv1.TransferProjectOwnershipResponse{Message: "Ownership transferred"}, nil
}

func (a *ProjectAPI) MoveProject(ctx context.Context, req *flowprojectv1.MoveProjectRequest) (*flowprojectv1.MoveProjectResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	if !a.Service.IsWorkspaceMember(ctx, authUser.UserID, req.GetWorkspaceId()) {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.MoveProject(ctx, domain.MoveProjectInput{
		ProjectID:   req.GetProjectId(),
		WorkspaceID: req.GetWorkspaceId(),
		CallerID:    authUser.UserID,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to move project")
	}
	return &flowprojectv1.MoveProjectResponse{Message: "Project moved"}, nil
}
