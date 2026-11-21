package application

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
)

// resolveNewTaskHierarchy settles a new task's type and where it attaches.
// input.ParentTaskID may name an epic or a task; on return it names a task
// only, with an epic moved to input.EpicID. input.ProjectID must already be
// final, because the parent has to share it.
func (s *Service) resolveNewTaskHierarchy(ctx context.Context, input *domain.TaskInput) error {
	if input.TaskType == "" {
		input.TaskType = domain.TaskTypeTask
	}
	if !domain.IsValidTaskType(input.TaskType) {
		return domain.ErrInvalidTaskType
	}
	color, err := domain.NormalizeEpicColor(input.EpicColor)
	if err != nil {
		return err
	}
	if color != "" && input.TaskType != domain.TaskTypeEpic {
		return domain.ErrEpicColorOnNonEpic
	}
	input.EpicColor = color
	// The epic is only ever derived from the parent, never taken as given.
	input.EpicID = nil

	if input.ParentTaskID == nil || *input.ParentTaskID == 0 {
		input.ParentTaskID = nil
		return nil
	}
	parent, found, err := s.repo.GetTaskHierarchyNode(ctx, *input.ParentTaskID, input.BusinessID)
	if err != nil {
		return err
	}
	if !found {
		return domain.ErrTaskParentNotFound
	}
	placement, err := domain.PlaceUnderParent(domain.HierarchyNode{
		ProjectID:  input.ProjectID,
		BusinessID: input.BusinessID,
		TaskType:   input.TaskType,
	}, parent)
	if err != nil {
		return err
	}
	input.ParentTaskID, input.EpicID = placement.ParentTaskID, placement.EpicID
	return nil
}

// resolveTaskTypeUpdate validates a type or epic colour change against the
// task as stored, and normalises the input so the repository writes only what
// actually changes.
func (s *Service) resolveTaskTypeUpdate(ctx context.Context, existing domain.Task, input *domain.UpdateTaskInput) error {
	resultingType := existing.TaskType
	if input.TaskType != "" {
		if !domain.IsValidTaskType(input.TaskType) {
			return domain.ErrInvalidTaskType
		}
		if input.TaskType == existing.TaskType {
			input.TaskType = ""
		} else {
			node, found, err := s.repo.GetTaskHierarchyNode(ctx, existing.ID, input.BusinessID)
			if err != nil {
				return err
			}
			if !found {
				return domain.ErrUnauthorized
			}
			if err := domain.ValidateTaskTypeChange(node, input.TaskType); err != nil {
				return err
			}
			resultingType = input.TaskType
		}
	}
	if input.EpicColor != nil {
		color, err := domain.NormalizeEpicColor(*input.EpicColor)
		if err != nil {
			return err
		}
		if color != "" && resultingType != domain.TaskTypeEpic {
			return domain.ErrEpicColorOnNonEpic
		}
		input.EpicColor = &color
	}
	return nil
}

// SetTaskParent moves a task under an epic or a task, or to the top level when
// parentID is nil or zero, and returns the task as it now stands.
//
// A parent in another business is reported as not found, the same answer as a
// parent that does not exist, so the call cannot be used to probe for ids.
// actorID only selects whose Today pin the returned task carries.
func (s *Service) SetTaskParent(ctx context.Context, taskID, actorID, businessID int64, parentID *int64) (domain.Task, error) {
	if businessID <= 0 {
		return domain.Task{}, domain.ErrUnauthorized
	}
	if parentID != nil && *parentID == 0 {
		parentID = nil
	}
	defer invalidateWorkspacesScope(ctx, businessID)

	err := s.repo.SetTaskParent(ctx, taskID, businessID, parentID,
		func(child domain.HierarchyNode, parent *domain.HierarchyNode) (domain.TaskPlacement, error) {
			if parentID == nil {
				return domain.TaskPlacement{}, nil
			}
			if parent == nil {
				return domain.TaskPlacement{}, domain.ErrTaskParentNotFound
			}
			return domain.PlaceUnderParent(child, *parent)
		})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Task{}, domain.ErrUnauthorized
	}
	if err != nil {
		return domain.Task{}, err
	}

	task, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	// Moving between the board's top level and a parent's checklist changes
	// what the assignee's my-tasks pages should list.
	invalidateUserScope(ctx, assigneeID(task.AssigneeID))

	tasks := []domain.Task{task}
	if err := s.applyFocusPins(ctx, actorID, businessID, tasks); err != nil {
		return domain.Task{}, err
	}
	return tasks[0], nil
}

// DeleteTaskDetachingEpicTasks deletes a task and reports how many tasks were
// detached from it. Only an epic has any: its tasks are kept, with no epic.
func (s *Service) DeleteTaskDetachingEpicTasks(ctx context.Context, taskID, businessID int64) (int64, error) {
	defer invalidateWorkspacesScope(ctx, businessID)

	// Read the assignee before the row goes, as DeleteTask does.
	if existing, err := s.repo.GetTaskByID(ctx, taskID); err == nil {
		defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))
	}
	return s.repo.DeleteTaskDetachingEpicTasks(ctx, taskID, businessID)
}

// ListEpicRollups summarises the epics of one project. The caller must already
// be allowed to read the project; the business scope is applied here.
func (s *Service) ListEpicRollups(ctx context.Context, projectID, businessID int64) ([]domain.EpicRollup, error) {
	if businessID <= 0 {
		return nil, domain.ErrUnauthorized
	}
	return s.repo.ListEpicRollups(ctx, projectID, businessID)
}
