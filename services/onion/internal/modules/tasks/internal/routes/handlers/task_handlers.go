package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"

	"github.com/voxire/onion/internal/modules/tasks/internal/application"
	"github.com/voxire/onion/internal/modules/tasks/internal/domain"
)

type TaskPayload struct {
	ID          int64           `json:"id"`
	OwnerID     int64           `json:"owner_id"`
	ProjectID   *int64          `json:"project_id"`
	WorkspaceID *int64          `json:"workspace_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	Priority    string          `json:"priority"`
	AssigneeID  *int64          `json:"assignee_id"`
	LabelIDs    json.RawMessage `json:"label_ids"`
	DueDate     *string         `json:"due_date"`
	CreatedAt   string          `json:"created_at"`
	UpdatedAt   string          `json:"updated_at"`
}

func newTaskPayload(t domain.Task) TaskPayload {
	labelIDs := json.RawMessage(t.LabelIDs)
	if labelIDs == nil {
		labelIDs = json.RawMessage("[]")
	}
	p := TaskPayload{
		ID:          t.ID,
		OwnerID:     t.OwnerID,
		ProjectID:   t.ProjectID,
		WorkspaceID: t.WorkspaceID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		AssigneeID:  t.AssigneeID,
		LabelIDs:    labelIDs,
		CreatedAt:   t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if t.DueDate != nil {
		s := t.DueDate.Format("2006-01-02")
		p.DueDate = &s
	}
	return p
}

// ─── List Tasks ───────────────────────────────────────────────────────────────

func TaskListRoutes() []string { return []string{"/flow/tasks"} }

type ListTasksHandler struct{ Service *application.Service }
type ListTasksRequest struct{}
type ListTasksResponse struct {
	Tasks []TaskPayload `json:"tasks"`
}

func (h *ListTasksHandler) Serve(ctx context.Context, _ *ListTasksRequest) (ListTasksResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListTasksResponse{}, errorModels.PermissionDeniedServerError()
	}
	tasks, err := h.Service.List(ctx, device.User.UserID)
	if err != nil {
		return ListTasksResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list tasks"}
	}
	payloads := make([]TaskPayload, len(tasks))
	for i, t := range tasks {
		payloads[i] = newTaskPayload(t)
	}
	return ListTasksResponse{Tasks: payloads}, nil
}

// ─── Create Task ──────────────────────────────────────────────────────────────

func TaskCreateRoutes() []string { return []string{"/flow/tasks"} }

type CreateTaskHandler struct{ Service *application.Service }

type CreateTaskRequest struct {
	ProjectID   *int64          `json:"project_id"`
	WorkspaceID *int64          `json:"workspace_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Priority    string          `json:"priority"`
	AssigneeID  *int64          `json:"assignee_id"`
	LabelIDs    json.RawMessage `json:"label_ids"`
	DueDate     *string         `json:"due_date"`
}

type CreateTaskResponse struct {
	TaskID  int64  `json:"task_id"`
	Message string `json:"message"`
}

func (h *CreateTaskHandler) Validate(_ context.Context, req *CreateTaskRequest) *errorModels.ServerError {
	if req.Title == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Title is required"}
	}
	return nil
}

func (h *CreateTaskHandler) Serve(ctx context.Context, req *CreateTaskRequest) (CreateTaskResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTaskResponse{}, errorModels.PermissionDeniedServerError()
	}
	labelIDs := []byte(req.LabelIDs)
	if labelIDs == nil {
		labelIDs = []byte("[]")
	}
	input := domain.TaskInput{
		OwnerID:     device.User.UserID,
		ProjectID:   req.ProjectID,
		WorkspaceID: req.WorkspaceID,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
		AssigneeID:  req.AssigneeID,
		LabelIDs:    labelIDs,
	}
	if req.DueDate != nil {
		if t, err := time.Parse("2006-01-02", *req.DueDate); err == nil {
			input.DueDate = &t
		}
	}
	id, err := h.Service.Create(ctx, input)
	if err != nil {
		return CreateTaskResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create task"}
	}
	return CreateTaskResponse{TaskID: id, Message: "Task created successfully"}, nil
}

// ─── Update Task ──────────────────────────────────────────────────────────────

func TaskUpdateRoutes() []string { return []string{"/flow/tasks/{id}"} }

type UpdateTaskHandler struct{ Service *application.Service }

type UpdateTaskRequest struct {
	ProjectID   *int64          `json:"project_id"`
	WorkspaceID *int64          `json:"workspace_id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Status      string          `json:"status"`
	Priority    string          `json:"priority"`
	AssigneeID  *int64          `json:"assignee_id"`
	LabelIDs    json.RawMessage `json:"label_ids"`
	DueDate     *string         `json:"due_date"`
}

type UpdateTaskResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTaskHandler) Validate(_ context.Context, req *UpdateTaskRequest) *errorModels.ServerError {
	if req.Title == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Title is required"}
	}
	return nil
}

func (h *UpdateTaskHandler) Serve(ctx context.Context, req *UpdateTaskRequest) (UpdateTaskResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return UpdateTaskResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return UpdateTaskResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid task ID"}
	}
	labelIDs := []byte(req.LabelIDs)
	if labelIDs == nil {
		labelIDs = []byte("[]")
	}
	input := domain.TaskUpdateInput{
		ID:          id,
		OwnerID:     device.User.UserID,
		ProjectID:   req.ProjectID,
		WorkspaceID: req.WorkspaceID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		AssigneeID:  req.AssigneeID,
		LabelIDs:    labelIDs,
	}
	if req.DueDate != nil {
		if t, parseErr := time.Parse("2006-01-02", *req.DueDate); parseErr == nil {
			input.DueDate = &t
		}
	}
	if updateErr := h.Service.Update(ctx, input); updateErr != nil {
		if errors.Is(updateErr, domain.ErrTaskNotFound) {
			return UpdateTaskResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Task not found"}
		}
		return UpdateTaskResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to update task"}
	}
	return UpdateTaskResponse{Message: "Task updated successfully"}, nil
}

// ─── Delete Task ──────────────────────────────────────────────────────────────

func TaskDeleteRoutes() []string { return []string{"/flow/tasks/{id}"} }

type DeleteTaskHandler struct{ Service *application.Service }
type DeleteTaskRequest struct{}
type DeleteTaskResponse struct {
	Message string `json:"message"`
}

func (h *DeleteTaskHandler) Serve(ctx context.Context, _ *DeleteTaskRequest) (DeleteTaskResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteTaskResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteTaskResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid task ID"}
	}
	if deleteErr := h.Service.Delete(ctx, device.User.UserID, id); deleteErr != nil {
		if errors.Is(deleteErr, domain.ErrTaskNotFound) {
			return DeleteTaskResponse{}, &errorModels.ServerError{Code: http.StatusNotFound, Message: "Task not found"}
		}
		return DeleteTaskResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete task"}
	}
	return DeleteTaskResponse{Message: "Task deleted successfully"}, nil
}

// ─── List Labels ──────────────────────────────────────────────────────────────

func LabelListRoutes() []string { return []string{"/flow/labels"} }

type ListLabelsHandler struct{ Service *application.Service }
type ListLabelsRequest struct{}

type LabelPayload struct {
	ID        int64  `json:"id"`
	OwnerID   int64  `json:"owner_id"`
	Name      string `json:"name"`
	Color     string `json:"color"`
	CreatedAt string `json:"created_at"`
}

type ListLabelsResponse struct {
	Labels []LabelPayload `json:"labels"`
}

func (h *ListLabelsHandler) Serve(ctx context.Context, _ *ListLabelsRequest) (ListLabelsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return ListLabelsResponse{}, errorModels.PermissionDeniedServerError()
	}
	labels, err := h.Service.ListLabels(ctx, device.User.UserID)
	if err != nil {
		return ListLabelsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to list labels"}
	}
	payloads := make([]LabelPayload, len(labels))
	for i, l := range labels {
		payloads[i] = LabelPayload{ID: l.ID, OwnerID: l.OwnerID, Name: l.Name, Color: l.Color, CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05Z")}
	}
	return ListLabelsResponse{Labels: payloads}, nil
}

// ─── Create Label ─────────────────────────────────────────────────────────────

func LabelCreateRoutes() []string { return []string{"/flow/labels"} }

type CreateLabelHandler struct{ Service *application.Service }

type CreateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

type CreateLabelResponse struct {
	LabelID int64  `json:"label_id"`
	Message string `json:"message"`
}

func (h *CreateLabelHandler) Validate(_ context.Context, req *CreateLabelRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	return nil
}

func (h *CreateLabelHandler) Serve(ctx context.Context, req *CreateLabelRequest) (CreateLabelResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateLabelResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.CreateLabel(ctx, domain.LabelInput{
		OwnerID: device.User.UserID,
		Name:    req.Name,
		Color:   req.Color,
	})
	if err != nil {
		return CreateLabelResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create label"}
	}
	return CreateLabelResponse{LabelID: id, Message: "Label created successfully"}, nil
}
