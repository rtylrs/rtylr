package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
)

// --- Task Activity ---

func GetTaskActivityRoutes() []string     { return []string{"/flow/tasks/{id}/activity"} }
func GetTaskActivityByQPRoutes() []string { return []string{"/flow/tasks/activity"} }

type (
	GetTaskActivityHandler struct{ FlowService *application.Service }
	GetTaskActivityRequest struct{}
)

type GetTaskActivityResponse struct {
	Activity []*flowtaskv1.TaskActivity `json:"activity"`
}

func (h *GetTaskActivityHandler) Serve(ctx context.Context, _ *GetTaskActivityRequest) (GetTaskActivityResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTaskActivityResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return GetTaskActivityResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid task ID"}
	}
	activities, err := h.FlowService.GetTaskActivity(ctx, taskID)
	if err != nil {
		return GetTaskActivityResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get task activity"}
	}
	payloads := make([]*flowtaskv1.TaskActivity, len(activities))
	for i, a := range activities {
		payloads[i] = toActivityPayload(a)
	}
	return GetTaskActivityResponse{Activity: payloads}, nil
}

// GetTaskActivityByQPHandler handles GET /flow/tasks/activity?task_id=.
type GetTaskActivityByQPHandler struct{ FlowService *application.Service }

type GetTaskActivityByQPRequest struct {
	TaskID int64 `json:"task_id"`
}

type GetTaskActivityByQPResponse = GetTaskActivityResponse

func (h *GetTaskActivityByQPHandler) Validate(_ context.Context, req *GetTaskActivityByQPRequest) *errorModels.ServerError {
	if req.TaskID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "task_id is required"}
	}
	return nil
}

func (h *GetTaskActivityByQPHandler) Serve(ctx context.Context, req *GetTaskActivityByQPRequest) (GetTaskActivityByQPResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTaskActivityByQPResponse{}, errorModels.PermissionDeniedServerError()
	}
	activities, err := h.FlowService.GetTaskActivity(ctx, req.TaskID)
	if err != nil {
		return GetTaskActivityByQPResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get task activity"}
	}
	payloads := make([]*flowtaskv1.TaskActivity, len(activities))
	for i, a := range activities {
		payloads[i] = toActivityPayload(a)
	}
	return GetTaskActivityByQPResponse{Activity: payloads}, nil
}

// --- Task Subtasks ---

func GetSubtasksRoutes() []string { return []string{"/flow/tasks/{id}/subtasks"} }

type (
	GetSubtasksHandler  struct{ FlowService *application.Service }
	GetSubtasksRequest  struct{}
	GetSubtasksResponse struct {
		Subtasks []TaskPayloadCompact `json:"subtasks"`
	}
)

type TaskPayloadCompact struct {
	ID        int64   `json:"id"`
	Title     string  `json:"title"`
	Status    string  `json:"status"`
	Priority  string  `json:"priority"`
	DueDate   *string `json:"due_date"`
	CreatedAt string  `json:"created_at"`
}

func (h *GetSubtasksHandler) Serve(ctx context.Context, _ *GetSubtasksRequest) (GetSubtasksResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetSubtasksResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return GetSubtasksResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid task ID"}
	}
	subtasks, err := h.FlowService.GetSubtasks(ctx, taskID)
	if err != nil {
		return GetSubtasksResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get subtasks"}
	}
	payloads := make([]TaskPayloadCompact, len(subtasks))
	for i, t := range subtasks {
		payloads[i] = TaskPayloadCompact{
			ID:        t.ID,
			Title:     t.Title,
			Status:    t.Status,
			Priority:  t.Priority,
			CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if t.DueDate != nil {
			s := t.DueDate.Format("2006-01-02")
			payloads[i].DueDate = &s
		}
	}
	return GetSubtasksResponse{Subtasks: payloads}, nil
}

func CreateSubtaskRoutes() []string { return []string{"/flow/tasks/{id}/subtasks"} }

type CreateSubtaskHandler struct{ FlowService *application.Service }

type CreateSubtaskRequest struct {
	ProjectID      int64    `json:"project_id"`
	WorkspaceID    int64    `json:"workspace_id"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	Priority       string   `json:"priority"`
	AssigneeID     *int64   `json:"assignee_id"`
	DueDate        *string  `json:"due_date"`
	StartDate      *string  `json:"start_date"`
	EstimatedHours *float64 `json:"estimated_hours"`
}

type CreateSubtaskResponse struct {
	TaskID  int64  `json:"task_id"`
	Message string `json:"message"`
}

func (h *CreateSubtaskHandler) Validate(_ context.Context, req *CreateSubtaskRequest) *errorModels.ServerError {
	if req.Title == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Title is required"}
	}
	if req.DueDate != nil {
		if _, err := time.Parse("2006-01-02", *req.DueDate); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "due_date must be in YYYY-MM-DD format"}
		}
	}
	if req.StartDate != nil {
		if _, err := time.Parse("2006-01-02", *req.StartDate); err != nil {
			return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "start_date must be in YYYY-MM-DD format"}
		}
	}
	return nil
}

func (h *CreateSubtaskHandler) Serve(ctx context.Context, req *CreateSubtaskRequest) (CreateSubtaskResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateSubtaskResponse{}, errorModels.PermissionDeniedServerError()
	}
	authUser := device.User
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateSubtaskResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	parentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return CreateSubtaskResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid task ID"}
	}
	input := domain.TaskInput{
		ProjectID:      req.ProjectID,
		WorkspaceID:    req.WorkspaceID,
		OwnerID:        authUser.UserID,
		BusinessID:     businessID,
		ParentTaskID:   &parentID,
		Title:          req.Title,
		Description:    req.Description,
		Priority:       req.Priority,
		AssigneeID:     req.AssigneeID,
		ReporterID:     authUser.UserID,
		EstimatedHours: req.EstimatedHours,
	}
	if req.DueDate != nil {
		if t, parseErr := time.Parse("2006-01-02", *req.DueDate); parseErr == nil {
			input.DueDate = &t
		}
	}
	if req.StartDate != nil {
		if t, parseErr := time.Parse("2006-01-02", *req.StartDate); parseErr == nil {
			input.StartDate = &t
		}
	}
	task, createErr := h.FlowService.CreateSubtask(ctx, input)
	if createErr != nil {
		return CreateSubtaskResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create subtask"}
	}
	return CreateSubtaskResponse{TaskID: task.ID, Message: "Subtask created successfully"}, nil
}

// --- Task Attachments ---

func GetAttachmentsRoutes() []string { return []string{"/flow/tasks/{id}/attachments"} }

type (
	GetAttachmentsHandler struct{ FlowService *application.Service }
	GetAttachmentsRequest struct{}
)

type AttachmentPayload struct {
	ID         int64  `json:"id"`
	TaskID     int64  `json:"task_id"`
	Name       string `json:"name"`
	URL        string `json:"url"`
	Size       int64  `json:"size"`
	UploadedBy int64  `json:"uploaded_by"`
	CreatedAt  string `json:"created_at"`
}

type GetAttachmentsResponse struct {
	Attachments []AttachmentPayload `json:"attachments"`
}

func (h *GetAttachmentsHandler) Serve(ctx context.Context, _ *GetAttachmentsRequest) (GetAttachmentsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetAttachmentsResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return GetAttachmentsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid task ID"}
	}
	attachments, err := h.FlowService.GetAttachments(ctx, taskID)
	if err != nil {
		return GetAttachmentsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get attachments"}
	}
	payloads := make([]AttachmentPayload, len(attachments))
	for i, a := range attachments {
		payloads[i] = AttachmentPayload{
			ID:         a.ID,
			TaskID:     a.TaskID,
			Name:       a.Name,
			URL:        a.URL,
			Size:       a.Size,
			UploadedBy: a.UploadedBy,
			CreatedAt:  a.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return GetAttachmentsResponse{Attachments: payloads}, nil
}

func CreateAttachmentRoutes() []string { return []string{"/flow/tasks/{id}/attachments"} }

type CreateAttachmentHandler struct{ FlowService *application.Service }

type CreateAttachmentRequest struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Size int64  `json:"size"`
}

type CreateAttachmentResponse struct {
	AttachmentID int64  `json:"attachment_id"`
	Message      string `json:"message"`
}

func (h *CreateAttachmentHandler) Validate(_ context.Context, req *CreateAttachmentRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Name is required"}
	}
	if req.URL == "" {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "URL is required"}
	}
	return nil
}

func (h *CreateAttachmentHandler) Serve(ctx context.Context, req *CreateAttachmentRequest) (CreateAttachmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateAttachmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	authUser := device.User
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateAttachmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return CreateAttachmentResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid task ID"}
	}
	input := domain.TaskAttachmentInput{
		TaskID:     taskID,
		OwnerID:    authUser.UserID,
		BusinessID: businessID,
		Name:       req.Name,
		URL:        req.URL,
		Size:       req.Size,
		UploadedBy: authUser.UserID,
	}
	id, createErr := h.FlowService.CreateAttachment(ctx, input)
	if createErr != nil {
		return CreateAttachmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create attachment"}
	}
	return CreateAttachmentResponse{AttachmentID: id, Message: "Attachment created successfully"}, nil
}

func DeleteAttachmentRoutes() []string { return []string{"/flow/tasks/{taskId}/attachments/{id}"} }

type (
	DeleteAttachmentHandler  struct{ FlowService *application.Service }
	DeleteAttachmentRequest  struct{}
	DeleteAttachmentResponse struct {
		Message string `json:"message"`
	}
)

func (h *DeleteAttachmentHandler) Serve(ctx context.Context, _ *DeleteAttachmentRequest) (DeleteAttachmentResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteAttachmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	authUser := device.User
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteAttachmentResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteAttachmentResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid attachment ID"}
	}
	if deleteErr := h.FlowService.DeleteAttachment(ctx, id, businessID); deleteErr != nil {
		return DeleteAttachmentResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete attachment"}
	}
	return DeleteAttachmentResponse{Message: "Attachment deleted successfully"}, nil
}

// --- Time Logs ---

func GetTimeLogsRoutes() []string { return []string{"/flow/tasks/{id}/time"} }

type (
	GetTimeLogsHandler struct{ FlowService *application.Service }
	GetTimeLogsRequest struct{}
)

type TimeLogPayload struct {
	ID              int64  `json:"id"`
	TaskID          int64  `json:"task_id"`
	UserID          int64  `json:"user_id"`
	DurationMinutes int    `json:"duration_minutes"`
	Description     string `json:"description"`
	LoggedAt        string `json:"logged_at"`
	CreatedAt       string `json:"created_at"`
}

type GetTimeLogsResponse struct {
	TimeLogs []TimeLogPayload `json:"time_logs"`
}

func (h *GetTimeLogsHandler) Serve(ctx context.Context, _ *GetTimeLogsRequest) (GetTimeLogsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTimeLogsResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return GetTimeLogsResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid task ID"}
	}
	logs, err := h.FlowService.GetTimeLogs(ctx, taskID)
	if err != nil {
		return GetTimeLogsResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get time logs"}
	}
	payloads := make([]TimeLogPayload, len(logs))
	for i, l := range logs {
		payloads[i] = TimeLogPayload{
			ID:              l.ID,
			TaskID:          l.TaskID,
			UserID:          l.UserID,
			DurationMinutes: l.DurationMinutes,
			Description:     l.Description,
			LoggedAt:        l.LoggedAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt:       l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return GetTimeLogsResponse{TimeLogs: payloads}, nil
}

func CreateTimeLogRoutes() []string { return []string{"/flow/tasks/{id}/time"} }

type CreateTimeLogHandler struct{ FlowService *application.Service }

type CreateTimeLogRequest struct {
	DurationMinutes int    `json:"duration_minutes"`
	Description     string `json:"description"`
}

type CreateTimeLogResponse struct {
	TimeLogID int64  `json:"time_log_id"`
	Message   string `json:"message"`
}

func (h *CreateTimeLogHandler) Validate(_ context.Context, req *CreateTimeLogRequest) *errorModels.ServerError {
	if req.DurationMinutes <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Duration must be greater than 0"}
	}
	return nil
}

func (h *CreateTimeLogHandler) Serve(ctx context.Context, req *CreateTimeLogRequest) (CreateTimeLogResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTimeLogResponse{}, errorModels.PermissionDeniedServerError()
	}
	authUser := device.User
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateTimeLogResponse{}, errorModels.PermissionDeniedServerError()
	}
	idStr := chi.URLParamFromCtx(ctx, "id")
	taskID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return CreateTimeLogResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid task ID"}
	}
	input := domain.TimeLogInput{
		OwnerID:         authUser.UserID,
		BusinessID:      businessID,
		TaskID:          taskID,
		UserID:          authUser.UserID,
		DurationMinutes: req.DurationMinutes,
		Description:     req.Description,
	}
	id, createErr := h.FlowService.CreateTimeLog(ctx, input)
	if createErr != nil {
		return CreateTimeLogResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create time log"}
	}
	return CreateTimeLogResponse{TimeLogID: id, Message: "Time log created successfully"}, nil
}

// --- Delete Time Log ---

func DeleteTimeLogRoutes() []string { return []string{"/flow/tasks/{taskId}/time/{id}"} }

type (
	DeleteTimeLogHandler  struct{ FlowService *application.Service }
	DeleteTimeLogRequest  struct{}
	DeleteTimeLogResponse struct {
		Message string `json:"message"`
	}
)

func (h *DeleteTimeLogHandler) Serve(ctx context.Context, _ *DeleteTimeLogRequest) (DeleteTimeLogResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteTimeLogResponse{}, errorModels.PermissionDeniedServerError()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return DeleteTimeLogResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid time log ID"}
	}

	taskIdStr := chi.URLParamFromCtx(ctx, "taskId")
	taskID, err := strconv.ParseInt(taskIdStr, 10, 64)
	if err != nil {
		return DeleteTimeLogResponse{}, &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Invalid task ID"}
	}

	if deleteErr := h.FlowService.DeleteTimeLog(ctx, id, device.User.UserID, taskID); deleteErr != nil {
		return DeleteTimeLogResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete time log"}
	}

	return DeleteTimeLogResponse{Message: "Time log deleted successfully"}, nil
}

// --- Flat-URL Time Log endpoints (spec-style: query params / body task_id) ---

func GetTimeLogsFlatRoutes() []string   { return []string{"/flow/tasks/time"} }
func CreateTimeLogFlatRoutes() []string { return []string{"/flow/tasks/time"} }
func DeleteTimeLogFlatRoutes() []string { return []string{"/flow/tasks/time"} }

// GET /flow/tasks/time?task_id=

type GetTimeLogsFlatHandler struct{ FlowService *application.Service }

type GetTimeLogsFlatRequest struct {
	TaskID int64 `json:"task_id"`
}

type GetTimeLogsFlatResponse = GetTimeLogsResponse

func (h *GetTimeLogsFlatHandler) Validate(_ context.Context, req *GetTimeLogsFlatRequest) *errorModels.ServerError {
	if req.TaskID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "task_id is required"}
	}
	return nil
}

func (h *GetTimeLogsFlatHandler) Serve(ctx context.Context, req *GetTimeLogsFlatRequest) (GetTimeLogsFlatResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTimeLogsFlatResponse{}, errorModels.PermissionDeniedServerError()
	}
	logs, err := h.FlowService.GetTimeLogs(ctx, req.TaskID)
	if err != nil {
		return GetTimeLogsFlatResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to get time logs"}
	}
	payloads := make([]TimeLogPayload, len(logs))
	for i, l := range logs {
		payloads[i] = TimeLogPayload{
			ID:              l.ID,
			TaskID:          l.TaskID,
			UserID:          l.UserID,
			DurationMinutes: l.DurationMinutes,
			Description:     l.Description,
			LoggedAt:        l.LoggedAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt:       l.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return GetTimeLogsFlatResponse{TimeLogs: payloads}, nil
}

// POST /flow/tasks/time

type CreateTimeLogFlatHandler struct{ FlowService *application.Service }

type CreateTimeLogFlatRequest struct {
	TaskID          int64      `json:"task_id"`
	DurationMinutes int        `json:"duration_minutes"`
	Description     string     `json:"description"`
	LoggedAt        *time.Time `json:"logged_at"`
}

type CreateTimeLogFlatResponse = CreateTimeLogResponse

func (h *CreateTimeLogFlatHandler) Validate(_ context.Context, req *CreateTimeLogFlatRequest) *errorModels.ServerError {
	if req.TaskID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "task_id is required"}
	}
	if req.DurationMinutes <= 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "duration_minutes must be greater than 0"}
	}
	return nil
}

func (h *CreateTimeLogFlatHandler) Serve(ctx context.Context, req *CreateTimeLogFlatRequest) (CreateTimeLogFlatResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return CreateTimeLogFlatResponse{}, errorModels.PermissionDeniedServerError()
	}
	authUser := device.User
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateTimeLogFlatResponse{}, errorModels.PermissionDeniedServerError()
	}
	input := domain.TimeLogInput{
		OwnerID:         authUser.UserID,
		BusinessID:      businessID,
		TaskID:          req.TaskID,
		UserID:          authUser.UserID,
		DurationMinutes: req.DurationMinutes,
		Description:     req.Description,
	}
	id, createErr := h.FlowService.CreateTimeLog(ctx, input)
	if createErr != nil {
		return CreateTimeLogFlatResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to create time log"}
	}
	return CreateTimeLogFlatResponse{TimeLogID: id, Message: "Time log created successfully"}, nil
}

// DELETE /flow/tasks/time?id=

type DeleteTimeLogFlatHandler struct{ FlowService *application.Service }

type DeleteTimeLogFlatRequest struct {
	ID int64 `json:"id"`
}

type DeleteTimeLogFlatResponse = DeleteTimeLogResponse

func (h *DeleteTimeLogFlatHandler) Validate(_ context.Context, req *DeleteTimeLogFlatRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "id is required"}
	}
	return nil
}

func (h *DeleteTimeLogFlatHandler) Serve(ctx context.Context, req *DeleteTimeLogFlatRequest) (DeleteTimeLogFlatResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return DeleteTimeLogFlatResponse{}, errorModels.PermissionDeniedServerError()
	}
	if deleteErr := h.FlowService.DeleteTimeLogFlat(ctx, req.ID, device.User.UserID); deleteErr != nil {
		return DeleteTimeLogFlatResponse{}, &errorModels.ServerError{Code: http.StatusInternalServerError, Message: "Failed to delete time log"}
	}
	return DeleteTimeLogFlatResponse{Message: "Time log deleted successfully"}, nil
}
