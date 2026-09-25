package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/wsbus"
)

// taskActivityEntry is one history row an update will produce, held between
// computing it from the request and writing it after the update succeeds.
type taskActivityEntry struct {
	kind string
	from string
	to   string
}

// assigneeChange is what a reassignment needs in order to announce itself once
// the write has landed: who lost the task, who picked it up, and enough of the
// task to render an inbox row without reading it back.
type assigneeChange struct {
	taskID       int64
	workspaceID  int64
	taskTitle    string
	project      string
	projectColor string
	actorID      int64
	previous     int64
	next         int64
}

// announceAssigneeChange is the realtime half of a reassignment.
//
// Both sides get a task event on their own `user:<id>` topic. The previous
// assignee's my-tasks list has to drop the task and the new assignee's has to
// pick it up, and neither of them is subscribed to the `task:<id>` or
// `project:<id>` topics the write already publishes on. A client sitting on My
// tasks or Today subscribes to `workspace:<id>` and `user:<id>` only, so a
// reassignment reached it through no topic at all.
//
// The new assignee also gets an inbox row, which is what makes "you were
// assigned a task" visible outside the flow board. That row is skipped when
// they turned task-assignment notifications off, and when they assigned the
// task to themselves and already know. The realtime event is deliberately not
// gated on either: it is an invalidation hint that keeps a list correct rather
// than a notification, and muting an inbox should not leave My tasks stale.
func (s *Service) announceAssigneeChange(ctx context.Context, change assigneeChange) {
	publishUserTaskEvent(change.previous, change.taskID, change.actorID)
	publishUserTaskEvent(change.next, change.taskID, change.actorID)

	if change.next == 0 || change.next == change.actorID {
		return
	}

	settings, err := s.repo.GetUserSettings(ctx, change.next)
	if err != nil {
		logger.Ctx(ctx).Error("announceAssigneeChange: failed to read notification settings, no inbox row created",
			logger.Error(err))
		return
	}
	if !settings.TaskAssigned {
		return
	}

	notification := domain.CreateNotificationInput{
		UserID:       change.next,
		Type:         assignedNotification,
		Title:        assignedNotificationTitle,
		TaskID:       &change.taskID,
		TaskTitle:    change.taskTitle,
		ActorID:      &change.actorID,
		Project:      change.project,
		ProjectColor: change.projectColor,
	}
	if change.workspaceID != 0 {
		notification.WorkspaceID = &change.workspaceID
	}

	// CreateNotification publishes `notification.created` to the same user
	// topic, so the inbox badge updates off the same call.
	if notifyErr := s.CreateNotification(ctx, notification); notifyErr != nil {
		logger.Ctx(ctx).Error("announceAssigneeChange: failed to create the assignment notification",
			logger.Error(notifyErr))
	}
}

func (s *Service) notifyTaskRecipient(
	ctx context.Context,
	task domain.Task,
	recipientID, actorID int64,
	notificationType, title, description string,
	settingEnabled func(domain.UserSettings) bool,
) {
	if recipientID == 0 || recipientID == actorID {
		return
	}

	settings, err := s.repo.GetUserSettings(ctx, recipientID)
	if err != nil {
		logger.Ctx(ctx).Error("notifyTaskRecipient: failed to read notification settings",
			logger.Int("owner_id", int(task.OwnerID)),
			logger.Int("task_id", int(task.ID)),
			logger.Int("recipient_id", int(recipientID)),
			logger.Error(err))
		return
	}
	if !settingEnabled(settings) {
		return
	}

	taskID := task.ID
	actor := actorID
	notification := domain.CreateNotificationInput{
		UserID:       recipientID,
		Type:         notificationType,
		Title:        title,
		Description:  description,
		TaskID:       &taskID,
		TaskTitle:    task.Title,
		ActorID:      &actor,
		Project:      task.ProjectName,
		ProjectColor: task.ProjectColor,
	}
	if task.WorkspaceID != 0 {
		workspaceID := task.WorkspaceID
		notification.WorkspaceID = &workspaceID
	}

	if err := s.CreateNotification(ctx, notification); err != nil {
		logger.Ctx(ctx).Error("notifyTaskRecipient: failed to create notification",
			logger.Int("owner_id", int(task.OwnerID)),
			logger.Int("task_id", int(task.ID)),
			logger.Int("recipient_id", int(recipientID)),
			logger.String("type", notificationType),
			logger.Error(err))
	}
}

func mentionedUsernames(content string) []string {
	usernames := make([]string, 0)
	seen := make(map[string]struct{})
	for index := 0; index < len(content); index++ {
		if content[index] != '@' || (index > 0 && isMentionPrefixChar(content[index-1])) {
			continue
		}

		start := index + 1
		if start >= len(content) || !isMentionFirstChar(content[start]) {
			continue
		}
		end := start
		for end < len(content) && isMentionTokenChar(content[end]) {
			end++
		}
		if end == start {
			continue
		}

		username := strings.ToLower(content[start:end])
		if _, found := seen[username]; found {
			index = end - 1
			continue
		}
		seen[username] = struct{}{}
		usernames = append(usernames, username)
		index = end - 1
	}
	return usernames
}

func isMentionTokenChar(character byte) bool {
	return (character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9') ||
		character == '.' || character == '_' || character == '-'
}

func isMentionFirstChar(character byte) bool {
	return (character >= 'a' && character <= 'z') ||
		(character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9')
}

func isMentionPrefixChar(character byte) bool {
	return isMentionTokenChar(character) || character == '@'
}

func (s *Service) announceCommentNotifications(ctx context.Context, task domain.Task, actorID int64, content string) {
	s.notifyTaskRecipient(ctx, task, assigneeID(task.AssigneeID), actorID,
		commentNotification, commentNotificationTitle, content,
		func(settings domain.UserSettings) bool { return settings.TaskComments })

	for _, username := range mentionedUsernames(content) {
		recipientID, found, err := s.repo.LookupUserByUsername(ctx, username, task.BusinessID)
		if err != nil {
			logger.Ctx(ctx).Error("announceCommentNotifications: failed to resolve mention",
				logger.Int("owner_id", int(task.OwnerID)),
				logger.Int("task_id", int(task.ID)),
				logger.String("username", username),
				logger.Error(err))
			continue
		}
		if !found {
			continue
		}
		s.notifyTaskRecipient(ctx, task, recipientID, actorID,
			mentionNotification, mentionNotificationTitle, content,
			func(settings domain.UserSettings) bool { return settings.Mentions })
	}
}

func (s *Service) announceStatusChange(ctx context.Context, task domain.Task, actorID int64, status string) {
	s.notifyTaskRecipient(ctx, task, assigneeID(task.AssigneeID), actorID,
		statusNotification, statusNotificationTitle, status,
		func(settings domain.UserSettings) bool { return settings.StatusChanges })
}

// publishUserTaskEvent sends a task event to one person's own user topic.
//
// This is the only place a task event reaches a `user:<id>` topic, and it only
// ever addresses the two people a reassignment is about, never a third party's
// topic. Topic authorization is checked once at subscribe time rather than per
// event, and the `user` scope resolves to "the caller must be that user", so a
// self-scoped topic needs no membership reasoning of its own.
//
// userID 0 means the task was unassigned, or was assigned from nobody. There is
// no topic to publish to.
func publishUserTaskEvent(userID, taskID, actorID int64) {
	if userID == 0 {
		return
	}
	wsbus.Publish(wsbus.Event{
		Topic:    fmt.Sprintf("user:%d", userID),
		Type:     taskUpdatedEvent,
		EntityID: taskID,
		ActorID:  actorID,
	})
}

func (s *Service) UpdateTaskStatus(ctx context.Context, input domain.UpdateTaskStatusInput, userID int64) error {
	existing, err := s.repo.GetTaskByID(ctx, input.ID)
	if err != nil {
		return err
	}

	// The my-tasks list filters on status and hides done and cancelled work, so
	// a status change can add or remove the task from the assignee's list.
	defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))

	if existing.Status != input.Status {
		if actErr := s.repo.LogActivity(ctx, input.ID, userID, "status_changed", existing.Status, input.Status); actErr != nil {
			logger.Ctx(ctx).Error("UpdateTaskStatus: failed to log activity", logger.Error(actErr))
		}
	}

	if updateErr := s.repo.UpdateTaskStatus(ctx, input); updateErr != nil {
		return updateErr
	}

	if existing.Status != input.Status {
		notificationTask := existing
		notificationTask.Status = input.Status
		s.announceStatusChange(ctx, notificationTask, userID, input.Status)
	}
	return nil
}
