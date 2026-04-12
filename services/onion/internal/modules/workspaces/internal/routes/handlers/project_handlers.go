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

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func ListProjectsRoutes() []string  { return []string{"/flow/projects"} }
func CreateProjectRoutes() []string { return []string{"/flow/projects"} }
func UpdateProjectRoutes() []string { return []string{"/flow/projects"} }
func DeleteProjectRoutes() []string { return []string{"/flow/projects"} }

// ─────────────────────────────────────────────
// Shared payload types
// ─────────────────────────────────────────────

type ProjectPayload struct {
	ID          int64      `json:"id"`
	WorkspaceID int64      `json:"workspace_id"`
	OwnerID     int64      `json:"owner_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Color       string     `json:"color"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedBy   int64      `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func toProjectPayload(p domain.Project) ProjectPayload {
	return ProjectPayload{
		ID:          p.ID,
		WorkspaceID: p.WorkspaceID,
		OwnerID:     p.OwnerID,
		Name:        p.Name,
		Description: p.Description,
		Icon:        p.Icon,
		Color:       p.Color,
		Status:      p.Status,
		StartDate:   p.StartDate,
		DueDate:     p.DueDate,
		CreatedBy:   p.CreatedBy,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ─────────────────────────────────────────────
// GET /flow/projects
// ─────────────────────────────────────────────

type ListProjectsHandler struct {
	FlowService *application.Service
}

type ListProjectsRequest struct {
	WorkspaceID int64 `json:"workspace_id"`
}

type ListProjectsResponse struct {
	Projects []ProjectPayload `json:"projects"`
}

func (h *ListProjectsHandler) Validate(_ context.Context, req *ListProjectsRequest) *errorModels.ServerError {
	if req.WorkspaceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "workspace_id is required"}
	}

	return nil
}

func (h *ListProjectsHandler) Serve(ctx context.Context, req *ListProjectsRequest) (ListProjectsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListProjectsResponse{}, errorModels.PermissionDeniedServerError()
	}

	projects, err := h.FlowService.ListProjects(ctx, req.WorkspaceID, authUser.UserID)
	if err != nil {
		return ListProjectsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list projects",
		}
	}

	payloads := make([]ProjectPayload, 0, len(projects))
	for _, p := range projects {
		payloads = append(payloads, toProjectPayload(p))
	}

	return ListProjectsResponse{Projects: payloads}, nil
}

// ─────────────────────────────────────────────
// POST /flow/projects
// ─────────────────────────────────────────────

type CreateProjectHandler struct {
	FlowService *application.Service
}

type CreateProjectRequest struct {
	WorkspaceID int64      `json:"workspace_id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Color       string     `json:"color"`
	StartDate   *time.Time `json:"start_date"`
	DueDate     *time.Time `json:"due_date"`
}

type CreateProjectResponse struct {
	Project ProjectPayload `json:"project"`
}

func (h *CreateProjectHandler) Validate(_ context.Context, req *CreateProjectRequest) *errorModels.ServerError {
	if req.WorkspaceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}

	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}

	return nil
}

func (h *CreateProjectHandler) Serve(ctx context.Context, req *CreateProjectRequest) (CreateProjectResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateProjectResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.ProjectInput{
		WorkspaceID: req.WorkspaceID,
		OwnerID:     authUser.UserID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		StartDate:   req.StartDate,
		DueDate:     req.DueDate,
		CreatedBy:   authUser.UserID,
	}

	project, err := h.FlowService.CreateProject(ctx, input)
	if err != nil {
		return CreateProjectResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create project",
		}
	}

	return CreateProjectResponse{Project: toProjectPayload(project)}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/projects
// ─────────────────────────────────────────────

type UpdateProjectHandler struct {
	FlowService *application.Service
}

type UpdateProjectRequest struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Color       string     `json:"color"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"start_date"`
	DueDate     *time.Time `json:"due_date"`
}

type UpdateProjectResponse struct {
	Message string `json:"message"`
}

func (h *UpdateProjectHandler) Validate(_ context.Context, req *UpdateProjectRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Project ID is required"}
	}

	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}

	return nil
}

func (h *UpdateProjectHandler) Serve(ctx context.Context, req *UpdateProjectRequest) (UpdateProjectResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateProjectResponse{}, errorModels.PermissionDeniedServerError()
	}

	status := req.Status
	if status == "" {
		status = "active"
	}

	input := domain.UpdateProjectInput{
		ID:          req.ID,
		OwnerID:     authUser.UserID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        req.Icon,
		Color:       req.Color,
		Status:      status,
		StartDate:   req.StartDate,
		DueDate:     req.DueDate,
	}

	if err := h.FlowService.UpdateProject(ctx, input); err != nil {
		return UpdateProjectResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update project",
		}
	}

	return UpdateProjectResponse{Message: "Project updated successfully"}, nil
}

// ─────────────────────────────────────────────
// DELETE /flow/projects
// ─────────────────────────────────────────────

type DeleteProjectHandler struct {
	FlowService *application.Service
}

type DeleteProjectRequest struct {
	ID int64 `json:"id"`
}

type DeleteProjectResponse struct {
	Message string `json:"message"`
}

func (h *DeleteProjectHandler) Validate(_ context.Context, req *DeleteProjectRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Project ID is required"}
	}

	return nil
}

func (h *DeleteProjectHandler) Serve(ctx context.Context, req *DeleteProjectRequest) (DeleteProjectResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteProjectResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.DeleteProject(ctx, req.ID, authUser.UserID); err != nil {
		return DeleteProjectResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete project",
		}
	}

	return DeleteProjectResponse{Message: "Project archived successfully"}, nil
}
