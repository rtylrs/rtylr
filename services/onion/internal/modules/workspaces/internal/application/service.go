package application

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/wsbus"
)

const (
	workspacesScope = "workspaces"
	workspacesTTL   = 2 * time.Minute

	// internalListCap bounds list reads made on behalf of another feature
	// (the dashboard's "my tasks" widget, a task's subtasks embedded in its
	// detail view, a project's columns re-fetched right after creating one)
	// rather than a paginated list RPC. Matches the cap these reads already
	// used before pagination existed.
	internalListCap = int64(500)
)

func invalidateWorkspacesScope(ctx context.Context, businessID int64) {
	if cache := cacher.GetRedisCache(); cache != nil {
		cache.DeleteByPattern(ctx, cacheKeys.ScopedPrefix(workspacesScope, businessID))
	}
}

type Service struct {
	repo *infrastructure.Repository
}

func NewService(_ context.Context, db *sql.DB) *Service {
	repo := infrastructure.NewRepository(db)

	return &Service{
		repo: repo,
	}
}

// ─────────────────────────────────────────────
// Workspaces
// ─────────────────────────────────────────────

type workspacesPage struct {
	Workspaces []domain.Workspace
	Total      int64
}

func (s *Service) ListWorkspaces(ctx context.Context, businessID, page, limit int64) ([]domain.Workspace, int64, error) {
	offset := (page - 1) * limit
	result, err := cacher.Of[workspacesPage]().GetOrComputeSingle(ctx,
		// page/limit are part of the cache key, not just the fetch - otherwise
		// page 2 would return page 1's cached rows until the TTL expires.
		cacheKeys.Scoped(workspacesScope, businessID, fmt.Sprintf("list:page=%d:limit=%d", page, limit)),
		func(ctx context.Context) (workspacesPage, error) {
			workspaces, total, fetchErr := s.repo.GetWorkspacesByOwner(ctx, businessID, limit, offset)
			return workspacesPage{Workspaces: workspaces, Total: total}, fetchErr
		},
		cacher.WithExpiry(ctx, workspacesTTL),
	)
	if err != nil {
		return nil, 0, err
	}
	return result.Workspaces, result.Total, nil
}

func (s *Service) CreateWorkspace(ctx context.Context, input domain.WorkspaceInput) (domain.Workspace, error) {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	count, err := s.repo.CountWorkspacesByOwner(ctx, input.BusinessID)
	if err != nil {
		return domain.Workspace{}, err
	}

	isDefault := count == 0

	id, err := s.repo.CreateWorkspace(ctx, input, isDefault)
	if err != nil {
		return domain.Workspace{}, err
	}

	return domain.Workspace{
		ID:          id,
		BusinessID:  input.BusinessID,
		OwnerID:     input.OwnerID,
		Name:        input.Name,
		Description: input.Description,
		Icon:        input.Icon,
		Color:       input.Color,
		IsDefault:   isDefault,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (s *Service) UpdateWorkspace(ctx context.Context, input domain.UpdateWorkspaceInput) error {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	return s.repo.UpdateWorkspace(ctx, input)
}

func (s *Service) DeleteWorkspace(ctx context.Context, id, businessID int64) error {
	defer invalidateWorkspacesScope(ctx, businessID)
	return s.repo.DeleteWorkspace(ctx, id, businessID)
}

func (s *Service) AddWorkspaceMember(ctx context.Context, input domain.AddWorkspaceMemberInput) error {
	return s.repo.AddWorkspaceMember(ctx, input)
}

func (s *Service) RemoveWorkspaceMember(ctx context.Context, input domain.RemoveWorkspaceMemberInput) error {
	return s.repo.RemoveWorkspaceMember(ctx, input)
}

func (s *Service) ListWorkspaceMembers(ctx context.Context, workspaceID, limit, offset int64) ([]domain.WorkspaceMember, int64, error) {
	return s.repo.ListWorkspaceMembers(ctx, workspaceID, limit, offset)
}

func (s *Service) UpdateWorkspaceMemberRole(ctx context.Context, input domain.UpdateWorkspaceMemberRoleInput) error {
	return s.repo.UpdateWorkspaceMemberRole(ctx, input)
}

func (s *Service) CreatePersonalWorkspace(ctx context.Context, userID int64) (int64, error) {
	return s.repo.CreatePersonalWorkspace(ctx, userID)
}

// ResolveWorkspaceID converts a workspace_id coming from the frontend into
// a numeric DB ID. The frontend uses the sentinel string "personal" to mean
// "the user's Personal workspace", which is created on demand. Any numeric
// string (or empty) is parsed as-is; numbers round-trip unchanged.
func (s *Service) ResolveWorkspaceID(ctx context.Context, userID int64, raw string) (int64, error) {
	if raw == "" {
		return 0, nil
	}
	if raw == "personal" {
		return s.repo.CreatePersonalWorkspace(ctx, userID)
	}
	return strconv.ParseInt(raw, 10, 64)
}

// IsWorkspaceMember reports whether userID is on the workspace's member
// roster or owns it. Used by the realtime hub to authorize
// `workspace:<id>` and `project:<id>` topic subscriptions.
func (s *Service) IsWorkspaceMember(ctx context.Context, userID, workspaceID int64) bool {
	return s.repo.IsWorkspaceMember(ctx, userID, workspaceID)
}

// IsProjectMember reports whether userID can access the project — either
// because they're a project member, a workspace member of the project's
// workspace, or its owner. Used for realtime topic auth.
func (s *Service) IsProjectMember(ctx context.Context, userID, projectID int64) bool {
	return s.repo.IsProjectMember(ctx, userID, projectID)
}

// IsTaskMember reports whether userID can subscribe to a task realtime topic.
func (s *Service) IsTaskMember(ctx context.Context, userID, taskID int64) bool {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return false
	}
	if task.ProjectID != 0 {
		return s.repo.IsProjectMember(ctx, userID, task.ProjectID)
	}
	if task.WorkspaceID != 0 {
		return s.repo.IsWorkspaceMember(ctx, userID, task.WorkspaceID)
	}
	return task.OwnerID == userID
}

// ─────────────────────────────────────────────
// Projects
// ─────────────────────────────────────────────

func (s *Service) ListProjects(ctx context.Context, workspaceID, businessID, limit, offset int64) ([]domain.Project, int64, error) {
	return s.repo.GetProjectsByWorkspace(ctx, workspaceID, businessID, limit, offset)
}

func (s *Service) CreateProject(ctx context.Context, input domain.ProjectInput) (domain.Project, error) {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	id, err := s.repo.CreateProject(ctx, input)
	if err != nil {
		return domain.Project{}, err
	}

	// Seed default columns for every new project
	_ = s.repo.SeedDefaultColumns(ctx, id, input.WorkspaceID, input.CreatedBy)

	return domain.Project{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		OwnerID:     input.OwnerID,
		BusinessID:  input.BusinessID,
		Name:        input.Name,
		Description: input.Description,
		Icon:        input.Icon,
		Color:       input.Color,
		Status:      "active",
		StartDate:   input.StartDate,
		DueDate:     input.DueDate,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (s *Service) UpdateProject(ctx context.Context, input domain.UpdateProjectInput) error {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	return s.repo.UpdateProject(ctx, input)
}

func (s *Service) DeleteProject(ctx context.Context, id, businessID int64) error {
	defer invalidateWorkspacesScope(ctx, businessID)
	return s.repo.ArchiveProject(ctx, id, businessID)
}

func (s *Service) ListProjectMembers(ctx context.Context, projectID, limit, offset int64) ([]domain.ProjectMember, int64, error) {
	return s.repo.ListProjectMembers(ctx, projectID, limit, offset)
}

func (s *Service) AddProjectMember(ctx context.Context, input domain.AddProjectMemberInput) error {
	return s.repo.AddProjectMember(ctx, input)
}

func (s *Service) RemoveProjectMember(ctx context.Context, input domain.RemoveProjectMemberInput) error {
	return s.repo.RemoveProjectMember(ctx, input)
}

func (s *Service) UpdateProjectMemberRole(ctx context.Context, input domain.UpdateProjectMemberRoleInput) error {
	return s.repo.UpdateProjectMemberRole(ctx, input)
}

func (s *Service) TransferProjectOwnership(ctx context.Context, input domain.TransferProjectOwnershipInput) error {
	return s.repo.TransferProjectOwnership(ctx, input)
}

func (s *Service) MoveProject(ctx context.Context, input domain.MoveProjectInput) error {
	return s.repo.MoveProject(ctx, input)
}

func (s *Service) InviteProjectMemberByEmail(ctx context.Context, projectID int64, email, role string, addedBy int64) (bool, error) {
	userID, found, err := s.repo.LookupUserByEmail(ctx, email)
	if err != nil {
		return false, err
	}
	if !found {
		return false, nil
	}
	return true, s.repo.AddProjectMember(ctx, domain.AddProjectMemberInput{
		ProjectID: projectID,
		UserID:    userID,
		Role:      role,
		AddedBy:   addedBy,
	})
}

// ─────────────────────────────────────────────
// Tasks
// ─────────────────────────────────────────────

func (s *Service) ListTasks(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, int64, error) {
	if filter.ProjectID == 0 && filter.WorkspaceID != 0 {
		return s.repo.GetTasksByWorkspace(ctx, filter)
	}
	return s.repo.GetTasksByProject(ctx, filter)
}

func (s *Service) CreateTask(ctx context.Context, input domain.TaskInput) (domain.Task, error) {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	id, err := s.repo.CreateTask(ctx, input)
	if err != nil {
		return domain.Task{}, err
	}

	if actErr := s.repo.LogActivity(ctx, id, input.ReporterID, "created_task", "", input.Title); actErr != nil {
		logger.Ctx(ctx).Error("CreateTask: failed to log activity", logger.Error(actErr))
	}

	return domain.Task{
		ID:           id,
		ProjectID:    input.ProjectID,
		WorkspaceID:  input.WorkspaceID,
		OwnerID:      input.OwnerID,
		ParentTaskID: input.ParentTaskID,
		Title:        input.Title,
		Description:  input.Description,
		Status:       "todo",
		Priority:     input.Priority,
		AssigneeID:   input.AssigneeID,
		ReporterID:   input.ReporterID,
		DueDate:      input.DueDate,
		StartDate:    input.StartDate,
		Labels:       input.Labels,
	}, nil
}

func (s *Service) GetTaskDetail(ctx context.Context, taskID int64) (domain.TaskDetail, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	subtasks, _, err := s.repo.GetSubtasks(ctx, taskID, internalListCap, 0)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	comments, err := s.repo.GetCommentsByTask(ctx, taskID)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	activity, err := s.repo.GetActivityByTask(ctx, taskID, 50)
	if err != nil {
		return domain.TaskDetail{}, err
	}

	return domain.TaskDetail{
		Task:     task,
		Subtasks: subtasks,
		Comments: comments,
		Activity: activity,
	}, nil
}

func (s *Service) UpdateTask(ctx context.Context, input domain.UpdateTaskInput, userID int64) error {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	existing, err := s.repo.GetTaskByID(ctx, input.ID)
	if err != nil {
		return err
	}

	// This is a partial update — only log a change for a field the request
	// actually supplied. An omitted field arrives as the zero value (empty
	// string / nil) and means "leave unchanged", so comparing it against the
	// existing value would otherwise log a spurious "changed to empty" entry.
	if input.Status != "" && existing.Status != input.Status {
		if actErr := s.repo.LogActivity(ctx, input.ID, userID, "status_changed", existing.Status, input.Status); actErr != nil {
			logger.Ctx(ctx).Error("UpdateTask: failed to log status activity", logger.Error(actErr))
		}
	}

	if input.Priority != "" && existing.Priority != input.Priority {
		if actErr := s.repo.LogActivity(ctx, input.ID, userID, "priority_changed", existing.Priority, input.Priority); actErr != nil {
			logger.Ctx(ctx).Error("UpdateTask: failed to log priority activity", logger.Error(actErr))
		}
	}

	if input.AssigneeID != nil {
		existingAssignee := int64(0)
		if existing.AssigneeID != nil {
			existingAssignee = *existing.AssigneeID
		}
		newAssignee := *input.AssigneeID

		if existingAssignee != newAssignee {
			if actErr := s.repo.LogActivity(ctx, input.ID, userID, "assignee_changed",
				strconv.FormatInt(existingAssignee, 10),
				strconv.FormatInt(newAssignee, 10),
			); actErr != nil {
				logger.Ctx(ctx).Error("UpdateTask: failed to log assignee activity", logger.Error(actErr))
			}
		}
	}

	return s.repo.UpdateTask(ctx, input)
}

func (s *Service) UpdateTaskStatus(ctx context.Context, input domain.UpdateTaskStatusInput, userID int64) error {
	existing, err := s.repo.GetTaskByID(ctx, input.ID)
	if err != nil {
		return err
	}

	if existing.Status != input.Status {
		if actErr := s.repo.LogActivity(ctx, input.ID, userID, "status_changed", existing.Status, input.Status); actErr != nil {
			logger.Ctx(ctx).Error("UpdateTaskStatus: failed to log activity", logger.Error(actErr))
		}
	}

	return s.repo.UpdateTaskStatus(ctx, input)
}

func (s *Service) DeleteTask(ctx context.Context, taskID, businessID int64) error {
	defer invalidateWorkspacesScope(ctx, businessID)
	return s.repo.DeleteTask(ctx, taskID, businessID)
}

type myTasksPage struct {
	Tasks []domain.Task
	Total int64
}

func (s *Service) GetMyTasks(ctx context.Context, userID, page, limit int64) ([]domain.Task, int64, error) {
	offset := (page - 1) * limit
	result, err := cacher.Of[myTasksPage]().GetOrComputeSingle(ctx,
		// page/limit are part of the cache key, not just the fetch - otherwise
		// page 2 would return page 1's cached rows until the 2-minute TTL expires.
		cacheKeys.Scoped("user", userID, fmt.Sprintf("workspace_tasks:page=%d:limit=%d", page, limit)),
		func(ctx context.Context) (myTasksPage, error) {
			tasks, total, fetchErr := s.repo.GetMyTasks(ctx, userID, limit, offset)
			return myTasksPage{Tasks: tasks, Total: total}, fetchErr
		},
		cacher.WithExpiry(ctx, workspacesTTL),
	)
	if err != nil {
		return nil, 0, err
	}
	return result.Tasks, result.Total, nil
}

func (s *Service) SearchTasks(ctx context.Context, workspaceID int64, q string, limit, offset int64) ([]domain.Task, int64, error) {
	return s.repo.SearchTasks(ctx, workspaceID, q, limit, offset)
}

// ─────────────────────────────────────────────
// Comments
// ─────────────────────────────────────────────

func (s *Service) AddComment(ctx context.Context, taskID, userID int64, content string) (domain.TaskComment, error) {
	id, err := s.repo.CreateComment(ctx, taskID, userID, content)
	if err != nil {
		return domain.TaskComment{}, err
	}

	if actErr := s.repo.LogActivity(ctx, taskID, userID, "comment_added", "", content); actErr != nil {
		logger.Ctx(ctx).Error("AddComment: failed to log activity", logger.Error(actErr))
	}

	return domain.TaskComment{
		ID:      id,
		TaskID:  taskID,
		UserID:  userID,
		Content: content,
	}, nil
}

func (s *Service) DeleteComment(ctx context.Context, commentID, userID int64) error {
	return s.repo.DeleteComment(ctx, commentID, userID)
}

// GetCommentTaskID returns the task a comment belongs to. Handlers call
// this before DeleteComment so the realtime event can route to the
// right topic after the row is gone.
func (s *Service) GetCommentTaskID(ctx context.Context, commentID int64) int64 {
	taskID, err := s.repo.GetCommentTaskID(ctx, commentID)
	if err != nil {
		return 0
	}
	return taskID
}

// ─────────────────────────────────────────────
// Dashboard
// ─────────────────────────────────────────────

func (s *Service) GetDashboard(ctx context.Context, workspaceID, businessID, userID int64) (domain.DashboardStats, error) {
	stats, err := s.repo.GetDashboardStats(ctx, workspaceID, businessID)
	if err != nil {
		return domain.DashboardStats{}, err
	}

	myTasks, _, err := s.repo.GetMyTasks(ctx, userID, internalListCap, 0)
	if err != nil {
		return domain.DashboardStats{}, err
	}

	stats.MyTasks = myTasks

	return stats, nil
}

// ─────────────────────────────────────────────
// Task Activity (standalone)
// ─────────────────────────────────────────────

func (s *Service) GetTaskActivity(ctx context.Context, taskID int64) ([]domain.TaskActivity, error) {
	return s.repo.GetActivityByTask(ctx, taskID, 100)
}

// ─────────────────────────────────────────────
// Task Subtasks (standalone)
// ─────────────────────────────────────────────

func (s *Service) GetSubtasks(ctx context.Context, taskID, limit, offset int64) ([]domain.Task, int64, error) {
	return s.repo.GetSubtasks(ctx, taskID, limit, offset)
}

func (s *Service) CreateSubtask(ctx context.Context, input domain.TaskInput) (domain.Task, error) {
	return s.CreateTask(ctx, input)
}

// ─────────────────────────────────────────────
// Task Attachments
// ─────────────────────────────────────────────

func (s *Service) GetAttachments(ctx context.Context, taskID int64) ([]domain.TaskAttachment, error) {
	return s.repo.GetAttachmentsByTask(ctx, taskID)
}

func (s *Service) CreateAttachment(ctx context.Context, input domain.TaskAttachmentInput) (int64, error) {
	return s.repo.CreateAttachment(ctx, input)
}

func (s *Service) DeleteAttachment(ctx context.Context, attachmentID, businessID int64) error {
	return s.repo.DeleteAttachment(ctx, attachmentID, businessID)
}

// ─────────────────────────────────────────────
// Time Logs
// ─────────────────────────────────────────────

func (s *Service) GetTimeLogs(ctx context.Context, taskID int64) ([]domain.TimeLog, error) {
	return s.repo.GetTimeLogsByTask(ctx, taskID)
}

func (s *Service) CreateTimeLog(ctx context.Context, input domain.TimeLogInput) (int64, error) {
	id, err := s.repo.CreateTimeLog(ctx, input)
	if err != nil {
		return 0, err
	}
	if recalcErr := s.repo.RecalcActualHours(ctx, input.TaskID, input.UserID); recalcErr != nil {
		logger.Ctx(ctx).Error("CreateTimeLog: failed to recalc actual_hours", logger.Error(recalcErr))
	}
	return id, nil
}

func (s *Service) DeleteTimeLog(ctx context.Context, id, userID, taskID int64) error {
	if err := s.repo.DeleteTimeLog(ctx, id, userID); err != nil {
		return err
	}
	if recalcErr := s.repo.RecalcActualHours(ctx, taskID, userID); recalcErr != nil {
		logger.Ctx(ctx).Error("DeleteTimeLog: failed to recalc actual_hours", logger.Error(recalcErr))
	}
	return nil
}

// DeleteTimeLogFlat deletes a time log by ID, looking up the task_id automatically.
func (s *Service) DeleteTimeLogFlat(ctx context.Context, id, userID int64) error {
	taskID, err := s.repo.GetTimeLogTaskID(ctx, id)
	if err != nil {
		return err
	}
	return s.DeleteTimeLog(ctx, id, userID, taskID)
}

// ─────────────────────────────────────────────
// Labels
// ─────────────────────────────────────────────

func (s *Service) ListLabels(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Label, int64, error) {
	return s.repo.GetLabelsByWorkspace(ctx, workspaceID, limit, offset)
}

func (s *Service) CreateLabel(ctx context.Context, input domain.LabelInput) (domain.Label, error) {
	id, err := s.repo.CreateLabel(ctx, input)
	if err != nil {
		return domain.Label{}, err
	}

	return domain.Label{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Color:       input.Color,
	}, nil
}

func (s *Service) UpdateLabel(ctx context.Context, id, actorID int64, input domain.LabelInput) error {
	return s.repo.UpdateLabel(ctx, id, actorID, input)
}

func (s *Service) DeleteLabel(ctx context.Context, id, actorID int64) error {
	return s.repo.DeleteLabel(ctx, id, actorID)
}

func (s *Service) SetTaskLabels(ctx context.Context, taskID, actorID int64, labelIDs []int64) error {
	return s.repo.SetTaskLabels(ctx, taskID, actorID, labelIDs)
}

// ─────────────────────────────────────────────
// Cycles
// ─────────────────────────────────────────────

func (s *Service) ListCycles(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Cycle, int64, error) {
	return s.repo.ListCycles(ctx, workspaceID, limit, offset)
}

func (s *Service) CreateCycle(ctx context.Context, input domain.CycleInput) (domain.Cycle, error) {
	id, err := s.repo.CreateCycle(ctx, input)
	if err != nil {
		return domain.Cycle{}, err
	}
	status := "upcoming"
	now := time.Now()
	if now.After(input.StartDate) && now.Before(input.EndDate) {
		status = "active"
	}
	return domain.Cycle{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: input.Description,
		Status:      status,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (s *Service) UpdateCycle(ctx context.Context, input domain.UpdateCycleInput) error {
	return s.repo.UpdateCycle(ctx, input)
}

// ─────────────────────────────────────────────
// Milestones
// ─────────────────────────────────────────────

func (s *Service) ListMilestones(ctx context.Context, workspaceID int64, projectID *int64, limit, offset int64) ([]domain.Milestone, int64, error) {
	return s.repo.ListMilestones(ctx, workspaceID, projectID, limit, offset)
}

func (s *Service) CreateMilestone(ctx context.Context, input domain.MilestoneInput) (domain.Milestone, error) {
	id, err := s.repo.CreateMilestone(ctx, input)
	if err != nil {
		return domain.Milestone{}, err
	}
	return domain.Milestone{
		ID:          id,
		ProjectID:   input.ProjectID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Description: input.Description,
		Status:      "planning",
		DueDate:     input.DueDate,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (s *Service) UpdateMilestone(ctx context.Context, input domain.UpdateMilestoneInput) error {
	return s.repo.UpdateMilestone(ctx, input)
}

// ─────────────────────────────────────────────
// Views
// ─────────────────────────────────────────────

func (s *Service) ListViews(ctx context.Context, workspaceID, limit, offset int64) ([]domain.View, int64, error) {
	return s.repo.ListViews(ctx, workspaceID, limit, offset)
}

func (s *Service) CreateView(ctx context.Context, input domain.ViewInput) (domain.View, error) {
	id, err := s.repo.CreateView(ctx, input)
	if err != nil {
		return domain.View{}, err
	}
	return domain.View{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Icon:        input.Icon,
		Filters:     input.Filters,
		GroupBy:     input.GroupBy,
		SortBy:      input.SortBy,
		SortDir:     input.SortDir,
		CreatedBy:   input.CreatedBy,
	}, nil
}

func (s *Service) UpdateView(ctx context.Context, input domain.UpdateViewInput) error {
	return s.repo.UpdateView(ctx, input)
}

func (s *Service) DeleteView(ctx context.Context, id, actorID int64) error {
	return s.repo.DeleteView(ctx, id, actorID)
}

// ─────────────────────────────────────────────
// Insights
// ─────────────────────────────────────────────

func (s *Service) GetInsights(ctx context.Context, workspaceID int64) (domain.Insights, error) {
	return s.repo.GetInsights(ctx, workspaceID)
}

// ─────────────────────────────────────────────
// Team Workload
// ─────────────────────────────────────────────

func (s *Service) GetTeamWorkload(ctx context.Context, workspaceID, limit, offset int64) ([]domain.TeamMemberWorkload, int64, error) {
	return s.repo.GetTeamWorkload(ctx, workspaceID, limit, offset)
}

// ─────────────────────────────────────────────
// Roadmap
// ─────────────────────────────────────────────

func (s *Service) GetRoadmap(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Project, int64, error) {
	return s.repo.GetRoadmap(ctx, workspaceID, limit, offset)
}

// ─────────────────────────────────────────────
// Notifications
// ─────────────────────────────────────────────

func (s *Service) ListNotifications(ctx context.Context, userID int64) ([]domain.Notification, error) {
	return s.repo.ListNotifications(ctx, userID)
}

func (s *Service) MarkNotificationRead(ctx context.Context, id, userID int64) error {
	return s.repo.MarkNotificationRead(ctx, id, userID)
}

func (s *Service) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	return s.repo.MarkAllNotificationsRead(ctx, userID)
}

func (s *Service) CreateNotification(ctx context.Context, input domain.CreateNotificationInput) error {
	if err := s.repo.CreateNotification(ctx, input); err != nil {
		return err
	}
	// Push a hint to the recipient's user topic so their inbox badge
	// and list update in real time. The entity id is the recipient
	// rather than the notification row — consumers invalidate their
	// `['notifications']` query rather than patching by id.
	if input.UserID != 0 {
		wsbus.Publish(wsbus.Event{
			Topic:    fmt.Sprintf("user:%d", input.UserID),
			Type:     "notification.created",
			EntityID: input.UserID,
		})
	}
	return nil
}

// ─────────────────────────────────────────────
// User Settings
// ─────────────────────────────────────────────

func (s *Service) GetUserSettings(ctx context.Context, userID int64) (domain.UserSettings, error) {
	return s.repo.GetUserSettings(ctx, userID)
}

func (s *Service) UpdateUserSettings(ctx context.Context, settings domain.UserSettings) error {
	return s.repo.UpsertUserSettings(ctx, settings)
}

// ─────────────────────────────────────────────
// Project task counts (for roadmap)
// ─────────────────────────────────────────────

func (s *Service) GetProjectTaskCounts(ctx context.Context, projectID int64) (total, done int) {
	return s.repo.GetProjectTaskCounts(ctx, projectID)
}

// ─────────────────────────────────────────────
// Project Columns
// ─────────────────────────────────────────────

func (s *Service) GetColumnsByProject(ctx context.Context, projectID, limit, offset int64) ([]domain.ProjectColumn, int64, error) {
	return s.repo.GetColumnsByProject(ctx, projectID, limit, offset)
}

func (s *Service) CreateColumn(ctx context.Context, input domain.ProjectColumnInput) (domain.ProjectColumn, error) {
	id, err := s.repo.CreateColumn(ctx, input)
	if err != nil {
		return domain.ProjectColumn{}, err
	}
	cols, _, err := s.repo.GetColumnsByProject(ctx, input.ProjectID, internalListCap, 0)
	if err != nil {
		return domain.ProjectColumn{}, err
	}
	for _, c := range cols {
		if c.ID == id {
			return c, nil
		}
	}
	return domain.ProjectColumn{
		ID: id, ProjectID: input.ProjectID, WorkspaceID: input.WorkspaceID,
		Name: input.Name, Color: input.Color, SortOrder: input.SortOrder,
	}, nil
}

func (s *Service) UpdateColumn(ctx context.Context, input domain.UpdateProjectColumnInput) error {
	return s.repo.UpdateColumn(ctx, input)
}

func (s *Service) DeleteColumn(ctx context.Context, id, actorID int64) error {
	return s.repo.DeleteColumn(ctx, id, actorID)
}

func (s *Service) MoveTaskToColumn(ctx context.Context, taskID, columnID, actorID int64) error {
	return s.repo.MoveTaskToColumn(ctx, taskID, columnID, actorID)
}

// GetColumnProject returns the project id of the given column, or 0 if
// the lookup fails. Used by the realtime publish path that needs a topic
// without the caller supplying the project id.
func (s *Service) GetColumnProject(ctx context.Context, columnID int64) int64 {
	return s.repo.GetColumnProject(ctx, columnID)
}

// GetProjectWorkspace returns the workspace id the given project belongs
// to. Used by realtime publish sites that route events to workspace-level
// subscribers without having to load the full project row.
func (s *Service) GetProjectWorkspace(ctx context.Context, projectID int64) int64 {
	return s.repo.GetProjectWorkspace(ctx, projectID)
}

// EnsureMyTasksProject returns the id of the caller's personal "My Tasks"
// project (and the workspace it lives in), creating it — plus its default
// kanban columns — on first call. This is the board the user lands on in Flow
// and where quick-created tasks go. Idempotent: it reuses the existing project,
// migrating the legacy "Inbox" name to "My Tasks" on access. Pass workspaceID 0
// to use (and lazily create) the caller's personal workspace.
func (s *Service) EnsureMyTasksProject(
	ctx context.Context,
	userID, workspaceID int64,
) (projectID, resolvedWorkspaceID int64, err error) {
	if workspaceID == 0 {
		wsID, wsErr := s.repo.CreatePersonalWorkspace(ctx, userID)
		if wsErr != nil {
			return 0, 0, wsErr
		}
		workspaceID = wsID
	}

	if existingID, existingName := s.repo.GetMyTasksProject(ctx, userID, workspaceID); existingID != 0 {
		if existingName == "Inbox" {
			if renameErr := s.repo.RenameProject(ctx, existingID, userID, workspaceID, "My Tasks"); renameErr != nil {
				logger.Ctx(ctx).Warn("migrate Inbox->My Tasks failed", logger.Error(renameErr))
			} else {
				invalidateWorkspacesScope(ctx, userID)
			}
		}
		return existingID, workspaceID, nil
	}

	id, err := s.repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: workspaceID,
		OwnerID:     userID,
		BusinessID:  userID,
		Name:        "My Tasks",
		Description: "Your personal task board. Quick-created tasks land here.",
		Icon:        "✅",
		Color:       "#64C5B9",
		CreatedBy:   userID,
	})
	if err != nil {
		return 0, 0, err
	}
	// Seed default kanban columns so the Board view works out of the box.
	if seedErr := s.repo.SeedDefaultColumns(ctx, id, workspaceID, userID); seedErr != nil {
		// Non-fatal — user can still create columns manually.
		_ = seedErr
	}
	invalidateWorkspacesScope(ctx, userID)
	return id, workspaceID, nil
}

func (s *Service) SeedDefaultColumns(ctx context.Context, projectID, workspaceID, createdBy int64) error {
	return s.repo.SeedDefaultColumns(ctx, projectID, workspaceID, createdBy)
}
