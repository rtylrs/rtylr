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

func ListTasksRoutes() []string       { return []string{"/flow/tasks"} }
func CreateTaskRoutes() []string      { return []string{"/flow/tasks"} }
func GetTaskDetailRoutes() []string   { return []string{"/flow/tasks/detail"} }
func UpdateTaskRoutes() []string      { return []string{"/flow/tasks"} }
func DeleteTaskRoutes() []string      { return []string{"/flow/tasks"} }
func UpdateTaskStatusRoutes() []string { return []string{"/flow/tasks/status"} }
func GetMyTasksRoutes() []string      { return []string{"/flow/tasks/my"} }
func AddCommentRoutes() []string      { return []string{"/flow/tasks/comment"} }
func DeleteCommentRoutes() []string   { return []string{"/flow/tasks/comment"} }

// ─────────────────────────────────────────────
// Shared payload types
// ─────────────────────────────────────────────

type TaskPayload struct {
	ID             int64      `json:"id"`
	ProjectID      int64      `json:"project_id"`
	WorkspaceID    int64      `json:"workspace_id"`
	OwnerID        int64      `json:"owner_id"`
	ParentTaskID   *int64     `json:"parent_task_id,omitempty"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"`
	Priority       string     `json:"priority"`
	AssigneeID     *int64     `json:"assignee_id,omitempty"`
	ReporterID     int64      `json:"reporter_id"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	StartDate      *time.Time `json:"start_date,omitempty"`
	CompletedAt    *time.Time `json:"completed_at,omitempty"`
	EstimatedHours *float64   `json:"estimated_hours,omitempty"`
	ActualHours    *float64   `json:"actual_hours,omitempty"`
	SortOrder      int        `json:"sort_order"`
	Labels         []int64    `json:"labels"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type TaskCommentPayload struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	Content   string    `json:"content"`
	IsEdited  bool      `json:"is_edited"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TaskActivityPayload struct {
	ID        int64     `json:"id"`
	TaskID    int64     `json:"task_id"`
	UserID    int64     `json:"user_id"`
	Action    string    `json:"action"`
	OldValue  string    `json:"old_value"`
	NewValue  string    `json:"new_value"`
	CreatedAt time.Time `json:"created_at"`
}

func toTaskPayload(t domain.Task) TaskPayload {
	labels := t.Labels
	if labels == nil {
		labels = []int64{}
	}

	return TaskPayload{
		ID:             t.ID,
		ProjectID:      t.ProjectID,
		WorkspaceID:    t.WorkspaceID,
		OwnerID:        t.OwnerID,
		ParentTaskID:   t.ParentTaskID,
		Title:          t.Title,
		Description:    t.Description,
		Status:         t.Status,
		Priority:       t.Priority,
		AssigneeID:     t.AssigneeID,
		ReporterID:     t.ReporterID,
		DueDate:        t.DueDate,
		StartDate:      t.StartDate,
		CompletedAt:    t.CompletedAt,
		EstimatedHours: t.EstimatedHours,
		ActualHours:    t.ActualHours,
		SortOrder:      t.SortOrder,
		Labels:         labels,
		CreatedAt:      t.CreatedAt,
		UpdatedAt:      t.UpdatedAt,
	}
}

func toCommentPayload(c domain.TaskComment) TaskCommentPayload {
	return TaskCommentPayload{
		ID:        c.ID,
		TaskID:    c.TaskID,
		UserID:    c.UserID,
		Content:   c.Content,
		IsEdited:  c.IsEdited,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func toActivityPayload(a domain.TaskActivity) TaskActivityPayload {
	return TaskActivityPayload{
		ID:        a.ID,
		TaskID:    a.TaskID,
		UserID:    a.UserID,
		Action:    a.Action,
		OldValue:  a.OldValue,
		NewValue:  a.NewValue,
		CreatedAt: a.CreatedAt,
	}
}

// ─────────────────────────────────────────────
// GET /flow/tasks
// ─────────────────────────────────────────────

type ListTasksHandler struct {
	FlowService *application.Service
}

type ListTasksRequest struct {
	ProjectID  int64  `json:"project_id"`
	Status     string `json:"status"`
	AssigneeID *int64 `json:"assignee_id"`
	Priority   string `json:"priority"`
}

type ListTasksResponse struct {
	Tasks []TaskPayload `json:"tasks"`
}

func (h *ListTasksHandler) Validate(_ context.Context, req *ListTasksRequest) *errorModels.ServerError {
	if req.ProjectID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "project_id is required"}
	}

	return nil
}

func (h *ListTasksHandler) Serve(ctx context.Context, req *ListTasksRequest) (ListTasksResponse, *errorModels.ServerError) {
	_, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListTasksResponse{}, errorModels.PermissionDeniedServerError()
	}

	filter := domain.ListTasksFilter{
		ProjectID:  req.ProjectID,
		Status:     req.Status,
		AssigneeID: req.AssigneeID,
		Priority:   req.Priority,
	}

	tasks, err := h.FlowService.ListTasks(ctx, filter)
	if err != nil {
		return ListTasksResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list tasks",
		}
	}

	payloads := make([]TaskPayload, 0, len(tasks))
	for _, t := range tasks {
		payloads = append(payloads, toTaskPayload(t))
	}

	return ListTasksResponse{Tasks: payloads}, nil
}

// ─────────────────────────────────────────────
// POST /flow/tasks
// ─────────────────────────────────────────────

type CreateTaskHandler struct {
	FlowService *application.Service
}

type CreateTaskRequest struct {
	ProjectID      int64      `json:"project_id"`
	WorkspaceID    int64      `json:"workspace_id"`
	ParentTaskID   *int64     `json:"parent_task_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Priority       string     `json:"priority"`
	AssigneeID     *int64     `json:"assignee_id"`
	DueDate        *time.Time `json:"due_date"`
	StartDate      *time.Time `json:"start_date"`
	EstimatedHours *float64   `json:"estimated_hours"`
	Labels         []int64    `json:"labels"`
}

type CreateTaskResponse struct {
	Task TaskPayload `json:"task"`
}

func (h *CreateTaskHandler) Validate(_ context.Context, req *CreateTaskRequest) *errorModels.ServerError {
	if req.ProjectID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Project ID is required"}
	}

	if req.WorkspaceID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Workspace ID is required"}
	}

	if req.Title == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Title is required"}
	}

	return nil
}

func (h *CreateTaskHandler) Serve(ctx context.Context, req *CreateTaskRequest) (CreateTaskResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateTaskResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.TaskInput{
		ProjectID:      req.ProjectID,
		WorkspaceID:    req.WorkspaceID,
		OwnerID:        authUser.UserID,
		ParentTaskID:   req.ParentTaskID,
		Title:          req.Title,
		Description:    req.Description,
		Priority:       req.Priority,
		AssigneeID:     req.AssigneeID,
		ReporterID:     authUser.UserID,
		DueDate:        req.DueDate,
		StartDate:      req.StartDate,
		EstimatedHours: req.EstimatedHours,
		Labels:         req.Labels,
	}

	task, err := h.FlowService.CreateTask(ctx, input)
	if err != nil {
		return CreateTaskResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create task",
		}
	}

	return CreateTaskResponse{Task: toTaskPayload(task)}, nil
}

// ─────────────────────────────────────────────
// GET /flow/tasks/detail
// ─────────────────────────────────────────────

type GetTaskDetailHandler struct {
	FlowService *application.Service
}

type GetTaskDetailRequest struct {
	ID int64 `json:"id"`
}

type GetTaskDetailResponse struct {
	Task     TaskPayload          `json:"task"`
	Subtasks []TaskPayload        `json:"subtasks"`
	Comments []TaskCommentPayload `json:"comments"`
	Activity []TaskActivityPayload `json:"activity"`
}

func (h *GetTaskDetailHandler) Validate(_ context.Context, req *GetTaskDetailRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Task ID is required"}
	}

	return nil
}

func (h *GetTaskDetailHandler) Serve(ctx context.Context, req *GetTaskDetailRequest) (GetTaskDetailResponse, *errorModels.ServerError) {
	_, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetTaskDetailResponse{}, errorModels.PermissionDeniedServerError()
	}

	detail, err := h.FlowService.GetTaskDetail(ctx, req.ID)
	if err != nil {
		return GetTaskDetailResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get task detail",
		}
	}

	subtasks := make([]TaskPayload, 0, len(detail.Subtasks))
	for _, t := range detail.Subtasks {
		subtasks = append(subtasks, toTaskPayload(t))
	}

	comments := make([]TaskCommentPayload, 0, len(detail.Comments))
	for _, c := range detail.Comments {
		comments = append(comments, toCommentPayload(c))
	}

	activity := make([]TaskActivityPayload, 0, len(detail.Activity))
	for _, a := range detail.Activity {
		activity = append(activity, toActivityPayload(a))
	}

	return GetTaskDetailResponse{
		Task:     toTaskPayload(detail.Task),
		Subtasks: subtasks,
		Comments: comments,
		Activity: activity,
	}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/tasks
// ─────────────────────────────────────────────

type UpdateTaskHandler struct {
	FlowService *application.Service
}

type UpdateTaskRequest struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssigneeID  *int64     `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
	StartDate   *time.Time `json:"start_date"`
	Labels      []int64    `json:"labels"`
}

type UpdateTaskResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTaskHandler) Validate(_ context.Context, req *UpdateTaskRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Task ID is required"}
	}

	if req.Title == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Title is required"}
	}

	return nil
}

func (h *UpdateTaskHandler) Serve(ctx context.Context, req *UpdateTaskRequest) (UpdateTaskResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateTaskResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.UpdateTaskInput{
		ID:          req.ID,
		OwnerID:     authUser.UserID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		AssigneeID:  req.AssigneeID,
		DueDate:     req.DueDate,
		StartDate:   req.StartDate,
		Labels:      req.Labels,
	}

	if err := h.FlowService.UpdateTask(ctx, input, authUser.UserID); err != nil {
		return UpdateTaskResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update task",
		}
	}

	return UpdateTaskResponse{Message: "Task updated successfully"}, nil
}

// ─────────────────────────────────────────────
// DELETE /flow/tasks
// ─────────────────────────────────────────────

type DeleteTaskHandler struct {
	FlowService *application.Service
}

type DeleteTaskRequest struct {
	ID int64 `json:"id"`
}

type DeleteTaskResponse struct {
	Message string `json:"message"`
}

func (h *DeleteTaskHandler) Validate(_ context.Context, req *DeleteTaskRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Task ID is required"}
	}

	return nil
}

func (h *DeleteTaskHandler) Serve(ctx context.Context, req *DeleteTaskRequest) (DeleteTaskResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteTaskResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.DeleteTask(ctx, req.ID, authUser.UserID); err != nil {
		return DeleteTaskResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete task",
		}
	}

	return DeleteTaskResponse{Message: "Task deleted successfully"}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/tasks/status
// ─────────────────────────────────────────────

type UpdateTaskStatusHandler struct {
	FlowService *application.Service
}

type UpdateTaskStatusRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

type UpdateTaskStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTaskStatusHandler) Validate(_ context.Context, req *UpdateTaskStatusRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Task ID is required"}
	}

	if req.Status == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Status is required"}
	}

	return nil
}

func (h *UpdateTaskStatusHandler) Serve(ctx context.Context, req *UpdateTaskStatusRequest) (UpdateTaskStatusResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateTaskStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.UpdateTaskStatusInput{
		ID:      req.ID,
		OwnerID: authUser.UserID,
		Status:  req.Status,
	}

	if err := h.FlowService.UpdateTaskStatus(ctx, input, authUser.UserID); err != nil {
		return UpdateTaskStatusResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update task status",
		}
	}

	return UpdateTaskStatusResponse{Message: "Task status updated"}, nil
}

// ─────────────────────────────────────────────
// GET /flow/tasks/my
// ─────────────────────────────────────────────

type GetMyTasksHandler struct {
	FlowService *application.Service
}

type GetMyTasksRequest struct{}

type GetMyTasksResponse struct {
	Tasks []TaskPayload `json:"tasks"`
}

func (h *GetMyTasksHandler) Serve(ctx context.Context, _ *GetMyTasksRequest) (GetMyTasksResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetMyTasksResponse{}, errorModels.PermissionDeniedServerError()
	}

	tasks, err := h.FlowService.GetMyTasks(ctx, authUser.UserID)
	if err != nil {
		return GetMyTasksResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get my tasks",
		}
	}

	payloads := make([]TaskPayload, 0, len(tasks))
	for _, t := range tasks {
		payloads = append(payloads, toTaskPayload(t))
	}

	return GetMyTasksResponse{Tasks: payloads}, nil
}

// ─────────────────────────────────────────────
// POST /flow/tasks/comment
// ─────────────────────────────────────────────

type AddCommentHandler struct {
	FlowService *application.Service
}

type AddCommentRequest struct {
	TaskID  int64  `json:"task_id"`
	Content string `json:"content"`
}

type AddCommentResponse struct {
	Comment TaskCommentPayload `json:"comment"`
}

func (h *AddCommentHandler) Validate(_ context.Context, req *AddCommentRequest) *errorModels.ServerError {
	if req.TaskID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Task ID is required"}
	}

	if req.Content == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Content is required"}
	}

	return nil
}

func (h *AddCommentHandler) Serve(ctx context.Context, req *AddCommentRequest) (AddCommentResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return AddCommentResponse{}, errorModels.PermissionDeniedServerError()
	}

	comment, err := h.FlowService.AddComment(ctx, req.TaskID, authUser.UserID, req.Content)
	if err != nil {
		return AddCommentResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to add comment",
		}
	}

	return AddCommentResponse{Comment: toCommentPayload(comment)}, nil
}

// ─────────────────────────────────────────────
// DELETE /flow/tasks/comment
// ─────────────────────────────────────────────

type DeleteCommentHandler struct {
	FlowService *application.Service
}

type DeleteCommentRequest struct {
	ID int64 `json:"id"`
}

type DeleteCommentResponse struct {
	Message string `json:"message"`
}

func (h *DeleteCommentHandler) Validate(_ context.Context, req *DeleteCommentRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Comment ID is required"}
	}

	return nil
}

func (h *DeleteCommentHandler) Serve(ctx context.Context, req *DeleteCommentRequest) (DeleteCommentResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteCommentResponse{}, errorModels.PermissionDeniedServerError()
	}

	if err := h.FlowService.DeleteComment(ctx, req.ID, authUser.UserID); err != nil {
		return DeleteCommentResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete comment",
		}
	}

	return DeleteCommentResponse{Message: "Comment deleted successfully"}, nil
}
