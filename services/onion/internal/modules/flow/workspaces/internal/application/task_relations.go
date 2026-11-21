package application

import (
	"context"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/wsbus"
)

// relationTask reads one end of a relation and refuses a task from another
// business. Role checks are not enough for that on their own: a business
// administrator passes RequireTaskRole before the task is ever read.
func (s *Service) relationTask(ctx context.Context, taskID, businessID int64) (domain.Task, error) {
	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil || businessID == 0 || task.BusinessID != businessID {
		return domain.Task{}, domain.ErrUnauthorized
	}
	return task, nil
}

// canSeeRelatedTask is CanReadTask's rule (project membership, or the right to
// open the task's project) applied to a task already read, so a listing can
// check each project once instead of reading every related task again.
func (s *Service) canSeeRelatedTask(ctx context.Context, task domain.Task, callerID, businessID int64) bool {
	if task.ProjectID == 0 {
		return s.IsTaskMember(ctx, callerID, task.ID)
	}
	return s.repo.IsProjectMember(ctx, callerID, task.ProjectID) ||
		s.opensProject(ctx, task.ProjectID, callerID, businessID)
}

// AddTaskRelation links taskID to otherTaskID. The caller's member role on
// taskID is checked by the API; here both tasks are confirmed to sit in the
// caller's business and the other task to be one the caller can see. The two
// may be in different projects.
func (s *Service) AddTaskRelation(ctx context.Context, taskID, otherTaskID, callerID, businessID int64, relationType, direction string) (domain.TaskRelationChange, error) {
	edge, err := domain.NewTaskRelationEdge(taskID, otherTaskID, relationType, direction)
	if err != nil {
		return domain.TaskRelationChange{}, err
	}
	task, err := s.relationTask(ctx, taskID, businessID)
	if err != nil {
		return domain.TaskRelationChange{}, err
	}
	other, err := s.relationTask(ctx, otherTaskID, businessID)
	if err != nil || !s.CanReadTask(ctx, callerID, businessID, other.ID) {
		return domain.TaskRelationChange{}, domain.ErrUnauthorized
	}

	if edge.LinkType == domain.RelationBlocks {
		// The reverse edge is the shortest cycle, so this covers it too.
		cycle, cycleErr := s.repo.HasBlockingPath(ctx, edge.DependsOnTaskID, edge.TaskID, businessID)
		if cycleErr != nil {
			return domain.TaskRelationChange{}, cycleErr
		}
		if cycle {
			return domain.TaskRelationChange{}, domain.ErrRelationCycle
		}
	} else {
		exists, existsErr := s.repo.TaskRelationExists(ctx, taskID, otherTaskID, edge.LinkType, businessID)
		if existsErr != nil {
			return domain.TaskRelationChange{}, existsErr
		}
		if exists {
			return domain.TaskRelationChange{}, domain.ErrRelationExists
		}
	}

	id, err := s.repo.AddTaskRelation(ctx, edge, businessID)
	if err != nil {
		return domain.TaskRelationChange{}, err
	}
	return domain.TaskRelationChange{
		RelationID: id,
		Tasks: [2]domain.TaskRef{
			{ID: task.ID, ProjectID: task.ProjectID},
			{ID: other.ID, ProjectID: other.ProjectID},
		},
	}, nil
}

// RemoveTaskRelation deletes a relation taskID is one end of. A relation of
// another task, or of another business, reports ErrRelationNotFound.
func (s *Service) RemoveTaskRelation(ctx context.Context, taskID, relationID, businessID int64) (domain.TaskRelationChange, error) {
	edge, err := s.repo.GetTaskRelation(ctx, relationID, businessID)
	if err != nil {
		return domain.TaskRelationChange{}, err
	}
	if edge.TaskID != taskID && edge.DependsOnTaskID != taskID {
		return domain.TaskRelationChange{}, domain.ErrRelationNotFound
	}
	if _, err := s.relationTask(ctx, taskID, businessID); err != nil {
		return domain.TaskRelationChange{}, err
	}
	if err := s.repo.DeleteTaskRelation(ctx, relationID, businessID); err != nil {
		return domain.TaskRelationChange{}, err
	}

	change := domain.TaskRelationChange{RelationID: relationID}
	for i, id := range []int64{taskID, edge.OtherTaskID(taskID)} {
		change.Tasks[i] = domain.TaskRef{ID: id}
		// The other end may have been deleted since; its event then goes to
		// the task topic alone.
		if end, readErr := s.repo.GetTaskByID(ctx, id); readErr == nil {
			change.Tasks[i].ProjectID = end.ProjectID
		}
	}
	return change, nil
}

// ListTaskRelations returns taskID's relations grouped by type and direction.
// A related task in a project the caller cannot open is left out, the same as
// it is from every task listing.
func (s *Service) ListTaskRelations(ctx context.Context, taskID, callerID, businessID int64) ([]domain.TaskRelationGroup, error) {
	if _, err := s.relationTask(ctx, taskID, businessID); err != nil {
		return nil, err
	}
	relations, err := s.repo.ListTaskRelations(ctx, taskID, businessID)
	if err != nil {
		return nil, err
	}
	visibleProjects := make(map[int64]bool)
	visible := relations[:0]
	for _, relation := range relations {
		projectID := relation.Task.ProjectID
		allowed, checked := visibleProjects[projectID]
		if !checked || projectID == 0 {
			allowed = s.canSeeRelatedTask(ctx,
				domain.Task{ID: relation.Task.ID, ProjectID: projectID}, callerID, businessID)
			visibleProjects[projectID] = allowed
		}
		if allowed {
			visible = append(visible, relation)
		}
	}
	return domain.GroupTaskRelations(visible), nil
}

// applyBlockedState fills Blocked and BlockerCount for a page of tasks with
// one query, whatever the page size.
func (s *Service) applyBlockedState(ctx context.Context, businessID int64, tasks []domain.Task) error {
	if len(tasks) == 0 || businessID == 0 {
		return nil
	}
	ids := make([]int64, 0, len(tasks))
	for i := range tasks {
		ids = append(ids, tasks[i].ID)
	}
	counts, err := s.repo.CountOpenBlockers(ctx, ids, businessID)
	if err != nil {
		return err
	}
	for i := range tasks {
		tasks[i].BlockerCount = counts[tasks[i].ID]
		tasks[i].Blocked = tasks[i].BlockerCount > 0
	}
	return nil
}

// watchUnblocks is called before a write that may close taskIDs. It notes
// which of them are open blockers right now and returns the func to call once
// the write has landed; that func announces every task the write unblocked.
//
// Reading the set first is what makes the announcement fire once: only a task
// that was open before and is closed after counts, so a second write that
// leaves it closed finds nothing to announce.
func (s *Service) watchUnblocks(ctx context.Context, businessID int64, taskIDs ...int64) func(actorID int64) {
	blockers, err := s.repo.ListOpenBlockerIDs(ctx, taskIDs, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("watchUnblocks: failed to read open blockers, no unblock notification will be sent",
			logger.Error(err))
		return func(int64) {}
	}
	if len(blockers) == 0 {
		return func(int64) {}
	}
	return func(actorID int64) { s.announceUnblocked(ctx, businessID, actorID, blockers) }
}

// announceUnblocked tells the people concerned about each task whose last open
// blocker was among blockerIDs and has now closed. blockerIDs must have been
// open before the write.
//
// The task's own topics get a task event so an open board drops the blocked
// badge. Its assignee gets an inbox row unless they closed the blocker
// themself. The row is a status_change notification but is not gated on that
// setting: it is off by default, and "you can start now" is the one status
// change an assignee is waiting for.
func (s *Service) announceUnblocked(ctx context.Context, businessID, actorID int64, blockerIDs []int64) {
	tasks, err := s.repo.ListUnblockedTasks(ctx, blockerIDs, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("announceUnblocked: failed to read unblocked tasks", logger.Error(err))
		return
	}
	for _, task := range tasks {
		publishTaskUpdated(task.ID, task.ProjectID, actorID)
		identifier := "T-" + strconv.FormatInt(task.ID, 10)
		s.notifyTaskRecipient(ctx, domain.Task{
			ID:           task.ID,
			OwnerID:      task.OwnerID,
			WorkspaceID:  task.WorkspaceID,
			Title:        task.Title,
			ProjectName:  task.ProjectName,
			ProjectColor: task.ProjectColor,
		}, task.AssigneeID, actorID,
			statusNotification, identifier+unblockedNotificationTitleSuffix, "",
			func(domain.UserSettings) bool { return true })
	}
}

const unblockedNotificationTitleSuffix = " is no longer blocked"

func publishTaskUpdated(taskID, projectID, actorID int64) {
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("task:%d", taskID),
		Type:     taskUpdatedEvent,
		EntityID: taskID,
		ActorID:  actorID,
	})
	if projectID == 0 {
		return
	}
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("project:%d", projectID),
		Type:     taskUpdatedEvent,
		EntityID: taskID,
		ActorID:  actorID,
	})
}
