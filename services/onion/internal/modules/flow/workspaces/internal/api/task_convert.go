package api

import (
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/wsbus"
	flowtaskv1 "github.com/voxire/voxie/gen/go/flow/task/v1"
)

func tasksToProto(tasks []domain.Task) []*flowtaskv1.Task {
	out := make([]*flowtaskv1.Task, 0, len(tasks))
	for _, t := range tasks {
		out = append(out, toTask(t))
	}
	return out
}

func toTask(t domain.Task) *flowtaskv1.Task {
	labels := t.Labels
	if labels == nil {
		labels = []int64{}
	}
	dependencies := t.Dependencies
	if dependencies == nil {
		dependencies = []int64{}
	}

	task := &flowtaskv1.Task{
		Id:               t.ID,
		ProjectId:        t.ProjectID,
		WorkspaceId:      t.WorkspaceID,
		OwnerId:          t.OwnerID,
		ParentTaskId:     t.ParentTaskID,
		Title:            t.Title,
		Description:      t.Description,
		Status:           t.Status,
		Priority:         t.Priority,
		AssigneeId:       t.AssigneeID,
		ReporterId:       t.ReporterID,
		EstimatedHours:   t.EstimatedHours,
		ActualHours:      t.ActualHours,
		SortOrder:        int64(t.SortOrder),
		Labels:           labels,
		CreatedAt:        t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        t.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Identifier:       t.Identifier,
		ProjectName:      t.ProjectName,
		ProjectColor:     t.ProjectColor,
		AssigneeName:     t.AssigneeName,
		AssigneeInitials: t.AssigneeInitials,
		SubtaskCount:     t.SubtaskCount,
		SectionId:        t.SectionID,
		CycleId:          t.CycleID,
		MilestoneId:      t.MilestoneID,
		Archived:         t.Archived,
		Dependencies:     dependencies,
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

func toComment(c domain.TaskComment) *flowtaskv1.TaskComment {
	return &flowtaskv1.TaskComment{
		Id:            c.ID,
		TaskId:        c.TaskID,
		UserId:        c.UserID,
		ActorName:     c.ActorName,
		ActorInitials: c.ActorInitials,
		Content:       c.Content,
		IsEdited:      c.IsEdited,
		CreatedAt:     c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:     c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toActivity(a domain.TaskActivity) *flowtaskv1.TaskActivity {
	return &flowtaskv1.TaskActivity{
		Id:            a.ID,
		TaskId:        a.TaskID,
		UserId:        a.UserID,
		ActorName:     a.ActorName,
		ActorInitials: a.ActorInitials,
		Action:        a.Action,
		OldValue:      a.OldValue,
		NewValue:      a.NewValue,
		CreatedAt:     a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toAttachment(a domain.TaskAttachment) *flowtaskv1.Attachment {
	return &flowtaskv1.Attachment{
		Id:         a.ID,
		TaskId:     a.TaskID,
		Name:       a.Name,
		Url:        a.URL,
		Size:       a.Size,
		UploadedBy: a.UploadedBy,
		CreatedAt:  a.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func toTimeLog(t domain.TimeLog) *flowtaskv1.TimeLog {
	return &flowtaskv1.TimeLog{
		Id:              t.ID,
		TaskId:          t.TaskID,
		UserId:          t.UserID,
		DurationMinutes: int64(t.DurationMinutes),
		Description:     t.Description,
		LoggedAt:        t.LoggedAt.Format("2006-01-02T15:04:05Z"),
		CreatedAt:       t.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

func publishTaskEvent(eventType string, task *flowtaskv1.Task, actorID int64) {
	if task == nil {
		return
	}
	publishTaskEventByProject(eventType, task.ProjectId, task.Id, actorID)
}

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

func subtasksToCompact(tasks []domain.Task) []*flowtaskv1.TaskPayloadCompact {
	out := make([]*flowtaskv1.TaskPayloadCompact, 0, len(tasks))
	for _, t := range tasks {
		compact := &flowtaskv1.TaskPayloadCompact{
			Id:        t.ID,
			Title:     t.Title,
			Status:    t.Status,
			Priority:  t.Priority,
			CreatedAt: t.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
		if t.DueDate != nil {
			compact.DueDate = t.DueDate.Format("2006-01-02")
		}
		out = append(out, compact)
	}
	return out
}

func parseTaskDate(raw string) (*time.Time, error) {
	if raw == "" || raw == "-" {
		return nil, nil
	}
	if t, err := time.Parse("2006-01-02T15:04:05Z", raw); err == nil {
		return &t, nil
	}
	t, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
