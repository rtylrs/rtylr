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

func ListLabelsRoutes() []string  { return []string{"/flow/labels"} }
func CreateLabelRoutes() []string { return []string{"/flow/labels"} }

// ─────────────────────────────────────────────
// Shared payload types
// ─────────────────────────────────────────────

type LabelPayload struct {
	ID          int64     `json:"id"`
	WorkspaceID int64     `json:"workspaceId"`
	Name        string    `json:"name"`
	Color       string    `json:"color"`
	CreatedAt   time.Time `json:"createdAt"`
}

func toLabelPayload(l domain.Label) LabelPayload {
	return LabelPayload{
		ID:          l.ID,
		WorkspaceID: l.WorkspaceID,
		Name:        l.Name,
		Color:       l.Color,
		CreatedAt:   l.CreatedAt,
	}
}

// ─────────────────────────────────────────────
// GET /flow/labels
// ─────────────────────────────────────────────

type ListLabelsHandler struct {
	FlowService *application.Service
}

type ListLabelsRequest struct {
	WorkspaceID int64 `json:"workspace_id"`
}

type ListLabelsResponse struct {
	Labels []LabelPayload `json:"labels"`
}

func (h *ListLabelsHandler) Validate(_ context.Context, req *ListLabelsRequest) *errorModels.ServerError {
	if req.WorkspaceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "workspace_id is required"}
	}

	return nil
}

func (h *ListLabelsHandler) Serve(ctx context.Context, req *ListLabelsRequest) (ListLabelsResponse, *errorModels.ServerError) {
	_, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListLabelsResponse{}, errorModels.PermissionDeniedServerError()
	}

	labels, err := h.FlowService.ListLabels(ctx, req.WorkspaceID)
	if err != nil {
		return ListLabelsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list labels",
		}
	}

	payloads := make([]LabelPayload, 0, len(labels))
	for _, l := range labels {
		payloads = append(payloads, toLabelPayload(l))
	}

	return ListLabelsResponse{Labels: payloads}, nil
}

// ─────────────────────────────────────────────
// POST /flow/labels
// ─────────────────────────────────────────────

type CreateLabelHandler struct {
	FlowService *application.Service
}

type CreateLabelRequest struct {
	WorkspaceID int64  `json:"workspaceId"`
	Name        string `json:"name"`
	Color       string `json:"color"`
}

type CreateLabelResponse struct {
	Label LabelPayload `json:"label"`
}

func (h *CreateLabelHandler) Validate(_ context.Context, req *CreateLabelRequest) *errorModels.ServerError {
	if req.WorkspaceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}

	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}

	return nil
}

func (h *CreateLabelHandler) Serve(ctx context.Context, req *CreateLabelRequest) (CreateLabelResponse, *errorModels.ServerError) {
	_, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateLabelResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.LabelInput{
		WorkspaceID: req.WorkspaceID,
		Name:        req.Name,
		Color:       req.Color,
	}

	label, err := h.FlowService.CreateLabel(ctx, input)
	if err != nil {
		return CreateLabelResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create label",
		}
	}

	return CreateLabelResponse{Label: toLabelPayload(label)}, nil
}
