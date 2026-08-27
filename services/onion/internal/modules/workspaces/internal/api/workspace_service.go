package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/wsbus"
	flowworkspacev1 "github.com/voxire/voxie/gen/go/flow/workspace/v1"
)

type WorkspaceAPI struct {
	Service *application.Service
}

func NewWorkspaceAPI(service *application.Service) *WorkspaceAPI {
	return &WorkspaceAPI{Service: service}
}

func toWorkspace(w domain.Workspace) *flowworkspacev1.Workspace {
	return &flowworkspacev1.Workspace{
		Id:          w.ID,
		OwnerId:     w.OwnerID,
		Name:        w.Name,
		Description: w.Description,
		Icon:        w.Icon,
		Color:       w.Color,
		IsDefault:   w.IsDefault,
		CreatedBy:   w.CreatedBy,
		CreatedAt:   w.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   w.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func (a *WorkspaceAPI) resolveBusinessID(ctx context.Context) (int64, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return 0, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return 0, onekiterrors.PermissionDenied()
	}
	return businessID, nil
}

func (a *WorkspaceAPI) ListWorkspaces(ctx context.Context, req *flowworkspacev1.ListWorkspacesRequest) (*flowworkspacev1.ListWorkspacesResponse, error) {
	businessID, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}

	page, limit, _ := resolvePagination(req.GetPage(), req.GetLimit())
	workspaces, total, err := a.Service.ListWorkspaces(ctx, businessID, int64(page), int64(limit))
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list workspaces")
	}

	out := make([]*flowworkspacev1.Workspace, 0, len(workspaces))
	for _, w := range workspaces {
		out = append(out, toWorkspace(w))
	}
	return &flowworkspacev1.ListWorkspacesResponse{
		Workspaces: out,
		Pagination: buildPagination(page, limit, total, len(workspaces)),
	}, nil
}

func (a *WorkspaceAPI) CreateWorkspace(ctx context.Context, req *flowworkspacev1.CreateWorkspaceRequest) (*flowworkspacev1.CreateWorkspaceResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	input := domain.WorkspaceInput{
		BusinessID:  businessID,
		OwnerID:     authUser.UserID,
		Name:        req.GetName(),
		Description: req.GetDescription(),
		Icon:        req.GetIcon(),
		Color:       req.GetColor(),
		CreatedBy:   authUser.UserID,
	}

	workspace, err := a.Service.CreateWorkspace(ctx, input)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to create workspace")
	}

	return &flowworkspacev1.CreateWorkspaceResponse{Workspace: toWorkspace(workspace)}, nil
}

func (a *WorkspaceAPI) UpdateWorkspace(ctx context.Context, req *flowworkspacev1.UpdateWorkspaceRequest) (*flowworkspacev1.UpdateWorkspaceResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireWorkspaceRole(ctx, req.GetId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	body := req.GetBody()
	input := domain.UpdateWorkspaceInput{
		ID:          req.GetId(),
		OwnerID:     authUser.UserID,
		BusinessID:  businessID,
		Name:        body.GetName(),
		Description: body.GetDescription(),
		Icon:        body.GetIcon(),
		Color:       body.GetColor(),
	}

	if err := a.Service.UpdateWorkspace(ctx, input); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update workspace")
	}

	return &flowworkspacev1.UpdateWorkspaceResponse{Message: "Workspace updated successfully"}, nil
}

func (a *WorkspaceAPI) DeleteWorkspace(ctx context.Context, req *flowworkspacev1.DeleteWorkspaceRequest) (*flowworkspacev1.DeleteWorkspaceResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireWorkspaceRole(ctx, req.GetId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.DeleteWorkspace(ctx, req.GetId(), businessID); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to delete workspace")
	}

	return &flowworkspacev1.DeleteWorkspaceResponse{Message: "Workspace deleted successfully"}, nil
}

func (a *WorkspaceAPI) AddWorkspaceMember(ctx context.Context, req *flowworkspacev1.AddWorkspaceMemberRequest) (*flowworkspacev1.AddWorkspaceMemberResponse, error) {
	if err := requireFlowPermission(ctx, "flow:team:manage"); err != nil {
		return nil, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireWorkspaceRole(ctx, req.GetWorkspaceId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}
	role := req.GetRole()
	if role == "" {
		role = domain.FlowRoleMember
	}
	canonicalRole, ok := domain.CanonicalFlowMemberRole(role)
	if !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "role must be admin, member, or viewer")
	}

	input := domain.AddWorkspaceMemberInput{
		WorkspaceID: req.GetWorkspaceId(),
		UserID:      req.GetUserId(),
		Role:        canonicalRole,
	}

	if err := a.Service.AddWorkspaceMember(ctx, input); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to add workspace member")
	}

	return &flowworkspacev1.AddWorkspaceMemberResponse{Message: "Member added successfully"}, nil
}

func (a *WorkspaceAPI) RemoveWorkspaceMember(ctx context.Context, req *flowworkspacev1.RemoveWorkspaceMemberRequest) (*flowworkspacev1.RemoveWorkspaceMemberResponse, error) {
	if err := requireFlowPermission(ctx, "flow:team:manage"); err != nil {
		return nil, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	if err := a.Service.RequireWorkspaceRole(ctx, req.GetWorkspaceId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	input := domain.RemoveWorkspaceMemberInput{
		WorkspaceID: req.GetWorkspaceId(),
		UserID:      req.GetUserId(),
	}

	if err := a.Service.RemoveWorkspaceMember(ctx, input); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to remove workspace member")
	}

	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("user:%d", req.GetUserId()),
		Type:     "workspace.evicted",
		EntityID: req.GetWorkspaceId(),
		ActorID:  authUser.UserID,
	})
	return &flowworkspacev1.RemoveWorkspaceMemberResponse{Message: "Member removed successfully"}, nil
}

func (a *WorkspaceAPI) resolveWorkspaceID(ctx context.Context, userID, businessID, workspaceID int64) (int64, error) {
	raw := strconv.FormatInt(workspaceID, 10)
	if workspaceID == 0 {
		raw = "personal"
	}
	resolved, err := a.Service.ResolveWorkspaceID(ctx, userID, businessID, raw)
	if err != nil || resolved == 0 {
		return 0, onekiterrors.New(http.StatusBadRequest, "Invalid workspace_id")
	}
	return resolved, nil
}

func toWorkspaceMember(m domain.WorkspaceMember) *flowworkspacev1.WorkspaceMember {
	role, ok := domain.CanonicalFlowMemberRole(m.Role)
	if !ok {
		role = m.Role
	}
	out := &flowworkspacev1.WorkspaceMember{
		Id:          m.ID,
		WorkspaceId: m.WorkspaceID,
		UserId:      m.UserID,
		Name:        m.Name,
		Email:       m.Email,
		Role:        role,
		JoinedAt:    m.JoinedAt.Format("2006-01-02T15:04:05Z"),
	}
	if m.InvitedBy != nil {
		out.InvitedBy = *m.InvitedBy
	}
	return out
}

func (a *WorkspaceAPI) ListWorkspaceMembers(ctx context.Context, req *flowworkspacev1.ListWorkspaceMembersRequest) (*flowworkspacev1.ListWorkspaceMembersResponse, error) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}

	workspaceID, err := a.resolveWorkspaceID(ctx, authUser.UserID, authUser.BusinessID, req.GetWorkspaceId())
	if err != nil {
		return nil, err
	}
	if !a.Service.IsWorkspaceMember(ctx, authUser.UserID, workspaceID) {
		return nil, onekiterrors.PermissionDenied()
	}

	page, limit, offset := resolvePagination(req.GetPage(), req.GetLimit())
	members, total, err := a.Service.ListWorkspaceMembers(ctx, workspaceID, int64(limit), offset)
	if err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to list members")
	}

	out := make([]*flowworkspacev1.WorkspaceMember, 0, len(members))
	for _, m := range members {
		out = append(out, toWorkspaceMember(m))
	}
	return &flowworkspacev1.ListWorkspaceMembersResponse{
		Members:    out,
		Pagination: buildPagination(page, limit, total, len(members)),
	}, nil
}

func (a *WorkspaceAPI) UpdateWorkspaceMemberRole(ctx context.Context, req *flowworkspacev1.UpdateWorkspaceMemberRoleRequest) (*flowworkspacev1.UpdateWorkspaceMemberRoleResponse, error) {
	if err := requireFlowPermission(ctx, "flow:managers:manage"); err != nil {
		return nil, err
	}
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return nil, onekiterrors.PermissionDenied()
	}
	businessID, err := a.resolveBusinessID(ctx)
	if err != nil {
		return nil, err
	}
	role, ok := domain.CanonicalFlowMemberRole(req.GetRole())
	if !ok {
		return nil, onekiterrors.New(http.StatusBadRequest, "role must be admin, member, or viewer")
	}
	if err := a.Service.RequireWorkspaceRole(ctx, req.GetWorkspaceId(), authUser.UserID, businessID, domain.FlowRoleAdmin); err != nil {
		return nil, onekiterrors.PermissionDenied()
	}

	if err := a.Service.UpdateWorkspaceMemberRole(ctx, domain.UpdateWorkspaceMemberRoleInput{
		WorkspaceID: req.GetWorkspaceId(),
		UserID:      req.GetUserId(),
		Role:        role,
	}); err != nil {
		return nil, onekiterrors.New(http.StatusInternalServerError, "Failed to update role")
	}
	return &flowworkspacev1.UpdateWorkspaceMemberRoleResponse{Message: "Role updated"}, nil
}
