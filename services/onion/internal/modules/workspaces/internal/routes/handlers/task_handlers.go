package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/application"
	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	"github.com/voxire/onion/internal/platform/wsbus"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
)

// publishTaskEvent fans out a task-level invalidation hint to every
// client subscribed to the owning project. The payload is intentionally
// minimal — the client refetches the canonical entity via REST, keeping
// the wire format for realtime and HTTP in sync.
func publishTaskEvent(eventType string, t *flowtaskv1.Task, actorID int64) {
	if t == nil {
		return
	}
	publishTaskEventByProject(eventType, t.ProjectId, t.Id, actorID)
}

// publishCommentEvent fans a comment-level change out to the task
// channel only. Board/list/timeline projections don't render comments,
// so there's no value in fanning to the project topic too.
func publishCommentEvent(eventType string, taskID, actorID int64) {
	if taskID == 0 {
		return
	}
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("task:%d", taskID),
		Type:     eventType,
		EntityID: taskID,
		ActorID:  actorID,
	})
}

// publishTaskEventByProject is for the write paths that don't return the
// full task (delete, status, move) — caller supplies project + task id.
// We look up the project id first in the handler so this is a pure
// notification call with no DB access.
//
// Events fan out to two topics: the owning project (so board/list/timeline
// subscribers refresh) and the task itself (so an open TaskDetail screen
// refreshes even if it isn't subscribed to the whole project firehose).
func publishTaskEventByProject(eventType string, projectID, taskID, actorID int64) {
	if taskID != 0 {
		wsbus.Publish(wsbus.Event{
			Topic:    fmt.Sprintf("task:%d", taskID),
			Type:     eventType,
			EntityID: taskID,
			ActorID:  actorID,
		})
	}
	if projectID == 0 {
		return
	}
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("project:%d", projectID),
		Type:     eventType,
		EntityID: taskID,
		ActorID:  actorID,
	})
}

// ─────────────────────────────────────────────
// Route declarations
// ─────────────────────────────────────────────

func ListTasksRoutes() []string        { return []string{"/flow/tasks"} }
func CreateTaskRoutes() []string       { return []string{"/flow/tasks"} }
func GetTaskDetailRoutes() []string    { return []string{"/flow/tasks/detail"} }
func UpdateTaskRoutes() []string       { return []string{"/flow/tasks"} }
func DeleteTaskRoutes() []string       { return []string{"/flow/tasks"} }
func UpdateTaskStatusRoutes() []string { return []string{"/flow/tasks/status"} }
func GetMyTasksRoutes() []string       { return []string{"/flow/tasks/my"} }
func AddCommentRoutes() []string       { return []string{"/flow/tasks/comment"} }
func DeleteCommentRoutes() []string    { return []string{"/flow/tasks/comment"} }
func SearchTasksRoutes() []string      { return []string{"/flow/tasks/search"} }

// ─────────────────────────────────────────────
// Shared payload types
// ─────────────────────────────────────────────

func toTaskPayload(t domain.Task) *flowtaskv1.Task {
	labels := t.Labels
	if labels == nil {
		labels = []int64{}
	}

	task := &flowtaskv1.Task{
		Id:             t.ID,
		ProjectId:      t.ProjectID,
		WorkspaceId:    t.WorkspaceID,
		OwnerId:        t.OwnerID,
		ParentTaskId:   t.ParentTaskID,
		Title:          t.Title,
		Description:    t.Description,
		Status:         t.Status,
		Priority:       t.Priority,
		AssigneeId:     t.AssigneeID,
		ReporterId:     t.ReporterID,
		EstimatedHours: t.EstimatedHours,
		ActualHours:    t.ActualHours,
		SortOrder:      int64(t.SortOrder),
		Labels:         labels,
		CreatedAt:      t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:      t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		// Display projections — populated by enriched repository queries.
		Identifier:       t.Identifier,
		ProjectName:      t.ProjectName,
		ProjectColor:     t.ProjectColor,
		AssigneeName:     t.AssigneeName,
		AssigneeInitials: t.AssigneeInitials,
		SubtaskCount:     t.SubtaskCount,
		SectionId:        t.SectionID,
		CycleId:          t.CycleID,
		MilestoneId:      t.MilestoneID,
	}
	if t.DueDate != nil {
		s := t.DueDate.Format("2006-01-02T15:04:05Z")
		task.DueDate = &s
	}
	if t.StartDate != nil {
		s := t.StartDate.Format("2006-01-02T15:04:05Z")
		task.StartDate = &s
	}
	if t.CompletedAt != nil {
		s := t.CompletedAt.Format("2006-01-02T15:04:05Z")
		task.CompletedAt = &s
	}
	return task
}

func toCommentPayload(c domain.TaskComment) *flowtaskv1.TaskComment {
	return &flowtaskv1.TaskComment{
		Id:        c.ID,
		TaskId:    c.TaskID,
		UserId:    c.UserID,
		Content:   c.Content,
		IsEdited:  c.IsEdited,
		CreatedAt: c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toActivityPayload(a domain.TaskActivity) *flowtaskv1.TaskActivity {
	return &flowtaskv1.TaskActivity{
		Id:        a.ID,
		TaskId:    a.TaskID,
		UserId:    a.UserID,
		Action:    a.Action,
		OldValue:  a.OldValue,
		NewValue:  a.NewValue,
		CreatedAt: a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// ─────────────────────────────────────────────
// GET /flow/tasks
// ─────────────────────────────────────────────

type ListTasksHandler struct {
	FlowService *application.Service
}

type ListTasksRequest struct {
	ProjectID   int64            `json:"project_id"`
	WorkspaceID WorkspaceIDField `json:"workspace_id"`
	Status      string           `json:"status"`
	AssigneeID  *int64           `json:"assignee_id"`
	Priority    string           `json:"priority"`
	ActiveOnly  bool             `json:"active_only"`
	Compact     bool             `json:"compact"`
}

type ListTasksResponse struct {
	Tasks []*flowtaskv1.Task `json:"tasks"`
}

func (h *ListTasksHandler) Validate(_ context.Context, req *ListTasksRequest) *errorModels.ServerError {
	if req.ProjectID == 0 && req.WorkspaceID.IsEmpty() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "project_id or workspace_id is required"}
	}

	return nil
}

func (h *ListTasksHandler) Serve(ctx context.Context, req *ListTasksRequest) (ListTasksResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListTasksResponse{}, errorModels.PermissionDeniedServerError()
	}

	var workspaceID int64
	if !req.WorkspaceID.IsEmpty() {
		resolved, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, req.WorkspaceID.String())
		if err != nil {
			return ListTasksResponse{}, &errorModels.ServerError{
				Code:    http.StatusBadRequest,
				Message: "Invalid workspace_id",
			}
		}
		workspaceID = resolved
	}

	if req.ProjectID != 0 && !h.FlowService.IsProjectMember(ctx, authUser.UserID, req.ProjectID) {
		return ListTasksResponse{}, errorModels.PermissionDeniedServerError()
	}

	filter := domain.ListTasksFilter{
		ProjectID:   req.ProjectID,
		WorkspaceID: workspaceID,
		Status:      req.Status,
		AssigneeID:  req.AssigneeID,
		Priority:    req.Priority,
		ActiveOnly:  req.ActiveOnly,
		Compact:     req.Compact,
	}

	tasks, err := h.FlowService.ListTasks(ctx, filter)
	if err != nil {
		return ListTasksResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to list tasks",
		}
	}

	payloads := make([]*flowtaskv1.Task, 0, len(tasks))
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
	ProjectID      JSONID           `json:"project_id"`
	WorkspaceID    WorkspaceIDField `json:"workspace_id"`
	ParentTaskID   *int64           `json:"parent_task_id"`
	Title          string           `json:"title"`
	Description    string           `json:"description"`
	Priority       string           `json:"priority"`
	AssigneeID     *JSONID          `json:"assignee_id"`
	DueDate        *time.Time       `json:"due_date"`
	StartDate      *time.Time       `json:"start_date"`
	EstimatedHours *float64         `json:"estimated_hours"`
	Labels         []int64          `json:"labels"`
}

type CreateTaskResponse struct {
	Task *flowtaskv1.Task `json:"task"`
}

func (h *CreateTaskHandler) Validate(_ context.Context, req *CreateTaskRequest) *errorModels.ServerError {
	// project_id is now optional — if omitted we resolve the caller's
	// auto-provisioned "Inbox" project (see Serve). workspace_id can be
	// empty / "personal" / numeric; Serve handles all three.
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
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateTaskResponse{}, errorModels.PermissionDeniedServerError()
	}

	// Resolve workspace: missing or "personal" → the caller's personal
	// workspace (auto-created). Numeric → use as-is.
	workspaceRaw := req.WorkspaceID.String()
	if workspaceRaw == "" {
		workspaceRaw = "personal"
	}
	workspaceID, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, workspaceRaw)
	if err != nil || workspaceID == 0 {
		return CreateTaskResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid workspace_id",
		}
	}

	// Resolve project: missing → the Inbox project in the resolved
	// workspace (auto-created on first use). This powers "quick create"
	// flows from My Tasks / the command bar.
	projectID := req.ProjectID.Int64()
	if projectID == 0 {
		myTasksID, _, ensureErr := h.FlowService.EnsureMyTasksProject(ctx, authUser.UserID, workspaceID)
		if ensureErr != nil || myTasksID == 0 {
			return CreateTaskResponse{}, &errorModels.ServerError{
				Code:    http.StatusInternalServerError,
				Message: "Failed to resolve My Tasks project",
			}
		}
		projectID = myTasksID
	}

	input := domain.TaskInput{
		ProjectID:      projectID,
		WorkspaceID:    workspaceID,
		OwnerID:        authUser.UserID,
		BusinessID:     businessID,
		ParentTaskID:   req.ParentTaskID,
		Title:          req.Title,
		Description:    req.Description,
		Priority:       req.Priority,
		AssigneeID:     jsonIDPtrToInt64Ptr(req.AssigneeID),
		ReporterID:     authUser.UserID,
		DueDate:        req.DueDate,
		StartDate:      req.StartDate,
		EstimatedHours: req.EstimatedHours,
		Labels:         req.Labels,
	}

	task, createErr := h.FlowService.CreateTask(ctx, input)
	if createErr != nil {
		return CreateTaskResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to create task",
		}
	}

	payload := toTaskPayload(task)
	publishTaskEvent("task.created", payload, authUser.UserID)
	return CreateTaskResponse{Task: payload}, nil
}

// ─────────────────────────────────────────────
// GET /flow/tasks/detail
// ─────────────────────────────────────────────

type GetTaskDetailHandler struct {
	FlowService *application.Service
}

type GetTaskDetailRequest struct {
	ID JSONID `json:"id"`
}

type GetTaskDetailResponse struct {
	Task     *flowtaskv1.Task           `json:"task"`
	Subtasks []*flowtaskv1.Task         `json:"subtasks"`
	Comments []*flowtaskv1.TaskComment  `json:"comments"`
	Activity []*flowtaskv1.TaskActivity `json:"activity"`
}

func (h *GetTaskDetailHandler) Validate(_ context.Context, req *GetTaskDetailRequest) *errorModels.ServerError {
	if req.ID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Task ID is required"}
	}

	return nil
}

func (h *GetTaskDetailHandler) Serve(ctx context.Context, req *GetTaskDetailRequest) (GetTaskDetailResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetTaskDetailResponse{}, errorModels.PermissionDeniedServerError()
	}

	if !h.FlowService.IsTaskMember(ctx, authUser.UserID, req.ID.Int64()) {
		return GetTaskDetailResponse{}, errorModels.PermissionDeniedServerError()
	}

	detail, err := h.FlowService.GetTaskDetail(ctx, req.ID.Int64())
	if err != nil {
		return GetTaskDetailResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to get task detail",
		}
	}

	subtasks := make([]*flowtaskv1.Task, 0, len(detail.Subtasks))
	for _, t := range detail.Subtasks {
		subtasks = append(subtasks, toTaskPayload(t))
	}

	comments := make([]*flowtaskv1.TaskComment, 0, len(detail.Comments))
	for _, c := range detail.Comments {
		comments = append(comments, toCommentPayload(c))
	}

	activity := make([]*flowtaskv1.TaskActivity, 0, len(detail.Activity))
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
	ID          JSONID     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssigneeID  *JSONID    `json:"assignee_id"`
	DueDate     *time.Time `json:"due_date"`
	StartDate   *time.Time `json:"start_date"`
	Labels      []int64    `json:"labels"`
	Archived    *bool      `json:"archived"`
}

type UpdateTaskResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTaskHandler) Validate(_ context.Context, req *UpdateTaskRequest) *errorModels.ServerError {
	if req.ID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Task ID is required"}
	}

	// This is a partial update (PATCH semantics over PUT): the client sends only
	// the fields it wants to change (e.g. just priority or assignee). An omitted
	// title means "leave unchanged", not "set empty", so we don't require it here
	// — the repository preserves any field the request leaves blank/nil.

	return nil
}

func (h *UpdateTaskHandler) Serve(ctx context.Context, req *UpdateTaskRequest) (UpdateTaskResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateTaskResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateTaskResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.UpdateTaskInput{
		ID:          req.ID.Int64(),
		OwnerID:     authUser.UserID,
		BusinessID:  businessID,
		Title:       req.Title,
		Description: req.Description,
		Status:      req.Status,
		Priority:    req.Priority,
		AssigneeID:  jsonIDPtrToInt64Ptr(req.AssigneeID),
		DueDate:     req.DueDate,
		StartDate:   req.StartDate,
		Labels:      req.Labels,
		Archived:    req.Archived,
	}

	if err := h.FlowService.UpdateTask(ctx, input, authUser.UserID); err != nil {
		return UpdateTaskResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update task",
		}
	}

	// Broadcast to subscribers of the owning project. We refetch just to
	// learn the project id — cheap, and avoids plumbing it through the
	// service return signature.
	if updated, fetchErr := h.FlowService.GetTaskDetail(ctx, input.ID); fetchErr == nil {
		publishTaskEventByProject("task.updated", updated.Task.ProjectID, input.ID, authUser.UserID)
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
	ID JSONID `json:"id"`
}

type DeleteTaskResponse struct {
	Message string `json:"message"`
}

func (h *DeleteTaskHandler) Validate(_ context.Context, req *DeleteTaskRequest) *errorModels.ServerError {
	if req.ID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Task ID is required"}
	}

	return nil
}

func (h *DeleteTaskHandler) Serve(ctx context.Context, req *DeleteTaskRequest) (DeleteTaskResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteTaskResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteTaskResponse{}, errorModels.PermissionDeniedServerError()
	}

	// Capture the project id BEFORE the row is deleted so the realtime
	// event has a valid topic. If the fetch fails we skip the publish.
	var projectID int64
	if existing, fetchErr := h.FlowService.GetTaskDetail(ctx, req.ID.Int64()); fetchErr == nil {
		projectID = existing.Task.ProjectID
	}

	if err := h.FlowService.DeleteTask(ctx, req.ID.Int64(), businessID); err != nil {
		return DeleteTaskResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete task",
		}
	}

	publishTaskEventByProject("task.deleted", projectID, req.ID.Int64(), authUser.UserID)
	return DeleteTaskResponse{Message: "Task deleted successfully"}, nil
}

// ─────────────────────────────────────────────
// PUT /flow/tasks/status
// ─────────────────────────────────────────────

type UpdateTaskStatusHandler struct {
	FlowService *application.Service
}

type UpdateTaskStatusRequest struct {
	ID     JSONID `json:"id"`
	Status string `json:"status"`
}

type UpdateTaskStatusResponse struct {
	Message string `json:"message"`
}

func (h *UpdateTaskStatusHandler) Validate(_ context.Context, req *UpdateTaskStatusRequest) *errorModels.ServerError {
	if req.ID.IsZero() {
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
	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateTaskStatusResponse{}, errorModels.PermissionDeniedServerError()
	}

	input := domain.UpdateTaskStatusInput{
		ID:         req.ID.Int64(),
		OwnerID:    authUser.UserID,
		BusinessID: businessID,
		Status:  req.Status,
	}

	if err := h.FlowService.UpdateTaskStatus(ctx, input, authUser.UserID); err != nil {
		return UpdateTaskStatusResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to update task status",
		}
	}

	if updated, fetchErr := h.FlowService.GetTaskDetail(ctx, req.ID.Int64()); fetchErr == nil {
		publishTaskEventByProject(
			"task.status_changed",
			updated.Task.ProjectID,
			req.ID.Int64(),
			authUser.UserID,
		)
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
	Tasks []*flowtaskv1.Task `json:"tasks"`
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

	payloads := make([]*flowtaskv1.Task, 0, len(tasks))
	for _, t := range tasks {
		payloads = append(payloads, toTaskPayload(t))
	}

	return GetMyTasksResponse{Tasks: payloads}, nil
}

// ─────────────────────────────────────────────
// GET /flow/my-tasks-project
// Get-or-create the caller's personal "My Tasks" project board and return its
// ids, so the Flow app can land the user directly on it.
// ─────────────────────────────────────────────

func GetMyTasksProjectRoutes() []string { return []string{"/flow/my-tasks-project"} }

type GetMyTasksProjectRequest struct{}

type GetMyTasksProjectResponse struct {
	ProjectID   int64 `json:"project_id"`
	WorkspaceID int64 `json:"workspace_id"`
}

type GetMyTasksProjectHandler struct {
	FlowService *application.Service
}

func (h *GetMyTasksProjectHandler) Serve(ctx context.Context, _ *GetMyTasksProjectRequest) (GetMyTasksProjectResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return GetMyTasksProjectResponse{}, errorModels.PermissionDeniedServerError()
	}

	projectID, workspaceID, err := h.FlowService.EnsureMyTasksProject(ctx, authUser.UserID, 0)
	if err != nil || projectID == 0 {
		return GetMyTasksProjectResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to resolve My Tasks project",
		}
	}

	return GetMyTasksProjectResponse{ProjectID: projectID, WorkspaceID: workspaceID}, nil
}

// ─────────────────────────────────────────────
// POST /flow/tasks/comment
// ─────────────────────────────────────────────

type AddCommentHandler struct {
	FlowService *application.Service
}

type AddCommentRequest struct {
	TaskID  JSONID `json:"task_id"`
	Content string `json:"content"`
}

type AddCommentResponse struct {
	Comment *flowtaskv1.TaskComment `json:"comment"`
}

func (h *AddCommentHandler) Validate(_ context.Context, req *AddCommentRequest) *errorModels.ServerError {
	if req.TaskID.IsZero() {
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

	if !h.FlowService.IsTaskMember(ctx, authUser.UserID, req.TaskID.Int64()) {
		return AddCommentResponse{}, errorModels.PermissionDeniedServerError()
	}

	comment, err := h.FlowService.AddComment(ctx, req.TaskID.Int64(), authUser.UserID, req.Content)
	if err != nil {
		return AddCommentResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to add comment",
		}
	}

	// Fan out to the task channel so an open TaskDetail screen refreshes
	// its comments without polling. We scope to the task topic only —
	// comment events aren't shown on board/list/timeline projections.
	publishCommentEvent("comment.added", req.TaskID.Int64(), authUser.UserID)
	return AddCommentResponse{Comment: toCommentPayload(comment)}, nil
}

// ─────────────────────────────────────────────
// DELETE /flow/tasks/comment
// ─────────────────────────────────────────────

type DeleteCommentHandler struct {
	FlowService *application.Service
}

type DeleteCommentRequest struct {
	ID JSONID `json:"id"`
}

type DeleteCommentResponse struct {
	Message string `json:"message"`
}

func (h *DeleteCommentHandler) Validate(_ context.Context, req *DeleteCommentRequest) *errorModels.ServerError {
	if req.ID.IsZero() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "Comment ID is required"}
	}

	return nil
}

func (h *DeleteCommentHandler) Serve(ctx context.Context, req *DeleteCommentRequest) (DeleteCommentResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteCommentResponse{}, errorModels.PermissionDeniedServerError()
	}

	// Capture the owning task before the row is gone so the realtime
	// event can route to the correct topic.
	taskID := h.FlowService.GetCommentTaskID(ctx, req.ID.Int64())

	if err := h.FlowService.DeleteComment(ctx, req.ID.Int64(), authUser.UserID); err != nil {
		return DeleteCommentResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to delete comment",
		}
	}

	publishCommentEvent("comment.deleted", taskID, authUser.UserID)
	return DeleteCommentResponse{Message: "Comment deleted successfully"}, nil
}

// ─────────────────────────────────────────────
// GET /flow/tasks/search
// ─────────────────────────────────────────────

type SearchTasksHandler struct {
	FlowService *application.Service
}

type SearchTasksRequest struct {
	Q           string           `json:"q"`
	WorkspaceID WorkspaceIDField `json:"workspace_id"`
}

type SearchTasksResponse struct {
	Tasks []*flowtaskv1.Task `json:"tasks"`
}

func (h *SearchTasksHandler) Validate(_ context.Context, req *SearchTasksRequest) *errorModels.ServerError {
	if req.WorkspaceID.IsEmpty() {
		return &errorModels.ServerError{Code: http.StatusBadRequest, Message: "workspace_id is required"}
	}
	return nil
}

func (h *SearchTasksHandler) Serve(ctx context.Context, req *SearchTasksRequest) (SearchTasksResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return SearchTasksResponse{}, errorModels.PermissionDeniedServerError()
	}

	workspaceID, err := h.FlowService.ResolveWorkspaceID(ctx, authUser.UserID, req.WorkspaceID.String())
	if err != nil || workspaceID == 0 {
		return SearchTasksResponse{}, &errorModels.ServerError{
			Code:    http.StatusBadRequest,
			Message: "Invalid workspace_id",
		}
	}

	tasks, err := h.FlowService.SearchTasks(ctx, workspaceID, req.Q)
	if err != nil {
		return SearchTasksResponse{}, &errorModels.ServerError{
			Code:    http.StatusInternalServerError,
			Message: "Failed to search tasks",
		}
	}

	payloads := make([]*flowtaskv1.Task, 0, len(tasks))
	for _, t := range tasks {
		payloads = append(payloads, toTaskPayload(t))
	}

	return SearchTasksResponse{Tasks: payloads}, nil
}
