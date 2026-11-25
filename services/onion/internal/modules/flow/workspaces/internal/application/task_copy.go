package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// DuplicateTask copies a task into the same project (T-343): description,
// priority, assignee, dates, estimates, labels, type, epic, milestone, custom
// field values and, for a top-level task, its subtasks. The copy starts in the
// project's default status, outside any cycle, and is linked to the original
// as a clone. Comments, attachments and time logs stay with the original.
// keepAssignee is false when the caller may not assign work to others; the
// copy is then unassigned.
func (s *Service) DuplicateTask(ctx context.Context, businessID, callerID, taskID int64, title string, keepAssignee bool) (domain.Task, error) {
	detail, err := s.GetTaskDetail(ctx, taskID)
	if err != nil {
		return domain.Task{}, err
	}
	if detail.Task.BusinessID != businessID {
		return domain.Task{}, domain.ErrUnauthorized
	}
	sources := []domain.Task{detail.Task}
	if err := s.AttachTaskMemberships(ctx, sources); err != nil {
		return domain.Task{}, err
	}
	if err := s.attachFieldValues(ctx, businessID, sources); err != nil {
		return domain.Task{}, err
	}
	src := sources[0]

	parentID := src.ParentTaskID
	if parentID == nil && src.TaskType != domain.TaskTypeEpic {
		parentID = src.EpicID
	}
	if title == "" {
		title = src.Title
	}
	assignee := src.AssigneeID
	if !keepAssignee {
		assignee = nil
	}
	fieldValues := make([]domain.FieldValueInput, 0, len(src.FieldValues))
	for fieldID, value := range src.FieldValues {
		fieldValues = append(fieldValues, domain.FieldValueInput{FieldID: fieldID, Value: value})
	}
	copied, err := s.CreateTask(ctx, domain.TaskInput{
		ProjectID:      src.ProjectID,
		WorkspaceID:    src.WorkspaceID,
		OwnerID:        callerID,
		BusinessID:     businessID,
		ParentTaskID:   parentID,
		Title:          title,
		Description:    src.Description,
		EffortMinutes:  src.EffortMinutes,
		Priority:       src.Priority,
		AssigneeID:     assignee,
		ReporterID:     callerID,
		DueDate:        src.DueDate,
		StartDate:      src.StartDate,
		EstimatedHours: src.EstimatedHours,
		Labels:         src.Labels,
		FieldValues:    fieldValues,
		TaskType:       src.TaskType,
		EpicColor:      src.EpicColor,
	})
	if err != nil {
		return domain.Task{}, err
	}

	// The copy exists from here on, so a failure below is logged rather than
	// returned: a retry would create a second copy.
	if src.MilestoneID != nil {
		if err := s.SetTaskMilestone(ctx, copied.ID, src.MilestoneID); err != nil {
			logger.Ctx(ctx).Error("DuplicateTask milestone failed", logger.Error(err))
		}
	}
	if src.ParentTaskID == nil {
		for _, sub := range detail.Subtasks {
			parent := copied.ID
			if _, err := s.CreateTask(ctx, domain.TaskInput{
				ProjectID: src.ProjectID, WorkspaceID: src.WorkspaceID, OwnerID: callerID, BusinessID: businessID,
				ParentTaskID: &parent, Title: sub.Title, Description: sub.Description, Priority: sub.Priority,
				AssigneeID: subAssignee(sub.AssigneeID, keepAssignee), ReporterID: callerID, DueDate: sub.DueDate, StartDate: sub.StartDate,
				EffortMinutes: sub.EffortMinutes, EstimatedHours: sub.EstimatedHours, Labels: sub.Labels,
			}); err != nil {
				logger.Ctx(ctx).Error("DuplicateTask subtask failed", logger.Error(err))
			}
		}
	}
	if _, err := s.AddTaskRelation(ctx, copied.ID, src.ID, callerID, businessID, domain.RelationClones, domain.RelationOutward); err != nil {
		logger.Ctx(ctx).Error("DuplicateTask clone link failed", logger.Error(err))
	}

	fresh, err := s.GetTaskDetail(ctx, copied.ID)
	if err != nil {
		return copied, nil
	}
	out := []domain.Task{fresh.Task}
	if err := s.AttachTaskMemberships(ctx, out); err != nil {
		return fresh.Task, nil
	}
	return out[0], nil
}

// ImportTaskRow is one task of an import, already validated by the caller,
// with the titles of its subtasks.
type ImportTaskRow struct {
	Input    domain.TaskInput
	Subtasks []string
}

// CheckImportProject reports a project setting that would refuse every row,
// such as a required custom field with no default, before anything is written.
func (s *Service) CheckImportProject(ctx context.Context, businessID, projectID int64) error {
	_, err := s.fieldValuesForNewTask(ctx, businessID, projectID, nil)
	return err
}

// ImportTasks creates the rows in order, each with its subtasks, and returns
// the new task ids. Rows are validated before this is called, so an error here
// is a fault; the ids created so far are returned with it.
func (s *Service) ImportTasks(ctx context.Context, rows []ImportTaskRow) ([]int64, error) {
	ids := make([]int64, 0, len(rows))
	for _, row := range rows {
		task, err := s.CreateTask(ctx, row.Input)
		if err != nil {
			return ids, err
		}
		ids = append(ids, task.ID)
		for _, title := range row.Subtasks {
			parent := task.ID
			if _, err := s.CreateTask(ctx, domain.TaskInput{
				ProjectID: row.Input.ProjectID, OwnerID: row.Input.OwnerID, BusinessID: row.Input.BusinessID,
				ParentTaskID: &parent, Title: title, ReporterID: row.Input.ReporterID,
			}); err != nil {
				return ids, err
			}
		}
	}
	return ids, nil
}

func subAssignee(id *int64, keep bool) *int64 {
	if keep {
		return id
	}
	return nil
}
