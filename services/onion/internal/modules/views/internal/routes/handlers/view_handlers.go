package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/views/internal/application"
	"github.com/voxire/onion/internal/modules/views/internal/domain"
)

type ViewPayload struct {
	ID          int64           `json:"id"`
	OwnerID     int64           `json:"ownerId"`
	WorkspaceID int64           `json:"workspaceId"`
	Name        string          `json:"name"`
	Filters     json.RawMessage `json:"filters"`
	SortBy      string          `json:"sortBy"`
	GroupBy     string          `json:"groupBy"`
	IsDefault   bool            `json:"isDefault"`
	CreatedAt   string          `json:"createdAt"`
	UpdatedAt   string          `json:"updatedAt"`
}

func newViewPayload(v domain.View) ViewPayload {
	filters := json.RawMessage(v.Filters)
	if filters == nil {
		filters = json.RawMessage("{}")
	}
	return ViewPayload{
		ID:          v.ID,
		OwnerID:     v.OwnerID,
		WorkspaceID: v.WorkspaceID,
		Name:        v.Name,
		Filters:     filters,
		SortBy:      v.SortBy,
		GroupBy:     v.GroupBy,
		IsDefault:   v.IsDefault,
		CreatedAt:   v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   v.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- List Views ---

func ViewListRoutes() []string { return []string{"/flow/views"} }

type ListViewsHandler struct{ Service *application.Service }
type ListViewsRequest struct {
	WorkspaceID int64 `json:"workspaceId"`
}
type ListViewsResponse struct {
	Views []ViewPayload `json:"views"`
}

func (h *ListViewsHandler) Validate(_ context.Context, req *ListViewsRequest) *errorModels.ServerError {
	if req.WorkspaceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}
	return nil
}

func (h *ListViewsHandler) Serve(ctx context.Context, req *ListViewsRequest) (ListViewsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListViewsResponse{}, errorModels.PermissionDeniedServerError()
	}
	views, err := h.Service.List(ctx, device.User.UserID, req.WorkspaceID)
	if err != nil {
		return ListViewsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list views"}
	}
	payloads := make([]ViewPayload, len(views))
	for i, v := range views {
		payloads[i] = newViewPayload(v)
	}
	return ListViewsResponse{Views: payloads}, nil
}

// --- Create View ---

func ViewCreateRoutes() []string { return []string{"/flow/views"} }

type CreateViewHandler struct{ Service *application.Service }

type CreateViewRequest struct {
	WorkspaceID int64           `json:"workspaceId"`
	Name        string          `json:"name"`
	Filters     json.RawMessage `json:"filters"`
	SortBy      string          `json:"sortBy"`
	GroupBy     string          `json:"groupBy"`
	IsDefault   bool            `json:"isDefault"`
}

type CreateViewResponse struct {
	ViewID  int64  `json:"viewId"`
	Message string `json:"message"`
}

func (h *CreateViewHandler) Validate(_ context.Context, req *CreateViewRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.WorkspaceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}
	return nil
}

func (h *CreateViewHandler) Serve(ctx context.Context, req *CreateViewRequest) (CreateViewResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateViewResponse{}, errorModels.PermissionDeniedServerError()
	}
	filters := []byte(req.Filters)
	if filters == nil {
		filters = []byte("{}")
	}
	input := domain.ViewInput{
		OwnerID:     device.User.UserID,
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		Filters:     filters,
		SortBy:      req.SortBy,
		GroupBy:     req.GroupBy,
		IsDefault:   req.IsDefault,
	}
	id, err := h.Service.Create(ctx, input)
	if err != nil {
		return CreateViewResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create view"}
	}
	return CreateViewResponse{ViewID: id, Message: "View created successfully"}, nil
}

// --- Update View ---

func ViewUpdateRoutes() []string { return []string{"/flow/views/{id}"} }

type UpdateViewHandler struct{ Service *application.Service }

type UpdateViewRequest struct {
	Name      string          `json:"name"`
	Filters   json.RawMessage `json:"filters"`
	SortBy    string          `json:"sortBy"`
	GroupBy   string          `json:"groupBy"`
	IsDefault bool            `json:"isDefault"`
}

type UpdateViewResponse struct {
	Message string `json:"message"`
}

func (h *UpdateViewHandler) Validate(_ context.Context, req *UpdateViewRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateViewHandler) Serve(ctx context.Context, req *UpdateViewRequest) (UpdateViewResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateViewResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateViewResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid view ID"}
	}
	filters := []byte(req.Filters)
	if filters == nil {
		filters = []byte("{}")
	}
	input := domain.ViewUpdateInput{
		ID:        id,
		OwnerID:   device.User.UserID,
		Name:      req.Name,
		Filters:   filters,
		SortBy:    req.SortBy,
		GroupBy:   req.GroupBy,
		IsDefault: req.IsDefault,
	}
	if updateErr := h.Service.Update(ctx, input); updateErr != nil {
		if errors.Is(updateErr, domain.ErrViewNotFound) {
			return UpdateViewResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "View not found"}
		}
		return UpdateViewResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update view"}
	}
	return UpdateViewResponse{Message: "View updated successfully"}, nil
}

// --- Delete View ---

func ViewDeleteRoutes() []string { return []string{"/flow/views/{id}"} }

type DeleteViewHandler struct{ Service *application.Service }
type DeleteViewRequest struct{}
type DeleteViewResponse struct {
	Message string `json:"message"`
}

func (h *DeleteViewHandler) Serve(ctx context.Context, _ *DeleteViewRequest) (DeleteViewResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteViewResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteViewResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid view ID"}
	}
	if deleteErr := h.Service.Delete(ctx, id, device.User.UserID); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrViewNotFound) {
			return DeleteViewResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "View not found"}
		}
		return DeleteViewResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete view"}
	}
	return DeleteViewResponse{Message: "View deleted successfully"}, nil
}
