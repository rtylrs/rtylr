package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─────────────────────────────────────────────
// Shared payload
// ─────────────────────────────────────────────

type BusinessPayload struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}

func toBusinessPayload(b domain.Business) BusinessPayload {
	return BusinessPayload{
		ID:    b.ID,
		Name:  b.Name,
		Color: b.Color,
		Icon:  b.Icon,
	}
}

// ─────────────────────────────────────────────
// GET /flow/businesses
// ─────────────────────────────────────────────

func ListBusinessesRoutes() []string { return []string{"/flow/businesses"} }

type ListBusinessesHandler struct{ FlowService *application.Service }
type ListBusinessesRequest struct{}
type ListBusinessesResponse struct {
	Businesses []BusinessPayload `json:"businesses"`
}

func (h *ListBusinessesHandler) Serve(ctx context.Context, _ *ListBusinessesRequest) (ListBusinessesResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListBusinessesResponse{}, errorModels.PermissionDeniedServerError()
	}

	businesses, err := h.FlowService.ListBusinesses(ctx, authUser.UserID)
	if err != nil {
		return ListBusinessesResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list businesses"}
	}

	payloads := make([]BusinessPayload, 0, len(businesses))
	for _, b := range businesses {
		payloads = append(payloads, toBusinessPayload(b))
	}
	return ListBusinessesResponse{Businesses: payloads}, nil
}

// ─────────────────────────────────────────────
// POST /flow/businesses
// ─────────────────────────────────────────────

func CreateBusinessRoutes() []string { return []string{"/flow/businesses"} }

type CreateBusinessHandler struct{ FlowService *application.Service }
type CreateBusinessRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}
type CreateBusinessResponse struct {
	Business BusinessPayload `json:"business"`
}

func (h *CreateBusinessHandler) Validate(_ context.Context, req *CreateBusinessRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateBusinessHandler) Serve(ctx context.Context, req *CreateBusinessRequest) (CreateBusinessResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateBusinessResponse{}, errorModels.PermissionDeniedServerError()
	}

	b, err := h.FlowService.CreateBusiness(ctx, domain.BusinessInput{
		OwnerID: authUser.UserID,
		Name:    req.Name,
		Color:   req.Color,
		Icon:    req.Icon,
	})
	if err != nil {
		return CreateBusinessResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create business"}
	}
	return CreateBusinessResponse{Business: toBusinessPayload(b)}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/businesses
// ─────────────────────────────────────────────

func UpdateBusinessRoutes() []string { return []string{"/flow/businesses"} }

type UpdateBusinessHandler struct{ FlowService *application.Service }
type UpdateBusinessRequest struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Icon  string `json:"icon"`
}
type UpdateBusinessResponse struct{ Message string `json:"message"` }

func (h *UpdateBusinessHandler) Validate(_ context.Context, req *UpdateBusinessRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *UpdateBusinessHandler) Serve(ctx context.Context, req *UpdateBusinessRequest) (UpdateBusinessResponse, *errorModels.ServerError) {
	if _, ok := authPkg.GetAuthUser(ctx); !ok {
		return UpdateBusinessResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.UpdateBusiness(ctx, req.ID, req.Name, req.Color, req.Icon); err != nil {
		return UpdateBusinessResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update business"}
	}
	return UpdateBusinessResponse{Message: "Business updated"}, nil
}

// ─────────────────────────────────────────────
// DELETE /flow/businesses
// ─────────────────────────────────────────────

func DeleteBusinessRoutes() []string { return []string{"/flow/businesses"} }

type DeleteBusinessHandler struct{ FlowService *application.Service }
type DeleteBusinessRequest struct {
	ID int64 `json:"id"`
}
type DeleteBusinessResponse struct{ Message string `json:"message"` }

func (h *DeleteBusinessHandler) Validate(_ context.Context, req *DeleteBusinessRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *DeleteBusinessHandler) Serve(ctx context.Context, req *DeleteBusinessRequest) (DeleteBusinessResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteBusinessResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.DeleteBusiness(ctx, req.ID, authUser.UserID); err != nil {
		return DeleteBusinessResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete business"}
	}
	return DeleteBusinessResponse{Message: "Business deleted"}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/workspaces/business
// ─────────────────────────────────────────────

func SetWorkspaceBusinessRoutes() []string { return []string{"/flow/workspaces/business"} }

type SetWorkspaceBusinessHandler struct{ FlowService *application.Service }
type SetWorkspaceBusinessRequest struct {
	WorkspaceID int64  `json:"workspace_id"`
	BusinessID  *int64 `json:"business_id"`
}
type SetWorkspaceBusinessResponse struct{ Message string `json:"message"` }

func (h *SetWorkspaceBusinessHandler) Validate(_ context.Context, req *SetWorkspaceBusinessRequest) *errorModels.ServerError {
	if req.WorkspaceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "workspace_id is required"}
	}
	return nil
}

func (h *SetWorkspaceBusinessHandler) Serve(ctx context.Context, req *SetWorkspaceBusinessRequest) (SetWorkspaceBusinessResponse, *errorModels.ServerError) {
	if _, ok := authPkg.GetAuthUser(ctx); !ok {
		return SetWorkspaceBusinessResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.SetWorkspaceBusiness(ctx, req.WorkspaceID, req.BusinessID); err != nil {
		return SetWorkspaceBusinessResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to set workspace business"}
	}
	return SetWorkspaceBusinessResponse{Message: "Workspace business updated"}, nil
}
