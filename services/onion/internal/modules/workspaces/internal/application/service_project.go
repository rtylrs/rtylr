package application

import (
	"context"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

func (s *Service) UpdateProject(ctx context.Context, input domain.UpdateProjectInput) error {
	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	return s.repo.UpdateProject(ctx, input)
}

func (s *Service) DeleteProject(ctx context.Context, id, businessID int64) error {
	defer invalidateWorkspacesScope(ctx, businessID)
	// Archive now carries recycle-bin semantics: stamp the permanent-deletion
	// deadline from the retention setting so every client sees the same clock.
	deadline := binPurgeDeadline(time.Now().UTC())
	return s.repo.ArchiveProject(ctx, id, businessID, &deadline)
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
	if input.Status == "" {
		input.Status = domain.TaskStatusTodo
	}
	if input.Priority == "" {
		input.Priority = "none"
	}

	// R13: an omitted project means "file into my My Tasks bucket", which is
	// lazily provisioned on first use. Every task belongs to exactly one
	// project; "outside a project" is a UI concept.
	if input.ProjectID == 0 {
		bucketID, _, ensureErr := s.EnsureMyTasksProject(ctx, input.OwnerID, input.BusinessID, 0)
		if ensureErr != nil {
			return domain.Task{}, ensureErr
		}
		input.ProjectID = bucketID
	}

	// T1: assignment carries access with it.
	if input.AssigneeID != nil && *input.AssigneeID > 0 && *input.AssigneeID != input.OwnerID {
		if grantErr := s.EnsureAssigneeMembership(ctx, input.ProjectID, input.BusinessID, *input.AssigneeID); grantErr != nil {
			return domain.Task{}, grantErr
		}
	}

	// A new task lands in the assignee's my-tasks list, whose cached pages were
	// computed before it existed.
	defer invalidateUserScope(ctx, assigneeID(input.AssigneeID))
	id, err := s.repo.CreateTask(ctx, input)
	if err != nil {
		return domain.Task{}, err
	}

	if actErr := s.repo.LogActivity(ctx, id, input.ReporterID, "created_task", "", input.Title); actErr != nil {
		logger.Ctx(ctx).Error("CreateTask: failed to log activity", logger.Error(actErr))
	}

	return s.repo.GetTaskByID(ctx, id)
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

	dependencyMap, err := s.repo.ListTaskDependencies(ctx, []int64{taskID}, task.BusinessID)
	if err != nil {
		return domain.TaskDetail{}, err
	}
	task.Dependencies = dependencyMap[taskID]
	if task.Dependencies == nil {
		task.Dependencies = []int64{}
	}

	return domain.TaskDetail{
		Task:     task,
		Subtasks: subtasks,
		Comments: comments,
		Activity: activity,
	}, nil
}

func (s *Service) applyFocusPins(ctx context.Context, userID, businessID int64, tasks []domain.Task) error {
	if len(tasks) == 0 {
		return nil
	}
	ids := make([]int64, 0, len(tasks))
	for _, task := range tasks {
		ids = append(ids, task.ID)
	}
	pins, err := s.repo.GetFocusPins(ctx, userID, businessID, ids)
	if err != nil {
		return err
	}
	for i := range tasks {
		tasks[i].FocusPin = pins[tasks[i].ID]
	}
	return nil
}

// GetTaskDetailForViewer adds the per-user Today projection to a detail read
// and refuses a task from another business before exposing its pin state.
func (s *Service) GetTaskDetailForViewer(ctx context.Context, taskID, userID, businessID int64) (domain.TaskDetail, error) {
	detail, err := s.GetTaskDetail(ctx, taskID)
	if err != nil {
		return domain.TaskDetail{}, err
	}
	if businessID == 0 || detail.Task.BusinessID != businessID {
		return domain.TaskDetail{}, domain.ErrUnauthorized
	}
	for _, subtask := range detail.Subtasks {
		if subtask.BusinessID != businessID {
			return domain.TaskDetail{}, domain.ErrUnauthorized
		}
	}
	tasks := make([]domain.Task, 0, 1+len(detail.Subtasks))
	tasks = append(tasks, detail.Task)
	tasks = append(tasks, detail.Subtasks...)
	if err := s.applyFocusPins(ctx, userID, businessID, tasks); err != nil {
		return domain.TaskDetail{}, err
	}
	detail.Task = tasks[0]
	detail.Subtasks = tasks[1:]
	return detail, nil
}

func (s *Service) UpdateTask(ctx context.Context, input domain.UpdateTaskInput, userID int64) error {
	// Registered ahead of the cache evictions below so that it runs after all
	// of them: announcing tells the affected clients to refetch, and that
	// refetch must not be able to find a pre-update my-tasks page still cached.
	// Stays nil unless the update both changed the assignee and landed.
	var announced *assigneeChange
	defer func() {
		if announced != nil {
			s.announceAssigneeChange(ctx, *announced)
		}
	}()

	defer invalidateWorkspacesScope(ctx, input.BusinessID)
	existing, err := s.repo.GetTaskByID(ctx, input.ID)
	if err != nil {
		return err
	}
	if input.BusinessID == 0 || existing.BusinessID != input.BusinessID {
		return domain.ErrUnauthorized
	}

	// Any edit changes what the current assignee's my-tasks list should show,
	// title and status included, so their cached pages go regardless of whether
	// this update touches the assignee.
	defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))

	// This is a partial update — only record a change for a field the request
	// actually supplied. An omitted field arrives as the zero value (empty
	// string / nil) and means "leave unchanged", so comparing it against the
	// existing value would otherwise log a spurious "changed to empty" entry.
	//
	// The entries are computed here, from existing versus input, and written
	// only once the update has landed. Each LogActivity is its own write with
	// no transaction around it, so writing them first left false history behind
	// every time the update then failed: the task stayed as it was and its
	// Activity tab claimed changes that never happened.
	var activity []taskActivityEntry

	if input.Status != "" && existing.Status != input.Status {
		activity = append(activity, taskActivityEntry{"status_changed", existing.Status, input.Status})
	}

	if input.Priority != "" && existing.Priority != input.Priority {
		activity = append(activity, taskActivityEntry{"priority_changed", existing.Priority, input.Priority})
	}

	var reassignment *assigneeChange
	if input.AssigneeID != nil {
		if *input.AssigneeID > 0 && *input.AssigneeID != userID {
			if err := s.EnsureAssigneeMembership(ctx, existing.ProjectID, input.BusinessID, *input.AssigneeID); err != nil {
				return err
			}
		}
		existingAssignee := int64(0)
		if existing.AssigneeID != nil {
			existingAssignee = *existing.AssigneeID
		}
		newAssignee := *input.AssigneeID

		if existingAssignee != newAssignee {
			activity = append(activity, taskActivityEntry{
				"assignee_changed",
				strconv.FormatInt(existingAssignee, 10),
				strconv.FormatInt(newAssignee, 10),
			})
			// Both sides of a reassignment need evicting. The previous
			// assignee's eviction is already deferred above; without this one
			// the new assignee keeps serving pages that predate the task
			// arriving in their list.
			defer invalidateUserScope(ctx, newAssignee)

			// The title comes from the request when the same call renames the
			// task, so the inbox row names what the assignee will actually see.
			title := existing.Title
			if input.Title != "" {
				title = input.Title
			}
			reassignment = &assigneeChange{
				taskID:       input.ID,
				workspaceID:  existing.WorkspaceID,
				taskTitle:    title,
				project:      existing.ProjectName,
				projectColor: existing.ProjectColor,
				actorID:      userID,
				previous:     existingAssignee,
				next:         newAssignee,
			}
		}
	}

	if updateErr := s.repo.UpdateTask(ctx, input); updateErr != nil {
		return updateErr
	}
	if input.FocusPin != nil {
		if pinErr := s.repo.SetTaskFocusPin(ctx, input.ID, userID, input.BusinessID, *input.FocusPin); pinErr != nil {
			return pinErr
		}
		defer invalidateUserScope(ctx, userID)
	}

	// The only place UpdateTask writes activity, and it is past the point of
	// failure. A logging failure is reported and swallowed as before: the edit
	// itself succeeded, and losing a history row is not a reason to tell the
	// caller their change did not land.
	for _, entry := range activity {
		if actErr := s.repo.LogActivity(ctx, input.ID, userID, entry.kind, entry.from, entry.to); actErr != nil {
			logger.Ctx(ctx).Error("UpdateTask: failed to log activity",
				logger.Any("kind", entry.kind), logger.Error(actErr))
		}
	}

	if input.Status != "" && existing.Status != input.Status {
		notificationTask := existing
		notificationTask.Status = input.Status
		if input.Title != "" {
			notificationTask.Title = input.Title
		}
		if input.AssigneeID != nil {
			notificationTask.AssigneeID = input.AssigneeID
		}
		s.announceStatusChange(ctx, notificationTask, userID, input.Status)
	}

	announced = reassignment
	return nil
}
