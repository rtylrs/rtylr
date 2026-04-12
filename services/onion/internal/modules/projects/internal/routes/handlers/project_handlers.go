package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/projects/internal/application"
	"github.com/voxire/onion/internal/modules/projects/internal/domain"
)

type ProjectPayload struct {
	ID          int64  `json:"id"`
	OwnerID     int64  `json:"owner_id"`
	WorkspaceID *int64 `json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Color       string `json:"color"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func newProjectPayload(p domain.Project) ProjectPayload {
	return ProjectPayload{
		ID:          p.ID,
		OwnerID:     p.OwnerID,
		WorkspaceID: p.WorkspaceID,
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status,
		Color:       p.Color,
		CreatedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─── List ─────────────────────────────────────────────────────────────────────

func ProjectListRoutes() []string { return []string{"/flow/projects"} }

type ListProjectsHandler struct{ Service *application.Service }
type ListProjectsRequest struct{}
type ListProjectsResponse struct {
	Projects []ProjectPayload `json:"projects"`
}

func (h *ListProjectsHandler) Serve(ctx context.Context, _ *ListProjectsRequest) (ListProjectsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListProjectsResponse{}, errorModels.PermissionDeniedServerError()
	}
	projects, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListProjectsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list projects"}
	}
	payloads := make([]ProjectPayload, len(projects))
	for i, p := range projects {
		payloads[i] = newProjectPayload(p)
	}
	return ListProjectsResponse{Projects: payloads}, nil
}

// ─── Create ───────────────────────────────────────────────────────────────────

func ProjectCreateRoutes() []string { return []string{"/flow/projects"} }

type CreateProjectHandler struct{ Service *application.Service }

type CreateProjectRequest struct {
	WorkspaceID *int64 `json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Color       string `json:"color"`
}

type CreateProjectResponse struct {
	ProjectID int64  `json:"project_id"`
	Message   string `json:"message"`
}

func (h *CreateProjectHandler) Validate(_ context.Context, req *CreateProjectRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *CreateProjectHandler) Serve(ctx context.Context, req *CreateProjectRequest) (CreateProjectResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateProjectResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.Create(ctx, domain.ProjectInput{
		OwnerID:     device.User.UserID,
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	})
	if err != nil {
		return CreateProjectResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create project"}
	}
	return CreateProjectResponse{ProjectID: id, Message: "Project created successfully"}, nil
}

// ─── Update ───────────────────────────────────────────────────────────────────

func ProjectUpdateRoutes() []string { return []string{"/flow/projects/{id}"} }

type UpdateProjectHandler struct{ Service *application.Service }

type UpdateProjectRequest struct {
	WorkspaceID *int64 `json:"workspace_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Color       string `json:"color"`
}

type UpdateProjectResponse struct {
	Message string `json:"message"`
}

func (h *UpdateProjectHandler) Validate(_ context.Context, req *UpdateProjectRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *UpdateProjectHandler) Serve(ctx context.Context, req *UpdateProjectRequest) (UpdateProjectResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateProjectResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateProjectResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid project ID"}
	}
	status := req.Status
	if status == "" {
		status = "active"
	}
	if updateErr := h.Service.Update(ctx, domain.ProjectUpdateInput{
		ID:          id,
		OwnerID:     device.User.UserID,
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		Description: req.Description,
		Status:      status,
		Color:       req.Color,
	}); updateErr != nil {
		if errors.Is(updateErr, domain.ErrProjectNotFound) {
			return UpdateProjectResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Project not found"}
		}
		return UpdateProjectResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update project"}
	}
	return UpdateProjectResponse{Message: "Project updated successfully"}, nil
}

// ─── Delete ───────────────────────────────────────────────────────────────────

func ProjectDeleteRoutes() []string { return []string{"/flow/projects/{id}"} }

type DeleteProjectHandler struct{ Service *application.Service }
type DeleteProjectRequest struct{}
type DeleteProjectResponse struct {
	Message string `json:"message"`
}

func (h *DeleteProjectHandler) Serve(ctx context.Context, _ *DeleteProjectRequest) (DeleteProjectResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteProjectResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteProjectResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid project ID"}
	}
	if deleteErr := h.Service.Delete(ctx, device.User.UserID, id); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrProjectNotFound) {
			return DeleteProjectResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Project not found"}
		}
		return DeleteProjectResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete project"}
	}
	return DeleteProjectResponse{Message: "Project deleted successfully"}, nil
}
