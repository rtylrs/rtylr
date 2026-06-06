package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/wsbus"
)

// publishColumnEvent fans a column change to the owning project's topic.
// Consumers (the kanban board) invalidate `['columns', projectId]` and
// redraw. Pure notification — no entity data on the wire.
func publishColumnEvent(eventType string, projectID, columnID, actorID int64) {
	if projectID == 0 {
		return
	}
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("project:%d", projectID),
		Type:     eventType,
		EntityID: columnID,
		ActorID:  actorID,
	})
}

// ─────────────────────────────────────────────
// Shared payload
// ─────────────────────────────────────────────

type ColumnPayload struct {
	ID          int64  `json:"id"`
	ProjectID   int64  `json:"projectId"`
	WorkspaceID int64  `json:"workspaceId"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	SortOrder   int    `json:"sortOrder"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

func toColumnPayload(c domain.ProjectColumn) ColumnPayload {
	return ColumnPayload{
		ID:          c.ID,
		ProjectID:   c.ProjectID,
		WorkspaceID: c.WorkspaceID,
		Name:        c.Name,
		Color:       c.Color,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─────────────────────────────────────────────
// GET /flow/projects/columns
// ─────────────────────────────────────────────

func ListColumnsRoutes() []string { return []string{"/flow/projects/columns"} }

type ListColumnsHandler struct{ FlowService *application.Service }

type ListColumnsRequest struct {
	ProjectID JSONID `json:"project_id"`
}

type ListColumnsResponse struct {
	Columns []ColumnPayload `json:"columns"`
}

func (h *ListColumnsHandler) Validate(_ context.Context, req *ListColumnsRequest) *errorModels.ServerError {
	if req.ProjectID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "project_id is required"}
	}
	return nil
}

func (h *ListColumnsHandler) Serve(ctx context.Context, req *ListColumnsRequest) (ListColumnsResponse, *errorModels.ServerError) {
	cols, err := h.FlowService.GetColumnsByProject(ctx, req.ProjectID.Int64())
	if err != nil {
		return ListColumnsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list columns"}
	}
	payloads := make([]ColumnPayload, 0, len(cols))
	for _, c := range cols {
		payloads = append(payloads, toColumnPayload(c))
	}
	return ListColumnsResponse{Columns: payloads}, nil
}

// ─────────────────────────────────────────────
// POST /flow/projects/columns
// ─────────────────────────────────────────────

func CreateColumnRoutes() []string { return []string{"/flow/projects/columns"} }

type CreateColumnHandler struct{ FlowService *application.Service }

type CreateColumnRequest struct {
	ProjectID   JSONID `json:"project_id"`
	WorkspaceID JSONID `json:"workspace_id"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	SortOrder   int    `json:"sort_order"`
}

type CreateColumnResponse struct {
	Column ColumnPayload `json:"column"`
}

func (h *CreateColumnHandler) Validate(_ context.Context, req *CreateColumnRequest) *errorModels.ServerError {
	if req.ProjectID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "project_id is required"}
	}
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *CreateColumnHandler) Serve(ctx context.Context, req *CreateColumnRequest) (CreateColumnResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateColumnResponse{}, errorModels.PermissionDeniedServerError()
	}

	color := req.Color
	if color == "" {
		color = "#6b7280"
	}

	col, err := h.FlowService.CreateColumn(ctx, domain.ProjectColumnInput{
		ProjectID:   req.ProjectID.Int64(),
		WorkspaceID: req.WorkspaceID.Int64(),
		Name:        req.Name,
		Color:       color,
		SortOrder:   req.SortOrder,
		CreatedBy:   authUser.UserID,
	})
	if err != nil {
		return CreateColumnResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create column"}
	}
	publishColumnEvent("column.created", col.ProjectID, col.ID, authUser.UserID)
	return CreateColumnResponse{Column: toColumnPayload(col)}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/projects/columns
// ─────────────────────────────────────────────

func UpdateColumnRoutes() []string { return []string{"/flow/projects/columns"} }

type UpdateColumnHandler struct{ FlowService *application.Service }

type UpdateColumnRequest struct {
	ID        JSONID `json:"id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	SortOrder int    `json:"sort_order"`
}

type UpdateColumnResponse struct {
	Message string `json:"message"`
}

func (h *UpdateColumnHandler) Validate(_ context.Context, req *UpdateColumnRequest) *errorModels.ServerError {
	if req.ID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "name is required"}
	}
	return nil
}

func (h *UpdateColumnHandler) Serve(ctx context.Context, req *UpdateColumnRequest) (UpdateColumnResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateColumnResponse{}, errorModels.PermissionDeniedServerError()
	}

	err := h.FlowService.UpdateColumn(ctx, domain.UpdateProjectColumnInput{
		ID:        req.ID.Int64(),
		ActorID:   authUser.UserID,
		Name:      req.Name,
		Color:     req.Color,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return UpdateColumnResponse{}, errorModels.PermissionDeniedServerError()
		}
		return UpdateColumnResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update column"}
	}
	publishColumnEvent(
		"column.updated",
		h.FlowService.GetColumnProject(ctx, req.ID.Int64()),
		req.ID.Int64(),
		authUser.UserID,
	)
	return UpdateColumnResponse{Message: "Column updated"}, nil
}

// ─────────────────────────────────────────────
// DELETE /flow/projects/columns
// ─────────────────────────────────────────────

func DeleteColumnRoutes() []string { return []string{"/flow/projects/columns"} }

type DeleteColumnHandler struct{ FlowService *application.Service }

type DeleteColumnRequest struct {
	ID JSONID `json:"id"`
}

type DeleteColumnResponse struct {
	Message string `json:"message"`
}

func (h *DeleteColumnHandler) Validate(_ context.Context, req *DeleteColumnRequest) *errorModels.ServerError {
	if req.ID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *DeleteColumnHandler) Serve(ctx context.Context, req *DeleteColumnRequest) (DeleteColumnResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteColumnResponse{}, errorModels.PermissionDeniedServerError()
	}

	// Snapshot the owning project before the row is gone so the realtime
	// event can route to the right topic.
	projectID := h.FlowService.GetColumnProject(ctx, req.ID.Int64())

	if err := h.FlowService.DeleteColumn(ctx, req.ID.Int64(), authUser.UserID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return DeleteColumnResponse{}, errorModels.PermissionDeniedServerError()
		}
		return DeleteColumnResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete column"}
	}
	publishColumnEvent("column.deleted", projectID, req.ID.Int64(), authUser.UserID)
	return DeleteColumnResponse{Message: "Column deleted"}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/projects/columns/move
// ─────────────────────────────────────────────

func MoveTaskToColumnRoutes() []string { return []string{"/flow/projects/columns/move"} }

type MoveTaskToColumnHandler struct{ FlowService *application.Service }

type MoveTaskToColumnRequest struct {
	TaskID   JSONID `json:"task_id"`
	ColumnID JSONID `json:"column_id"`
}

type MoveTaskToColumnResponse struct {
	Message string `json:"message"`
}

func (h *MoveTaskToColumnHandler) Validate(_ context.Context, req *MoveTaskToColumnRequest) *errorModels.ServerError {
	if req.TaskID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "task_id is required"}
	}
	return nil
}

func (h *MoveTaskToColumnHandler) Serve(ctx context.Context, req *MoveTaskToColumnRequest) (MoveTaskToColumnResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return MoveTaskToColumnResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.MoveTaskToColumn(ctx, req.TaskID.Int64(), req.ColumnID.Int64(), authUser.UserID); err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return MoveTaskToColumnResponse{}, errorModels.PermissionDeniedServerError()
		}
		return MoveTaskToColumnResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to move task"}
	}

	// A move is a task update from the board's perspective — consumers
	// invalidate their task list and the card re-appears in the new
	// column. Route to the project the destination column belongs to.
	projectID := h.FlowService.GetColumnProject(ctx, req.ColumnID.Int64())
	publishTaskEventByProject("task.updated", projectID, req.TaskID.Int64(), authUser.UserID)

	return MoveTaskToColumnResponse{Message: "Task moved"}, nil
}
