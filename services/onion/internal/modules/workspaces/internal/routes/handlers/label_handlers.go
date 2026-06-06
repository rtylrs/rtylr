package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	flowlabelv1 "github.com/voxire/voxie/gen/go/flow/label/v1"
)

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func ListLabelsRoutes() []string    { return []string{"/flow/labels"} }
func CreateLabelRoutes() []string   { return []string{"/flow/labels"} }
func UpdateLabelRoutes() []string   { return []string{"/flow/labels"} }
func DeleteLabelRoutes() []string   { return []string{"/flow/labels"} }
func SetTaskLabelsRoutes() []string { return []string{"/flow/tasks/labels"} }

// ─────────────────────────────────────────────
// Shared payload types
// ─────────────────────────────────────────────

func toLabelPayload(l domain.Label) *flowlabelv1.Label {
	return &flowlabelv1.Label{
		Id:          l.ID,
		WorkspaceId: l.WorkspaceID,
		Name:        l.Name,
		Color:       l.Color,
		CreatedAt:   l.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─────────────────────────────────────────────
// GET /flow/labels
// ─────────────────────────────────────────────

type ListLabelsHandler struct {
	FlowService *application.Service
}

type ListLabelsRequest struct {
	WorkspaceID WorkspaceIDField `json:"workspace_id"`
}

type ListLabelsResponse struct {
	Labels []*flowlabelv1.Label `json:"labels"`
}

func (h *ListLabelsHandler) Validate(_ context.Context, _ *ListLabelsRequest) *errorModels.ServerError {
	// Missing workspace_id is OK — we fall back to the caller's personal
	// workspace below. This matches the ergonomic the frontend already
	// assumed via the deprecated `getLabels()` barrel method.
	return nil
}

func (h *ListLabelsHandler) Serve(ctx context.Context, req *ListLabelsRequest) (ListLabelsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListLabelsResponse{}, errorModels.PermissionDeniedServerError()
	}

	// Empty or "personal" → the caller's personal workspace. Explicit
	// numeric → use it. Anything else is a bad request.
	raw := req.WorkspaceID.String()
	if raw == "" {
		raw = "personal"
	}
	workspaceID, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, raw)
	if err != nil || workspaceID == 0 {
		return ListLabelsResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid workspace_id",
		}
	}

	labels, err := h.FlowService.ListLabels(ctx, workspaceID)
	if err != nil {
		return ListLabelsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list labels",
		}
	}

	payloads := make([]*flowlabelv1.Label, 0, len(labels))
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
	WorkspaceID WorkspaceIDField `json:"workspace_id"`
	Name        string           `json:"name"`
	Color       string           `json:"color"`
}

type CreateLabelResponse struct {
	Label *flowlabelv1.Label `json:"label"`
}

func (h *CreateLabelHandler) Validate(_ context.Context, req *CreateLabelRequest) *errorModels.ServerError {
	if req.WorkspaceID.IsEmpty() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}

	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}

	return nil
}

func (h *CreateLabelHandler) Serve(ctx context.Context, req *CreateLabelRequest) (CreateLabelResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateLabelResponse{}, errorModels.PermissionDeniedServerError()
	}

	workspaceID, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, req.WorkspaceID.String())
	if err != nil || workspaceID == 0 {
		return CreateLabelResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid workspace_id",
		}
	}

	input := domain.LabelInput{
		WorkspaceID: workspaceID,
		Name:        req.Name,
		Color:       req.Color,
	}

	label, createErr := h.FlowService.CreateLabel(ctx, input)
	if createErr != nil {
		return CreateLabelResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create label",
		}
	}

	return CreateLabelResponse{Label: toLabelPayload(label)}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/labels
// ─────────────────────────────────────────────

type UpdateLabelHandler struct {
	FlowService *application.Service
}

type UpdateLabelRequest struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type UpdateLabelResponse struct {
	Message string `json:"message"`
}

func (h *UpdateLabelHandler) Validate(_ context.Context, req *UpdateLabelRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *UpdateLabelHandler) Serve(ctx context.Context, req *UpdateLabelRequest) (UpdateLabelResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateLabelResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.LabelInput{
		Name:  req.Name,
		Color: req.Color,
	}

	if err := h.FlowService.UpdateLabel(ctx, req.ID, authUser.UserID, input); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return UpdateLabelResponse{}, errorModels.PermissionDeniedServerError()
		}
		return UpdateLabelResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update label",
		}
	}

	return UpdateLabelResponse{Message: "Label updated"}, nil
}

// ─────────────────────────────────────────────
// DELETE /flow/labels
// ─────────────────────────────────────────────

type DeleteLabelHandler struct {
	FlowService *application.Service
}

type DeleteLabelRequest struct {
	ID int64 `json:"id"`
}

type DeleteLabelResponse struct {
	Message string `json:"message"`
}

func (h *DeleteLabelHandler) Validate(_ context.Context, req *DeleteLabelRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *DeleteLabelHandler) Serve(ctx context.Context, req *DeleteLabelRequest) (DeleteLabelResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteLabelResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.DeleteLabel(ctx, req.ID, authUser.UserID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return DeleteLabelResponse{}, errorModels.PermissionDeniedServerError()
		}
		return DeleteLabelResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete label",
		}
	}

	return DeleteLabelResponse{Message: "Label deleted"}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/tasks/labels
// ─────────────────────────────────────────────

type SetTaskLabelsHandler struct {
	FlowService *application.Service
}

type SetTaskLabelsRequest struct {
	TaskID   int64   `json:"task_id"`
	LabelIDs []int64 `json:"label_ids"`
}

type SetTaskLabelsResponse struct {
	Message string `json:"message"`
}

func (h *SetTaskLabelsHandler) Validate(_ context.Context, req *SetTaskLabelsRequest) *errorModels.ServerError {
	if req.TaskID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "task_id is required"}
	}
	return nil
}

func (h *SetTaskLabelsHandler) Serve(ctx context.Context, req *SetTaskLabelsRequest) (SetTaskLabelsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return SetTaskLabelsResponse{}, errorModels.PermissionDeniedServerError()
	}

	labelIDs := req.LabelIDs
	if labelIDs == nil {
		labelIDs = []int64{}
	}

	if err := h.FlowService.SetTaskLabels(ctx, req.TaskID, authUser.UserID, labelIDs); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return SetTaskLabelsResponse{}, errorModels.PermissionDeniedServerError()
		}
		return SetTaskLabelsResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to set task labels",
		}
	}

	return SetTaskLabelsResponse{Message: "Task labels updated"}, nil
}
