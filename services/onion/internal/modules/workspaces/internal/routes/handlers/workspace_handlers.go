package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/wsbus"
	flowworkspacev1 "github.com/voxire/voxie/gen/go/flow/workspace/v1"
)

// publishWorkspaceMemberEvent fans member add/remove events to the
// workspace topic so every connected peer refreshes its member list +
// sidebar avatar stack. The affected user ID is carried in entity_id so
// the client can also target per-user invalidations if it wants.
func publishWorkspaceMemberEvent(eventType string, workspaceID, memberUserID, actorID int64) {
	if workspaceID == 0 {
		return
	}
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("workspace:%d", workspaceID),
		Type:     eventType,
		EntityID: memberUserID,
		ActorID:  actorID,
	})
}

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func ListWorkspacesRoutes() []string        { return []string{"/flow/workspaces"} }
func CreateWorkspaceRoutes() []string       { return []string{"/flow/workspaces"} }
func UpdateWorkspaceRoutes() []string       { return []string{"/flow/workspaces"} }
func DeleteWorkspaceRoutes() []string       { return []string{"/flow/workspaces"} }
func AddWorkspaceMemberRoutes() []string    { return []string{"/flow/workspaces/member"} }
func RemoveWorkspaceMemberRoutes() []string { return []string{"/flow/workspaces/member"} }

// ─────────────────────────────────────────────
// Shared payload types
// ─────────────────────────────────────────────

func toWorkspacePayload(w domain.Workspace) *flowworkspacev1.Workspace {
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

// ─────────────────────────────────────────────
// GET /flow/workspaces
// ─────────────────────────────────────────────

type ListWorkspacesHandler struct {
	FlowService *application.Service
}

type ListWorkspacesRequest struct{}

type ListWorkspacesResponse struct {
	Workspaces []*flowworkspacev1.Workspace `json:"workspaces"`
}

func (h *ListWorkspacesHandler) Serve(ctx context.Context, _ *ListWorkspacesRequest) (ListWorkspacesResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListWorkspacesResponse{}, errorModels.PermissionDeniedServerError()
	}

	workspaces, err := h.FlowService.ListWorkspaces(ctx, authUser.UserID)
	if err != nil {
		return ListWorkspacesResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list workspaces",
		}
	}

	payloads := make([]*flowworkspacev1.Workspace, 0, len(workspaces))
	for _, w := range workspaces {
		payloads = append(payloads, toWorkspacePayload(w))
	}

	return ListWorkspacesResponse{Workspaces: payloads}, nil
}

// ─────────────────────────────────────────────
// POST /flow/workspaces
// ─────────────────────────────────────────────

type CreateWorkspaceHandler struct {
	FlowService *application.Service
}

type CreateWorkspaceRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

type CreateWorkspaceResponse struct {
	Workspace *flowworkspacev1.Workspace `json:"workspace"`
}

func (h *CreateWorkspaceHandler) Validate(_ context.Context, req *CreateWorkspaceRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Name is required",
		}
	}

	return nil
}

func (h *CreateWorkspaceHandler) Serve(ctx context.Context, req *CreateWorkspaceRequest) (CreateWorkspaceResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateWorkspaceResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.WorkspaceInput{
		OwnerID:     authUser.UserID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		CreatedBy:   authUser.UserID,
	}

	workspace, err := h.FlowService.CreateWorkspace(ctx, input)
	if err != nil {
		return CreateWorkspaceResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create workspace",
		}
	}

	return CreateWorkspaceResponse{Workspace: toWorkspacePayload(workspace)}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/workspaces
// ─────────────────────────────────────────────

type UpdateWorkspaceHandler struct {
	FlowService *application.Service
}

type UpdateWorkspaceRequest struct {
	ID          JSONID `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
}

type UpdateWorkspaceResponse struct {
	Message string `json:"message"`
}

func (h *UpdateWorkspaceHandler) Validate(_ context.Context, req *UpdateWorkspaceRequest) *errorModels.ServerError {
	if req.ID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}

	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}

	return nil
}

func (h *UpdateWorkspaceHandler) Serve(ctx context.Context, req *UpdateWorkspaceRequest) (UpdateWorkspaceResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateWorkspaceResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.UpdateWorkspaceInput{
		ID:          req.ID.Int64(),
		OwnerID:     authUser.UserID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
	}

	if err := h.FlowService.UpdateWorkspace(ctx, input); err != nil {
		return UpdateWorkspaceResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update workspace",
		}
	}

	return UpdateWorkspaceResponse{Message: "Workspace updated successfully"}, nil
}

// ─────────────────────────────────────────────
// DELETE /flow/workspaces
// ─────────────────────────────────────────────

type DeleteWorkspaceHandler struct {
	FlowService *application.Service
}

type DeleteWorkspaceRequest struct {
	ID JSONID `json:"id"`
}

type DeleteWorkspaceResponse struct {
	Message string `json:"message"`
}

func (h *DeleteWorkspaceHandler) Validate(_ context.Context, req *DeleteWorkspaceRequest) *errorModels.ServerError {
	if req.ID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}

	return nil
}

func (h *DeleteWorkspaceHandler) Serve(ctx context.Context, req *DeleteWorkspaceRequest) (DeleteWorkspaceResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteWorkspaceResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.DeleteWorkspace(ctx, req.ID.Int64(), authUser.UserID); err != nil {
		return DeleteWorkspaceResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete workspace",
		}
	}

	return DeleteWorkspaceResponse{Message: "Workspace deleted successfully"}, nil
}

// ─────────────────────────────────────────────
// POST /flow/workspaces/member
// ─────────────────────────────────────────────

type AddWorkspaceMemberHandler struct {
	FlowService *application.Service
}

type AddWorkspaceMemberRequest struct {
	WorkspaceID JSONID `json:"workspace_id"`
	UserID      JSONID `json:"user_id"`
	Role        string `json:"role"`
}

type AddWorkspaceMemberResponse struct {
	Message string `json:"message"`
}

func (h *AddWorkspaceMemberHandler) Validate(_ context.Context, req *AddWorkspaceMemberRequest) *errorModels.ServerError {
	if req.WorkspaceID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}

	if req.UserID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "User ID is required"}
	}

	return nil
}

func (h *AddWorkspaceMemberHandler) Serve(ctx context.Context, req *AddWorkspaceMemberRequest) (AddWorkspaceMemberResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return AddWorkspaceMemberResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !h.FlowService.IsWorkspaceMember(ctx, authUser.UserID, req.WorkspaceID.Int64()) {
		return AddWorkspaceMemberResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.AddWorkspaceMemberInput{
		WorkspaceID: req.WorkspaceID.Int64(),
		UserID:      req.UserID.Int64(),
		Role:        req.Role,
	}

	if err := h.FlowService.AddWorkspaceMember(ctx, input); err != nil {
		return AddWorkspaceMemberResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to add workspace member",
		}
	}

	publishWorkspaceMemberEvent(
		"workspace.member_added",
		req.WorkspaceID.Int64(),
		req.UserID.Int64(),
		authUser.UserID,
	)
	return AddWorkspaceMemberResponse{Message: "Member added successfully"}, nil
}

// ─────────────────────────────────────────────
// DELETE /flow/workspaces/member
// ─────────────────────────────────────────────

type RemoveWorkspaceMemberHandler struct {
	FlowService *application.Service
}

type RemoveWorkspaceMemberRequest struct {
	WorkspaceID JSONID `json:"workspace_id"`
	UserID      JSONID `json:"user_id"`
}

type RemoveWorkspaceMemberResponse struct {
	Message string `json:"message"`
}

func (h *RemoveWorkspaceMemberHandler) Validate(_ context.Context, req *RemoveWorkspaceMemberRequest) *errorModels.ServerError {
	if req.WorkspaceID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}

	if req.UserID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "User ID is required"}
	}

	return nil
}

func (h *RemoveWorkspaceMemberHandler) Serve(ctx context.Context, req *RemoveWorkspaceMemberRequest) (RemoveWorkspaceMemberResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return RemoveWorkspaceMemberResponse{}, errorModels.PermissionDeniedServerError()
	}
	if !h.FlowService.IsWorkspaceMember(ctx, authUser.UserID, req.WorkspaceID.Int64()) {
		return RemoveWorkspaceMemberResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.RemoveWorkspaceMemberInput{
		WorkspaceID: req.WorkspaceID.Int64(),
		UserID:      req.UserID.Int64(),
	}

	if err := h.FlowService.RemoveWorkspaceMember(ctx, input); err != nil {
		return RemoveWorkspaceMemberResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to remove workspace member",
		}
	}

	publishWorkspaceMemberEvent(
		"workspace.member_removed",
		req.WorkspaceID.Int64(),
		req.UserID.Int64(),
		authUser.UserID,
	)
	// Also notify the evicted user directly so their own UI ejects them.
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("user:%d", req.UserID.Int64()),
		Type:     "workspace.evicted",
		EntityID: req.WorkspaceID.Int64(),
		ActorID:  authUser.UserID,
	})
	return RemoveWorkspaceMemberResponse{Message: "Member removed successfully"}, nil
}
