package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func mapWorkloadMemberTasks(rows []sqlc.GetTeamWorkloadMemberTasksRow) []domain.Task {
	tasks := make([]domain.Task, 0, len(rows))
	for _, row := range rows {
		t := domain.Task{
			ID:          row.ID,
			ProjectID:   row.ProjectID,
			WorkspaceID: row.WorkspaceID,
			OwnerID:     row.OwnerID,
			Title:       row.Title,
			Description: row.Description.String,
			Status:      string(row.Status.FlowTasksStatus),
			Priority:    string(row.Priority.FlowTasksPriority),
			ReporterID:  row.ReporterID,
			SortOrder:   int(row.SortOrder.Int32),
			Archived:    row.Archived.Bool,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
			Identifier:  "T-" + strconv.FormatInt(row.ID, 10),
		}
		if row.ParentTaskID.Valid {
			t.ParentTaskID = &row.ParentTaskID.Int64
		}
		if row.SectionID.Valid {
			t.SectionID = &row.SectionID.Int64
		}
		if row.AssigneeID.Valid {
			t.AssigneeID = &row.AssigneeID.Int64
		}
		if row.DueDate.Valid {
			t.DueDate = &row.DueDate.Time
		}
		if row.StartDate.Valid {
			t.StartDate = &row.StartDate.Time
		}
		if row.CompletedAt.Valid {
			t.CompletedAt = &row.CompletedAt.Time
		}
		if row.EstimatedHours.Valid {
			if f, parseErr := strconv.ParseFloat(row.EstimatedHours.String, 64); parseErr == nil {
				t.EstimatedHours = &f
			}
		}
		if row.ActualHours.Valid {
			if f, parseErr := strconv.ParseFloat(row.ActualHours.String, 64); parseErr == nil {
				t.ActualHours = &f
			}
		}
		if len(row.Labels) > 0 {
			if err := json.Unmarshal(row.Labels, &t.Labels); err != nil {
				logger.Global().Warn("failed to parse task labels JSON from database",
					logger.Any("task_id", row.ID), logger.Error(err))
			}
		}
		tasks = append(tasks, t)
	}
	return tasks
}

// ─────────────────────────────────────────────
// Roadmap (derived from projects + milestones)
// ─────────────────────────────────────────────

func (r *Repository) GetRoadmap(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Project, int64, error) {
	total, err := r.queries.CountRoadmapProjects(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("CountRoadmapProjects failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.GetRoadmap(ctx, sqlc.GetRoadmapParams{
		WorkspaceID: workspaceID,
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetRoadmap failed", logger.Error(err))
		return nil, 0, err
	}

	projects := make([]domain.Project, 0, len(rows))
	for _, row := range rows {
		p := domain.Project{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			OwnerID:     row.OwnerID,
			Name:        row.Name,
			Description: row.Description.String,
			Icon:        row.Icon.String,
			Color:       row.Color.String,
			Status:      string(row.Status.FlowProjectsStatus),
			CreatedBy:   row.CreatedBy,
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
		if row.StartDate.Valid {
			p.StartDate = &row.StartDate.Time
		}
		if row.DueDate.Valid {
			p.DueDate = &row.DueDate.Time
		}
		projects = append(projects, p)
	}
	return projects, total, nil
}

// ─────────────────────────────────────────────
// Notifications
// ─────────────────────────────────────────────

func (r *Repository) ListNotifications(ctx context.Context, userID, limit, offset int64) ([]domain.Notification, int64, error) {
	total, err := r.queries.CountNotifications(ctx, userID)
	if err != nil {
		logger.Ctx(ctx).Error("CountNotifications failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListNotifications(ctx, sqlc.ListNotificationsParams{
		UserID: userID,
		Limit:  int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset: int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListNotifications failed", logger.Error(err))
		return nil, 0, err
	}

	notifs := make([]domain.Notification, 0, len(rows))
	for _, row := range rows {
		actorInitials := row.ActorInitials
		if actorInitials == "" {
			actorInitials = initialsOf(row.ActorName)
		}
		n := domain.Notification{
			ID:            row.ID,
			UserID:        row.UserID,
			Type:          string(row.Type),
			Title:         row.Title,
			Description:   row.Description,
			TaskTitle:     row.TaskTitle,
			ActorName:     row.ActorName,
			ActorInitials: actorInitials,
			Project:       row.Project,
			ProjectColor:  row.ProjectColor,
			IsRead:        row.IsRead.Bool,
			CreatedAt:     row.CreatedAt,
		}
		if row.WorkspaceID.Valid {
			n.WorkspaceID = &row.WorkspaceID.Int64
		}
		if row.TaskID.Valid {
			n.TaskID = &row.TaskID.Int64
		}
		if row.ActorID.Valid {
			n.ActorID = &row.ActorID.Int64
		}
		notifs = append(notifs, n)
	}
	return notifs, total, nil
}

func (r *Repository) CreateNotification(ctx context.Context, input domain.CreateNotificationInput) error {
	err := r.queries.CreateNotification(ctx, sqlc.CreateNotificationParams{
		UserID:        input.UserID,
		WorkspaceID:   nullInt64PtrFromPtr(input.WorkspaceID),
		Type:          sqlc.FlowNotificationsType(input.Type),
		Title:         input.Title,
		Description:   sql.NullString{String: input.Description, Valid: input.Description != ""},
		TaskID:        nullInt64PtrFromPtr(input.TaskID),
		TaskTitle:     sql.NullString{String: input.TaskTitle, Valid: input.TaskTitle != ""},
		ActorID:       nullInt64PtrFromPtr(input.ActorID),
		ActorInitials: sql.NullString{String: input.ActorInitials, Valid: input.ActorInitials != ""},
		Project:       sql.NullString{String: input.Project, Valid: input.Project != ""},
		ProjectColor:  sql.NullString{String: input.ProjectColor, Valid: input.ProjectColor != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateNotification failed", logger.Error(err))
	}
	return err
}

func (r *Repository) MarkNotificationRead(ctx context.Context, id, userID int64) error {
	return r.queries.MarkNotificationRead(ctx, sqlc.MarkNotificationReadParams{ID: id, UserID: userID})
}

func (r *Repository) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	return r.queries.MarkAllNotificationsRead(ctx, userID)
}

// ─────────────────────────────────────────────
// User Settings
// ─────────────────────────────────────────────
