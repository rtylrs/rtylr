package application

import (
	"context"
	"fmt"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
	"github.com/voxire/onion/internal/platform/logger"
)

func (s *Service) DeleteTask(ctx context.Context, taskID, businessID int64) error {
	defer invalidateWorkspacesScope(ctx, businessID)

	// Read the assignee before the row goes. Once the task is deleted there is
	// nothing left to say whose cached my-tasks pages still list it.
	existing, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTask: failed to read task before delete, my-tasks cache left to expire",
			logger.Error(err))
	} else {
		defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))
	}

	return s.repo.DeleteTask(ctx, taskID, businessID)
}

type myTasksPage struct {
	Tasks []domain.Task
	Total int64
}

func (s *Service) GetMyTasks(ctx context.Context, userID, page, limit int64) ([]domain.Task, int64, error) {
	return s.getMyTasks(ctx, userID, 0, page, limit)
}

// GetMyTasksForBusiness is the v2/Today read. Its cache key includes the
// verified business as well as the user, page, and limit; otherwise switching
// businesses could serve the previous tenant's rows until the two-minute TTL.
func (s *Service) GetMyTasksForBusiness(ctx context.Context, userID, businessID, page, limit int64) ([]domain.Task, int64, error) {
	if businessID <= 0 {
		return nil, 0, domain.ErrUnauthorized
	}
	return s.getMyTasks(ctx, userID, businessID, page, limit)
}

func (s *Service) getMyTasks(ctx context.Context, userID, businessID, page, limit int64) ([]domain.Task, int64, error) {
	offset := (page - 1) * limit
	keyDetail := fmt.Sprintf("workspace_tasks:page=%d:limit=%d", page, limit)
	if businessID != 0 {
		keyDetail = fmt.Sprintf("business_tasks:business=%d:%s", businessID, keyDetail)
	}
	result, err := cacher.Of[myTasksPage]().GetOrComputeSingle(ctx,
		cacheKeys.Scoped("user", userID, keyDetail),
		func(ctx context.Context) (myTasksPage, error) {
			// workspaceID=0: this is the cross-workspace "My tasks" board, which
			// intentionally spans every workspace the user belongs to within the
			// selected business. The zero-business path is retained for legacy
			// workspace-era callers; v2 always supplies a verified business id.
			var tasks []domain.Task
			var total int64
			var fetchErr error
			if businessID != 0 {
				tasks, total, fetchErr = s.repo.GetMyTasksForBusiness(ctx, userID, businessID, 0, limit, offset)
			} else {
				tasks, total, fetchErr = s.repo.GetMyTasks(ctx, userID, 0, limit, offset)
			}
			if fetchErr != nil {
				return myTasksPage{}, fetchErr
			}
			if businessID != 0 {
				if pinErr := s.applyFocusPins(ctx, userID, businessID, tasks); pinErr != nil {
					return myTasksPage{}, pinErr
				}
			}
			return myTasksPage{Tasks: tasks, Total: total}, nil
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
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return domain.TaskComment{}, err
	}

	id, err := s.repo.CreateComment(ctx, taskID, userID, content)
	if err != nil {
		return domain.TaskComment{}, err
	}

	if actErr := s.repo.LogActivity(ctx, taskID, userID, "comment_added", "", content); actErr != nil {
		logger.Ctx(ctx).Error("AddComment: failed to log activity", logger.Error(actErr))
	}

	comment := domain.TaskComment{
		ID:      id,
		TaskID:  taskID,
		UserID:  userID,
		Content: content,
	}
	if stored, readErr := s.repo.GetCommentByID(ctx, id); readErr != nil {
		logger.Ctx(ctx).Error("AddComment: failed to read stored comment", logger.Error(readErr))
	} else {
		comment = stored
	}

	s.announceCommentNotifications(ctx, task, userID, content)
	return comment, nil
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

	// Scope "my tasks" to this workspace so the dashboard never surfaces work
	// assigned in a different workspace (issue #783).
	myTasks, _, err := s.repo.GetMyTasks(ctx, userID, workspaceID, internalListCap, 0)
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

func (s *Service) UpdateLabel(ctx context.Context, id, actorID, businessID int64, input domain.LabelInput) error {
	return s.repo.UpdateLabel(ctx, id, actorID, businessID, input)
}
