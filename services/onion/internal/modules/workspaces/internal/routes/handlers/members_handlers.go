package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─────────────────────────────────────────────────────────────────────────────
// Shared member payload types
// ─────────────────────────────────────────────────────────────────────────────

type WorkspaceMemberPayload struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspace_id"`
	UserID      int64     `json:"user_id"`
	Name        string    `json:"name"`
	Email       string    `json:"email"`
	Role        string    `json:"role"`
	InvitedBy   *int64    `json:"invited_by,omitempty"`
	JoinedAt    time.Time `json:"joined_at"`
}

type ProjectMemberPayload struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	UserID    int64     `json:"user_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	AddedBy   *int64    `json:"added_by,omitempty"`
	AddedAt   time.Time `json:"added_at"`
}

func toWsMemberPayload(m domain.WorkspaceMember) WorkspaceMemberPayload {
	return WorkspaceMemberPayload{
		ID:          m.ID,
		WorkspaceID: m.WorkspaceID,
		UserID:      m.UserID,
		Name:        m.Name,
		Email:       m.Email,
		Role:        m.Role,
		InvitedBy:   m.InvitedBy,
		JoinedAt:    m.JoinedAt,
	}
}

func toProjMemberPayload(m domain.ProjectMember) ProjectMemberPayload {
	return ProjectMemberPayload{
		ID:        m.ID,
		ProjectID: m.ProjectID,
		UserID:    m.UserID,
		Name:      m.Name,
		Email:     m.Email,
		Role:      m.Role,
		AddedBy:   m.AddedBy,
		AddedAt:   m.AddedAt,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /flow/workspaces/members
// ─────────────────────────────────────────────────────────────────────────────

func ListWorkspaceMembersRoutes() []string { return []string{"/flow/workspaces/members"} }

type (
	ListWorkspaceMembersHandler struct{ FlowService *application.Service }
	ListWorkspaceMembersRequest struct {
		WorkspaceID WorkspaceIDField `json:"workspace_id"`
	}
)

type ListWorkspaceMembersResponse struct {
	Members []WorkspaceMemberPayload `json:"members"`
}

func (h *ListWorkspaceMembersHandler) Validate(_ context.Context, req *ListWorkspaceMembersRequest) *errorModels.ServerError {
	if req.WorkspaceID.IsEmpty() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "workspace_id is required"}
	}
	return nil
}

func (h *ListWorkspaceMembersHandler) Serve(ctx context.Context, req *ListWorkspaceMembersRequest) (ListWorkspaceMembersResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListWorkspaceMembersResponse{}, errorModels.PermissionDeniedServerError()
	}

	workspaceID, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, req.WorkspaceID.String())
	if err != nil || workspaceID == 0 {
		return ListWorkspaceMembersResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid workspace_id"}
	}
	if !h.FlowService.IsWorkspaceMember(ctx, authUser.UserID, workspaceID) {
		return ListWorkspaceMembersResponse{}, errorModels.PermissionDeniedServerError()
	}

	members, err := h.FlowService.ListWorkspaceMembers(ctx, workspaceID)
	if err != nil {
		return ListWorkspaceMembersResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list members"}
	}

	payloads := make([]WorkspaceMemberPayload, 0, len(members))
	for _, m := range members {
		payloads = append(payloads, toWsMemberPayload(m))
	}
	return ListWorkspaceMembersResponse{Members: payloads}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /flow/workspaces/member/role
// ─────────────────────────────────────────────────────────────────────────────

func UpdateWorkspaceMemberRoleRoutes() []string { return []string{"/flow/workspaces/member/role"} }

type (
	UpdateWorkspaceMemberRoleHandler struct{ FlowService *application.Service }
	UpdateWorkspaceMemberRoleRequest struct {
		WorkspaceID JSONID `json:"workspace_id"`
		UserID      JSONID `json:"user_id"`
		Role        string `json:"role"`
	}
)

type UpdateWorkspaceMemberRoleResponse struct {
	Message string `json:"message"`
}

func (h *UpdateWorkspaceMemberRoleHandler) Validate(_ context.Context, req *UpdateWorkspaceMemberRoleRequest) *errorModels.ServerError {
	if req.WorkspaceID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "workspace_id is required"}
	}
	if req.UserID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "user_id is required"}
	}
	if req.Role != "manager" && req.Role != "member" && req.Role != "viewer" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "role must be manager, member, or viewer"}
	}
	return nil
}

func (h *UpdateWorkspaceMemberRoleHandler) Serve(ctx context.Context, req *UpdateWorkspaceMemberRoleRequest) (UpdateWorkspaceMemberRoleResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateWorkspaceMemberRoleResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !h.FlowService.IsWorkspaceMember(ctx, authUser.UserID, req.WorkspaceID.Int64()) {
		return UpdateWorkspaceMemberRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.UpdateWorkspaceMemberRole(ctx, domain.UpdateWorkspaceMemberRoleInput{
		WorkspaceID: req.WorkspaceID.Int64(),
		UserID:      req.UserID.Int64(),
		Role:        req.Role,
	}); err != nil {
		return UpdateWorkspaceMemberRoleResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update role"}
	}
	return UpdateWorkspaceMemberRoleResponse{Message: "Role updated"}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GET /flow/projects/members
// ─────────────────────────────────────────────────────────────────────────────

func ListProjectMembersRoutes() []string { return []string{"/flow/projects/members"} }

type (
	ListProjectMembersHandler struct{ FlowService *application.Service }
	ListProjectMembersRequest struct {
		ProjectID JSONID `json:"project_id"`
	}
)

type ListProjectMembersResponse struct {
	Members []ProjectMemberPayload `json:"members"`
}

func (h *ListProjectMembersHandler) Validate(_ context.Context, req *ListProjectMembersRequest) *errorModels.ServerError {
	if req.ProjectID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "project_id is required"}
	}
	return nil
}

func (h *ListProjectMembersHandler) Serve(ctx context.Context, req *ListProjectMembersRequest) (ListProjectMembersResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListProjectMembersResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !h.FlowService.IsProjectMember(ctx, authUser.UserID, req.ProjectID.Int64()) {
		return ListProjectMembersResponse{}, errorModels.PermissionDeniedServerError()
	}

	members, err := h.FlowService.ListProjectMembers(ctx, req.ProjectID.Int64())
	if err != nil {
		return ListProjectMembersResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list project members"}
	}

	payloads := make([]ProjectMemberPayload, 0, len(members))
	for _, m := range members {
		payloads = append(payloads, toProjMemberPayload(m))
	}
	return ListProjectMembersResponse{Members: payloads}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /flow/projects/member
// ─────────────────────────────────────────────────────────────────────────────

func AddProjectMemberRoutes() []string { return []string{"/flow/projects/member"} }

type (
	AddProjectMemberHandler struct{ FlowService *application.Service }
	AddProjectMemberRequest struct {
		ProjectID JSONID `json:"project_id"`
		UserID    JSONID `json:"user_id"`
		Role      string `json:"role"`
	}
)

type AddProjectMemberResponse struct {
	Message string `json:"message"`
}

func (h *AddProjectMemberHandler) Validate(_ context.Context, req *AddProjectMemberRequest) *errorModels.ServerError {
	if req.ProjectID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "project_id is required"}
	}
	if req.UserID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "user_id is required"}
	}
	return nil
}

func (h *AddProjectMemberHandler) Serve(ctx context.Context, req *AddProjectMemberRequest) (AddProjectMemberResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return AddProjectMemberResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !h.FlowService.IsProjectMember(ctx, authUser.UserID, req.ProjectID.Int64()) {
		return AddProjectMemberResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: req.ProjectID.Int64(),
		UserID:    req.UserID.Int64(),
		Role:      req.Role,
		AddedBy:   authUser.UserID,
	}); err != nil {
		return AddProjectMemberResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to add project member"}
	}
	return AddProjectMemberResponse{Message: "Member added"}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// DELETE /flow/projects/member
// ─────────────────────────────────────────────────────────────────────────────

func RemoveProjectMemberRoutes() []string { return []string{"/flow/projects/member"} }

type (
	RemoveProjectMemberHandler struct{ FlowService *application.Service }
	RemoveProjectMemberRequest struct {
		ProjectID JSONID `json:"project_id"`
		UserID    JSONID `json:"user_id"`
	}
)

type RemoveProjectMemberResponse struct {
	Message string `json:"message"`
}

func (h *RemoveProjectMemberHandler) Validate(_ context.Context, req *RemoveProjectMemberRequest) *errorModels.ServerError {
	if req.ProjectID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "project_id is required"}
	}
	if req.UserID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "user_id is required"}
	}
	return nil
}

func (h *RemoveProjectMemberHandler) Serve(ctx context.Context, req *RemoveProjectMemberRequest) (RemoveProjectMemberResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return RemoveProjectMemberResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !h.FlowService.IsProjectMember(ctx, authUser.UserID, req.ProjectID.Int64()) {
		return RemoveProjectMemberResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.RemoveProjectMember(ctx, domain.RemoveProjectMemberInput{
		ProjectID: req.ProjectID.Int64(),
		UserID:    req.UserID.Int64(),
	}); err != nil {
		return RemoveProjectMemberResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to remove project member"}
	}
	return RemoveProjectMemberResponse{Message: "Member removed"}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /flow/projects/member/role
// ─────────────────────────────────────────────────────────────────────────────

func UpdateProjectMemberRoleRoutes() []string { return []string{"/flow/projects/member/role"} }

type (
	UpdateProjectMemberRoleHandler struct{ FlowService *application.Service }
	UpdateProjectMemberRoleRequest struct {
		ProjectID JSONID `json:"project_id"`
		UserID    JSONID `json:"user_id"`
		Role      string `json:"role"`
	}
)

type UpdateProjectMemberRoleResponse struct {
	Message string `json:"message"`
}

func (h *UpdateProjectMemberRoleHandler) Validate(_ context.Context, req *UpdateProjectMemberRoleRequest) *errorModels.ServerError {
	if req.ProjectID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "project_id is required"}
	}
	if req.UserID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "user_id is required"}
	}
	if req.Role != "manager" && req.Role != "member" && req.Role != "viewer" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "role must be manager, member, or viewer"}
	}
	return nil
}

func (h *UpdateProjectMemberRoleHandler) Serve(ctx context.Context, req *UpdateProjectMemberRoleRequest) (UpdateProjectMemberRoleResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateProjectMemberRoleResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !h.FlowService.IsProjectMember(ctx, authUser.UserID, req.ProjectID.Int64()) {
		return UpdateProjectMemberRoleResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.UpdateProjectMemberRole(ctx, domain.UpdateProjectMemberRoleInput{
		ProjectID: req.ProjectID.Int64(),
		UserID:    req.UserID.Int64(),
		Role:      req.Role,
	}); err != nil {
		return UpdateProjectMemberRoleResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update role"}
	}
	return UpdateProjectMemberRoleResponse{Message: "Role updated"}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// PUT /flow/projects/transfer
// ─────────────────────────────────────────────────────────────────────────────

func TransferProjectOwnershipRoutes() []string { return []string{"/flow/projects/transfer"} }

type (
	TransferProjectOwnershipHandler struct{ FlowService *application.Service }
	TransferProjectOwnershipRequest struct {
		ProjectID  JSONID `json:"project_id"`
		NewOwnerID JSONID `json:"new_owner_id"`
	}
)

type TransferProjectOwnershipResponse struct {
	Message string `json:"message"`
}

func (h *TransferProjectOwnershipHandler) Validate(_ context.Context, req *TransferProjectOwnershipRequest) *errorModels.ServerError {
	if req.ProjectID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "project_id is required"}
	}
	if req.NewOwnerID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "new_owner_id is required"}
	}
	return nil
}

func (h *TransferProjectOwnershipHandler) Serve(ctx context.Context, req *TransferProjectOwnershipRequest) (TransferProjectOwnershipResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return TransferProjectOwnershipResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.TransferProjectOwnership(ctx, domain.TransferProjectOwnershipInput{
		ProjectID:  req.ProjectID.Int64(),
		NewOwnerID: req.NewOwnerID.Int64(),
		CallerID:   authUser.UserID,
	}); err != nil {
		return TransferProjectOwnershipResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to transfer ownership"}
	}
	return TransferProjectOwnershipResponse{Message: "Ownership transferred"}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// POST /flow/projects/member/invite
// ─────────────────────────────────────────────────────────────────────────────

func InviteProjectMemberByEmailRoutes() []string { return []string{"/flow/projects/member/invite"} }

type (
	InviteProjectMemberByEmailHandler struct{ FlowService *application.Service }
	InviteProjectMemberByEmailRequest struct {
		ProjectID JSONID `json:"project_id"`
		Email     string `json:"email"`
		Role      string `json:"role"`
	}
)

type InviteProjectMemberByEmailResponse struct {
	Message string `json:"message"`
}

func (h *InviteProjectMemberByEmailHandler) Validate(_ context.Context, req *InviteProjectMemberByEmailRequest) *errorModels.ServerError {
	if req.ProjectID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "project_id is required"}
	}
	if req.Email == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "email is required"}
	}
	return nil
}

func (h *InviteProjectMemberByEmailHandler) Serve(ctx context.Context, req *InviteProjectMemberByEmailRequest) (InviteProjectMemberByEmailResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return InviteProjectMemberByEmailResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !h.FlowService.IsProjectMember(ctx, authUser.UserID, req.ProjectID.Int64()) {
		return InviteProjectMemberByEmailResponse{}, errorModels.PermissionDeniedServerError()
	}

	role := req.Role
	if role == "" {
		role = "member"
	}

	found, err := h.FlowService.InviteProjectMemberByEmail(ctx, req.ProjectID.Int64(), req.Email, role, authUser.UserID)
	if err != nil {
		return InviteProjectMemberByEmailResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to invite member"}
	}
	if !found {
		return InviteProjectMemberByEmailResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, HTTPCode: http.StatusNotFound, Message: "No user found with that email"}
	}
	return InviteProjectMemberByEmailResponse{Message: "Member invited"}, nil
}
