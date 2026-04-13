package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/logger"
)

type Service struct {
	repo *infrastructure.Repository
}

func NewService(ctx context.Context) *Service {
	repo := infrastructure.NewRepository(ctx)

	return &Service{
		repo: repo,
	}
}

// ─────────────────────────────────────────────
// Workspaces
// ─────────────────────────────────────────────

func (s *Service) ListWorkspaces(ctx context.Context, ownerID int64) ([]domain.Workspace, error) {
	return s.repo.GetWorkspacesByOwner(ctx, ownerID)
}

func (s *Service) CreateWorkspace(ctx context.Context, input domain.WorkspaceInput) (domain.Workspace, error) {
	count, err := s.repo.CountWorkspacesByOwner(ctx, input.OwnerID)
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
	return s.repo.UpdateWorkspace(ctx, input)
}

func (s *Service) DeleteWorkspace(ctx context.Context, id, ownerID int64) error {
	return s.repo.DeleteWorkspace(ctx, id, ownerID)
}

func (s *Service) AddWorkspaceMember(ctx context.Context, input domain.AddWorkspaceMemberInput) error {
	return s.repo.AddWorkspaceMember(ctx, input)
}

func (s *Service) RemoveWorkspaceMember(ctx context.Context, input domain.RemoveWorkspaceMemberInput) error {
	return s.repo.RemoveWorkspaceMember(ctx, input)
}

func (s *Service) ListWorkspaceMembers(ctx context.Context, workspaceID int64) ([]domain.WorkspaceMember, error) {
	return s.repo.ListWorkspaceMembers(ctx, workspaceID)
}

func (s *Service) UpdateWorkspaceMemberRole(ctx context.Context, input domain.UpdateWorkspaceMemberRoleInput) error {
	return s.repo.UpdateWorkspaceMemberRole(ctx, input)
}

func (s *Service) CreatePersonalWorkspace(ctx context.Context, userID int64) (int64, error) {
	return s.repo.CreatePersonalWorkspace(ctx, userID)
}

// ─────────────────────────────────────────────
// Projects
// ─────────────────────────────────────────────

func (s *Service) ListProjects(ctx context.Context, workspaceID, ownerID int64) ([]domain.Project, error) {
	return s.repo.GetProjectsByWorkspace(ctx, workspaceID, ownerID)
}

func (s *Service) CreateProject(ctx context.Context, input domain.ProjectInput) (domain.Project, error) {
	id, err := s.repo.CreateProject(ctx, input)
	if err != nil {
		return domain.Project{}, err
	}

	return domain.Project{
		ID:          id,
		WorkspaceID: input.WorkspaceID,
		OwnerID:     input.OwnerID,
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
	return s.repo.UpdateProject(ctx, input)
}

func (s *Service) DeleteProject(ctx context.Context, id, ownerID int64) error {
	return s.repo.ArchiveProject(ctx, id, ownerID)
}

func (s *Service) ListProjectMembers(ctx context.Context, projectID int64) ([]domain.ProjectMember, error) {
	return s.repo.ListProjectMembers(ctx, projectID)
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

// ─────────────────────────────────────────────
// Tasks
// ─────────────────────────────────────────────

func (s *Service) ListTasks(ctx context.Context, filter domain.ListTasksFilter) ([]domain.Task, error) {
	return s.repo.GetTasksByProject(ctx, filter)
}

func (s *Service) CreateTask(ctx context.Context, input domain.TaskInput) (domain.Task, error) {
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

	subtasks, err := s.repo.GetSubtasks(ctx, taskID)
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
	existing, err := s.repo.GetTaskByID(ctx, input.ID)
	if err != nil {
		return err
	}

	if existing.Status != input.Status {
		if actErr := s.repo.LogActivity(ctx, input.ID, userID, "status_changed", existing.Status, input.Status); actErr != nil {
			logger.Ctx(ctx).Error("UpdateTask: failed to log status activity", logger.Error(actErr))
		}
	}

	if existing.Priority != input.Priority {
		if actErr := s.repo.LogActivity(ctx, input.ID, userID, "priority_changed", existing.Priority, input.Priority); actErr != nil {
			logger.Ctx(ctx).Error("UpdateTask: failed to log priority activity", logger.Error(actErr))
		}
	}

	existingAssignee := int64(0)
	if existing.AssigneeID != nil {
		existingAssignee = *existing.AssigneeID
	}

	newAssignee := int64(0)
	if input.AssigneeID != nil {
		newAssignee = *input.AssigneeID
	}

	if existingAssignee != newAssignee {
		if actErr := s.repo.LogActivity(ctx, input.ID, userID, "assignee_changed",
			fmt.Sprintf("%d", existingAssignee),
			fmt.Sprintf("%d", newAssignee),
		); actErr != nil {
			logger.Ctx(ctx).Error("UpdateTask: failed to log assignee activity", logger.Error(actErr))
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

func (s *Service) DeleteTask(ctx context.Context, taskID, ownerID int64) error {
	return s.repo.DeleteTask(ctx, taskID, ownerID)
}

func (s *Service) GetMyTasks(ctx context.Context, userID int64) ([]domain.Task, error) {
	return s.repo.GetMyTasks(ctx, userID)
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

// ─────────────────────────────────────────────
// Dashboard
// ─────────────────────────────────────────────

func (s *Service) GetDashboard(ctx context.Context, workspaceID, ownerID, userID int64) (domain.DashboardStats, error) {
	stats, err := s.repo.GetDashboardStats(ctx, workspaceID, ownerID)
	if err != nil {
		return domain.DashboardStats{}, err
	}

	myTasks, err := s.repo.GetMyTasks(ctx, userID)
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

func (s *Service) GetSubtasks(ctx context.Context, taskID int64) ([]domain.Task, error) {
	return s.repo.GetSubtasks(ctx, taskID)
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

func (s *Service) DeleteAttachment(ctx context.Context, attachmentID, ownerID int64) error {
	return s.repo.DeleteAttachment(ctx, attachmentID, ownerID)
}

// ─────────────────────────────────────────────
// Time Logs
// ─────────────────────────────────────────────

func (s *Service) GetTimeLogs(ctx context.Context, taskID int64) ([]domain.TimeLog, error) {
	return s.repo.GetTimeLogsByTask(ctx, taskID)
}

func (s *Service) CreateTimeLog(ctx context.Context, input domain.TimeLogInput) (int64, error) {
	return s.repo.CreateTimeLog(ctx, input)
}

// ─────────────────────────────────────────────
// Labels
// ─────────────────────────────────────────────

func (s *Service) ListLabels(ctx context.Context, workspaceID int64) ([]domain.Label, error) {
	return s.repo.GetLabelsByWorkspace(ctx, workspaceID)
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
