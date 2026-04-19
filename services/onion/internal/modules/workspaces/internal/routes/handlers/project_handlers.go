package handlers

import (
	"context"
	"net/http"
	"time"

	flowv1 "github.com/voxire/voxie/gen/go/flow/v1"

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

func toProjectPayload(p domain.Project) *flowv1.Project {
	proj := &flowv1.Project{
		Id:             p.ID,
		WorkspaceId:    p.WorkspaceID,
		OwnerId:        p.OwnerID,
		Name:           p.Name,
		Description:    p.Description,
		Icon:           p.Icon,
		Color:          p.Color,
		Status:         p.Status,
		CreatedBy:      p.CreatedBy,
		CreatedAt:      p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      p.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		TaskCount:      p.TaskCount,
		CompletedCount: p.CompletedCount,
		MemberCount:    p.MemberCount,
		Code:           p.Code,
	}
	if p.StartDate != nil {
		proj.StartDate = p.StartDate.Format("2006-01-02")
	}
	if p.DueDate != nil {
		proj.DueDate = p.DueDate.Format("2006-01-02")
	}
	return proj
}

// ─────────────────────────────────────────────
// GET /flow/projects
// ─────────────────────────────────────────────

type ListProjectsHandler struct {
	FlowService *application.Service
}

type ListProjectsRequest struct {
	// Accepts a numeric ID, a numeric string, or the sentinel "personal".
	WorkspaceID WorkspaceIDField `json:"workspace_id"`
}

type ListProjectsResponse struct {
	Projects []*flowv1.Project `json:"projects"`
}

func (h *ListProjectsHandler) Validate(_ context.Context, req *ListProjectsRequest) *errorModels.ServerError {
	if req.WorkspaceID.IsEmpty() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "workspace_id is required"}
	}

	return nil
}

func (h *ListProjectsHandler) Serve(ctx context.Context, req *ListProjectsRequest) (ListProjectsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListProjectsResponse{}, errorModels.PermissionDeniedServerError()
	}

	workspaceID, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, req.WorkspaceID.String())
	if err != nil || workspaceID == 0 {
		return ListProjectsResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid workspace_id",
		}
	}

	projects, err := h.FlowService.ListProjects(ctx, workspaceID, authUser.UserID)
	if err != nil {
		return ListProjectsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list projects",
		}
	}

	payloads := make([]*flowv1.Project, 0, len(projects))
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
	// Accepts a numeric ID, a numeric string, or the sentinel "personal".
	WorkspaceID WorkspaceIDField `json:"workspace_id"`
	Name        string           `json:"name"`
	Description string     `json:"description"`
	Icon        string     `json:"icon"`
	Color       string     `json:"color"`
	StartDate   *time.Time `json:"start_date"`
	DueDate     *time.Time `json:"due_date"`
}

type CreateProjectResponse struct {
	Project *flowv1.Project `json:"project"`
}

func (h *CreateProjectHandler) Validate(_ context.Context, req *CreateProjectRequest) *errorModels.ServerError {
	if req.WorkspaceID.IsEmpty() {
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

	workspaceID, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, req.WorkspaceID.String())
	if err != nil || workspaceID == 0 {
		return CreateProjectResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid workspace_id",
		}
	}

	input := domain.ProjectInput{
		WorkspaceID: workspaceID,
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
	ID          JSONID     `json:"id"`
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
	if req.ID.IsZero() {
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
		ID:          req.ID.Int64(),
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
	ID JSONID `json:"id"`
}

type DeleteProjectResponse struct {
	Message string `json:"message"`
}

func (h *DeleteProjectHandler) Validate(_ context.Context, req *DeleteProjectRequest) *errorModels.ServerError {
	if req.ID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Project ID is required"}
	}

	return nil
}

func (h *DeleteProjectHandler) Serve(ctx context.Context, req *DeleteProjectRequest) (DeleteProjectResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteProjectResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.DeleteProject(ctx, req.ID.Int64(), authUser.UserID); err != nil {
		return DeleteProjectResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete project",
		}
	}

	return DeleteProjectResponse{Message: "Project archived successfully"}, nil
}
