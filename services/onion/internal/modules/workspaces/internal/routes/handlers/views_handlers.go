package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

// ─── List Views ───────────────────────────────────────────────────────────────

func ListViewsRoutes() []string { return []string{"/flow/views"} }

type ListViewsRequest struct {
	WorkspaceID string `query:"workspace_id"`
}

type ListViewsResponse struct {
	Views []viewPayload `json:"views"`
}

type viewPayload struct {
	ID      string                 `json:"id"`
	Name    string                 `json:"name"`
	Icon    string                 `json:"icon"`
	Filters map[string]interface{} `json:"filters"`
	GroupBy string                 `json:"groupBy,omitempty"`
	SortBy  string                 `json:"sortBy,omitempty"`
	SortDir string                 `json:"sortDir,omitempty"`
}

func (r *ListViewsRequest) Validate() *errorModels.ServerError { return nil }

type ListViewsHandler struct{ FlowService *application.Service }

func (h *ListViewsHandler) Serve(ctx context.Context, r *ListViewsRequest) (ListViewsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListViewsResponse{}, errorModels.PermissionDeniedServerError()
	}

	wsID := authUser.BusinessID
	if r.WorkspaceID != "" {
		if id, err := strconv.ParseInt(r.WorkspaceID, 10, 64); err == nil {
			wsID = id
		}
	}

	views, err := h.FlowService.ListViews(ctx, wsID)
	if err != nil {
		return ListViewsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	payloads := make([]viewPayload, 0, len(views))
	for _, v := range views {
		payloads = append(payloads, viewPayload{
			ID:      strconv.FormatInt(v.ID, 10),
			Name:    v.Name,
			Icon:    v.Icon,
			Filters: v.Filters,
			GroupBy: v.GroupBy,
			SortBy:  v.SortBy,
			SortDir: v.SortDir,
		})
	}
	return ListViewsResponse{Views: payloads}, nil
}

// ─── Create View ──────────────────────────────────────────────────────────────

func CreateViewRoutes() []string { return []string{"/flow/views"} }

type CreateViewRequest struct {
	WorkspaceID string                 `json:"workspaceId"`
	Name        string                 `json:"name"`
	Icon        string                 `json:"icon"`
	Filters     map[string]interface{} `json:"filters"`
	GroupBy     string                 `json:"groupBy"`
	SortBy      string                 `json:"sortBy"`
	SortDir     string                 `json:"sortDir"`
}

type CreateViewResponse struct {
	View viewPayload `json:"view"`
}

func (r *CreateViewRequest) Validate() *errorModels.ServerError {
	if r.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

type CreateViewHandler struct{ FlowService *application.Service }

func (h *CreateViewHandler) Serve(ctx context.Context, r *CreateViewRequest) (CreateViewResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateViewResponse{}, errorModels.PermissionDeniedServerError()
	}

	wsID := authUser.BusinessID
	if r.WorkspaceID != "" {
		if id, err := strconv.ParseInt(r.WorkspaceID, 10, 64); err == nil {
			wsID = id
		}
	}
	if r.Filters == nil {
		r.Filters = map[string]interface{}{}
	}

	view, err := h.FlowService.CreateView(ctx, domain.ViewInput{
		WorkspaceID: wsID,
		Name:        r.Name,
		Icon:        r.Icon,
		Filters:     r.Filters,
		GroupBy:     r.GroupBy,
		SortBy:      r.SortBy,
		SortDir:     r.SortDir,
		CreatedBy:   authUser.UserID,
	})
	if err != nil {
		return CreateViewResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}

	return CreateViewResponse{View: viewPayload{
		ID:      strconv.FormatInt(view.ID, 10),
		Name:    view.Name,
		Icon:    view.Icon,
		Filters: view.Filters,
		GroupBy: view.GroupBy,
		SortBy:  view.SortBy,
		SortDir: view.SortDir,
	}}, nil
}

// ─── Update View ──────────────────────────────────────────────────────────────

func UpdateViewRoutes() []string { return []string{"/flow/views/{id}"} }

type UpdateViewRequest struct {
	ID      string                 `path:"id"`
	Name    string                 `json:"name"`
	Icon    string                 `json:"icon"`
	Filters map[string]interface{} `json:"filters"`
	GroupBy string                 `json:"groupBy"`
	SortBy  string                 `json:"sortBy"`
	SortDir string                 `json:"sortDir"`
}

type UpdateViewResponse struct{}

func (r *UpdateViewRequest) Validate() *errorModels.ServerError { return nil }

type UpdateViewHandler struct{ FlowService *application.Service }

func (h *UpdateViewHandler) Serve(ctx context.Context, r *UpdateViewRequest) (UpdateViewResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateViewResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := strconv.ParseInt(r.ID, 10, 64)
	if err != nil || id <= 0 {
		return UpdateViewResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid view id"}
	}
	if r.Filters == nil {
		r.Filters = map[string]interface{}{}
	}

	if err := h.FlowService.UpdateView(ctx, domain.UpdateViewInput{
		ID:      id,
		ActorID: authUser.UserID,
		Name:    r.Name,
		Icon:    r.Icon,
		Filters: r.Filters,
		GroupBy: r.GroupBy,
		SortBy:  r.SortBy,
		SortDir: r.SortDir,
	}); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return UpdateViewResponse{}, errorModels.PermissionDeniedServerError()
		}
		return UpdateViewResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return UpdateViewResponse{}, nil
}

// ─── Delete View ──────────────────────────────────────────────────────────────

func DeleteViewRoutes() []string { return []string{"/flow/views/{id}"} }

type DeleteViewRequest struct {
	ID string `path:"id"`
}

type DeleteViewResponse struct{}

func (r *DeleteViewRequest) Validate() *errorModels.ServerError { return nil }

type DeleteViewHandler struct{ FlowService *application.Service }

func (h *DeleteViewHandler) Serve(ctx context.Context, r *DeleteViewRequest) (DeleteViewResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteViewResponse{}, errorModels.PermissionDeniedServerError()
	}

	id, err := strconv.ParseInt(r.ID, 10, 64)
	if err != nil || id <= 0 {
		return DeleteViewResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid view id"}
	}
	if err := h.FlowService.DeleteView(ctx, id, authUser.UserID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return DeleteViewResponse{}, errorModels.PermissionDeniedServerError()
		}
		return DeleteViewResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: err.Error()}
	}
	return DeleteViewResponse{}, nil
}
