package application

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"time"

	"github.com/voxire/onion/internal/modules/flow/workspaces/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/wsbus"
)

func (s *Service) DeleteLabel(ctx context.Context, id, actorID, businessID int64) error {
	return s.repo.DeleteLabel(ctx, id, actorID, businessID)
}

// SetTaskLabels is a task mutation: labels are part of every my-tasks payload,
// so the assignee's cached pages have to go with the write, exactly as they do
// for the other task-visible write paths.
func (s *Service) SetTaskLabels(ctx context.Context, taskID, actorID int64, labelIDs []int64) error {
	existing, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("SetTaskLabels: failed to read task, my-tasks cache left to expire",
			logger.Error(err))
	} else {
		defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))
	}

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
	now := time.Now().UTC()
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

// ListViews returns the business's shared views plus the caller's own private
// ones. callerID is required rather than optional: passing 0 would hide every
// private view from its own creator.
func (s *Service) ListViews(ctx context.Context, businessID, workspaceID, callerID, limit, offset int64) ([]domain.View, int64, error) {
	return s.repo.ListViews(ctx, businessID, workspaceID, callerID, limit, offset)
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

func (s *Service) DeleteView(ctx context.Context, id, actorID, businessID int64) error {
	return s.repo.DeleteView(ctx, id, actorID, businessID)
}

// ─────────────────────────────────────────────
// Insights
// ─────────────────────────────────────────────

func (s *Service) GetInsights(ctx context.Context, businessID, workspaceID int64) (domain.Insights, error) {
	ins, err := s.repo.GetInsights(ctx, workspaceID)
	if err != nil {
		return domain.Insights{}, err
	}
	minutesByUser := make(map[int64]int64, len(ins.TimeByMember))
	for _, member := range ins.TimeByMember {
		userID, parseErr := strconv.ParseInt(member.ID, 10, 64)
		if parseErr != nil || userID <= 0 {
			return domain.Insights{}, fmt.Errorf("invalid Flow insights member id %q", member.ID)
		}
		minutesByUser[userID] += member.Minutes
	}
	if s.workdayLines != nil {
		const batchSize int32 = 500
		var afterTaskID int64
		for {
			taskIDs, listErr := s.repo.ListInsightsTaskIDs(ctx, workspaceID, afterTaskID, batchSize)
			if listErr != nil {
				return domain.Insights{}, listErr
			}
			if len(taskIDs) == 0 {
				break
			}
			rows, sumErr := s.workdayLines.SumWorkDayLineMinutesByMemberAndTaskIDs(ctx, businessID, taskIDs)
			if sumErr != nil {
				return domain.Insights{}, sumErr
			}
			for _, row := range rows {
				if row.UserID > 0 {
					minutesByUser[row.UserID] += row.Minutes
				}
			}
			afterTaskID = taskIDs[len(taskIDs)-1]
			if len(taskIDs) < int(batchSize) {
				break
			}
		}
	}

	userIDs := make([]int64, 0, len(minutesByUser))
	for userID := range minutesByUser {
		userIDs = append(userIDs, userID)
	}
	sort.Slice(userIDs, func(i, j int) bool {
		left, right := minutesByUser[userIDs[i]], minutesByUser[userIDs[j]]
		if left == right {
			return userIDs[i] < userIDs[j]
		}
		return left > right
	})
	if len(userIDs) > 10 {
		userIDs = userIDs[:10]
	}
	profiles := map[int64]InsightsMemberProfile{}
	if s.memberProfiles != nil && len(userIDs) > 0 {
		profiles, err = s.memberProfiles.GetInsightsMemberProfiles(ctx, userIDs)
		if err != nil {
			return domain.Insights{}, err
		}
	}

	ins.TimeByMember = make([]domain.InsightsMemberTime, 0, len(userIDs))
	for _, userID := range userIDs {
		minutes := minutesByUser[userID]
		profile := profiles[userID]
		name := profile.Name
		if name == "" {
			name = fmt.Sprintf("User %d", userID)
		}
		ins.TimeByMember = append(ins.TimeByMember, domain.InsightsMemberTime{
			ID: strconv.FormatInt(userID, 10), Name: name,
			Hours: math.Round((float64(minutes)/60)*10) / 10,
		})
	}
	return ins, nil
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

func (s *Service) ListNotifications(ctx context.Context, userID, limit, offset int64) ([]domain.Notification, int64, error) {
	return s.repo.ListNotifications(ctx, userID, limit, offset)
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

func (s *Service) GetProjectTaskCountsByProjectIDs(ctx context.Context, projectIDs []int64) (map[int64]domain.ProjectTaskCounts, error) {
	return s.repo.GetProjectTaskCountsByProjectIDs(ctx, projectIDs)
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

// MoveTaskToColumn is the board's drag-and-drop write path, reached from the
// flow project API. It reassigns the task to a different column, and the column
// a task sits in is part of every my-tasks payload, so the assignee's cached
// pages have to go with it.
func (s *Service) MoveTaskToColumn(ctx context.Context, taskID, columnID, actorID int64) error {
	existing, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("MoveTaskToColumn: failed to read task, my-tasks cache left to expire",
			logger.Error(err))
	} else {
		defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))
	}

	return s.repo.MoveTaskToColumn(ctx, taskID, columnID, actorID)
}

// MoveTaskToColumnOrdered is MoveTaskToColumn plus the card's position: when
// orderedTaskIDs is non-empty the move and the new sort_order of every listed
// card are saved in one transaction. An empty list behaves exactly like
// MoveTaskToColumn. Validation failures wrap domain.ErrInvalidTaskOrder or
// domain.ErrTaskOrderOutsideProject.
func (s *Service) MoveTaskToColumnOrdered(ctx context.Context, taskID, columnID, actorID int64, orderedTaskIDs []int64) error {
	if err := domain.ValidateTaskOrder(taskID, orderedTaskIDs); err != nil {
		return err
	}
	if len(orderedTaskIDs) == 0 {
		return s.MoveTaskToColumn(ctx, taskID, columnID, actorID)
	}

	// Only the moved task's assignee is invalidated. The other reordered cards'
	// assignees are not looked up: that would be one read per id (up to 500)
	// for a position-only change, and my-tasks pages are ordered by their own
	// rules, so a stale sort_order there heals when the short-lived cache expires.
	existing, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("MoveTaskToColumnOrdered: failed to read task, my-tasks cache left to expire",
			logger.Error(err))
	} else {
		defer invalidateUserScope(ctx, assigneeID(existing.AssigneeID))
	}

	return s.repo.MoveTaskToColumnOrdered(ctx, taskID, columnID, actorID, orderedTaskIDs)
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
// to use (and lazily create) the caller's personal workspace. businessID must
// be the caller's resolved tenant id (tenant.ResolveBusinessID), not their
// user id — this is what the created project's business_id column stores.
func (s *Service) EnsureMyTasksProject(
	ctx context.Context,
	userID, businessID, workspaceID int64,
) (projectID, resolvedWorkspaceID int64, err error) {
	if workspaceID == 0 {
		wsID, wsErr := s.repo.CreatePersonalWorkspace(ctx, userID, businessID)
		if wsErr != nil {
			return 0, 0, wsErr
		}
		workspaceID = wsID
	} else if !s.repo.IsWorkspaceMember(ctx, userID, workspaceID) {
		return 0, 0, fmt.Errorf("workspace access denied")
	}

	if existingID, existingName := s.repo.GetMyTasksProject(ctx, userID, workspaceID); existingID != 0 {
		if existingName == "Inbox" {
			if renameErr := s.repo.RenameProject(ctx, existingID, userID, workspaceID, "My Tasks"); renameErr != nil {
				logger.Ctx(ctx).Warn("migrate Inbox->My Tasks failed", logger.Error(renameErr))
			} else {
				invalidateWorkspacesScope(ctx, businessID)
			}
		}
		return existingID, workspaceID, nil
	}

	id, err := s.repo.CreateProject(ctx, domain.ProjectInput{
		WorkspaceID: workspaceID,
		OwnerID:     userID,
		BusinessID:  businessID,
		Name:        "My Tasks",
		Description: "Your personal task board. Quick-created tasks land here.",
		Icon:        "✅",
		Color:       "#64C5B9",
		Visibility:  domain.VisibilityPrivate,
		IsPersonal:  true,
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
	invalidateWorkspacesScope(ctx, businessID)
	return id, workspaceID, nil
}

func (s *Service) SeedDefaultColumns(ctx context.Context, projectID, workspaceID, createdBy int64) error {
	return s.repo.SeedDefaultColumns(ctx, projectID, workspaceID, createdBy)
}
