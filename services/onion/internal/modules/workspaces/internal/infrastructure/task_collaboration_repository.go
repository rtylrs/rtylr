package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// Comments
// ─────────────────────────────────────────────

func (r *Repository) CreateComment(ctx context.Context, taskID, userID int64, content string) (int64, error) {
	res, err := r.queries.CreateComment(ctx, sqlc.CreateCommentParams{
		TaskID:  taskID,
		UserID:  userID,
		Content: content,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateComment failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetCommentByID(ctx context.Context, commentID int64) (domain.TaskComment, error) {
	row, err := r.queries.GetCommentByID(ctx, commentID)
	if err != nil {
		logger.Ctx(ctx).Error("GetCommentByID failed", logger.Error(err))
		return domain.TaskComment{}, err
	}
	return domain.TaskComment{
		ID:            row.ID,
		TaskID:        row.TaskID,
		UserID:        row.UserID,
		ActorName:     row.ActorName,
		ActorInitials: initialsOf(row.ActorName),
		Content:       row.Content,
		IsEdited:      row.IsEdited.Bool,
		CreatedAt:     row.CreatedAt.Time,
		UpdatedAt:     row.UpdatedAt.Time,
	}, nil
}

func (r *Repository) GetCommentsByTask(ctx context.Context, taskID int64) ([]domain.TaskComment, error) {
	rows, err := r.queries.GetCommentsByTask(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("GetCommentsByTask failed", logger.Error(err))
		return nil, err
	}

	comments := make([]domain.TaskComment, 0, len(rows))
	for _, row := range rows {
		comments = append(comments, domain.TaskComment{
			ID:            row.ID,
			TaskID:        row.TaskID,
			UserID:        row.UserID,
			ActorName:     row.ActorName,
			ActorInitials: initialsOf(row.ActorName),
			Content:       row.Content,
			IsEdited:      row.IsEdited.Bool,
			CreatedAt:     row.CreatedAt.Time,
			UpdatedAt:     row.UpdatedAt.Time,
		})
	}
	return comments, nil
}

func (r *Repository) DeleteComment(ctx context.Context, commentID, userID int64) error {
	err := r.queries.DeleteComment(ctx, sqlc.DeleteCommentParams{
		ID:     commentID,
		UserID: userID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteComment failed", logger.Error(err))
		return err
	}
	return nil
}

// GetCommentTaskID returns the task a comment belongs to. Used by the
// comment handlers to route realtime events to the correct task topic
// before the row disappears.
func (r *Repository) GetCommentTaskID(ctx context.Context, commentID int64) (int64, error) {
	return r.queries.GetCommentTaskID(ctx, commentID)
}

// GetAttachmentTaskID returns the task an attachment hangs off, so a delete can
// be authorized against that task's project.
func (r *Repository) GetAttachmentTaskID(ctx context.Context, attachmentID int64) (int64, error) {
	return r.queries.GetAttachmentTaskID(ctx, attachmentID)
}

// Activity
// ─────────────────────────────────────────────

func (r *Repository) LogActivity(ctx context.Context, taskID, userID int64, action, oldValue, newValue string) error {
	err := r.queries.LogActivity(ctx, sqlc.LogActivityParams{
		TaskID:   taskID,
		UserID:   userID,
		Action:   action,
		OldValue: sql.NullString{String: oldValue, Valid: oldValue != ""},
		NewValue: sql.NullString{String: newValue, Valid: newValue != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("LogActivity failed", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) GetActivityByTask(ctx context.Context, taskID int64, limit int) ([]domain.TaskActivity, error) {
	rows, err := r.queries.GetActivityByTask(ctx, sqlc.GetActivityByTaskParams{
		TaskID: taskID,
		Limit:  int32(limit), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetActivityByTask failed", logger.Error(err))
		return nil, err
	}

	activities := make([]domain.TaskActivity, 0, len(rows))
	for _, row := range rows {
		activities = append(activities, domain.TaskActivity{
			ID:            row.ID,
			TaskID:        row.TaskID,
			UserID:        row.UserID,
			ActorName:     row.ActorName,
			ActorInitials: initialsOf(row.ActorName),
			Action:        row.Action,
			OldValue:      row.OldValue.String,
			NewValue:      row.NewValue.String,
			CreatedAt:     row.CreatedAt.Time,
		})
	}
	return activities, nil
}

// Attachments
// ─────────────────────────────────────────────

func (r *Repository) CreateAttachment(ctx context.Context, input domain.TaskAttachmentInput) (int64, error) {
	res, err := r.queries.CreateAttachment(ctx, sqlc.CreateAttachmentParams{
		TaskID:     input.TaskID,
		OwnerID:    input.OwnerID,
		BusinessID: input.BusinessID,
		Name:       input.Name,
		Url:        input.URL,
		Size:       sql.NullInt64{Int64: input.Size, Valid: true},
		UploadedBy: input.UploadedBy,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateAttachment failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetAttachmentsByTask(ctx context.Context, taskID int64) ([]domain.TaskAttachment, error) {
	rows, err := r.queries.GetAttachmentsByTask(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("GetAttachmentsByTask failed", logger.Error(err))
		return nil, err
	}

	attachments := make([]domain.TaskAttachment, 0, len(rows))
	for _, row := range rows {
		attachments = append(attachments, domain.TaskAttachment{
			ID:         row.ID,
			TaskID:     row.TaskID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			URL:        row.Url,
			Size:       row.Size.Int64,
			UploadedBy: row.UploadedBy,
			CreatedAt:  row.CreatedAt.Time,
		})
	}
	return attachments, nil
}

func (r *Repository) DeleteAttachment(ctx context.Context, attachmentID, businessID int64) error {
	err := r.queries.DeleteAttachment(ctx, sqlc.DeleteAttachmentParams{
		ID:         attachmentID,
		BusinessID: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteAttachment failed", logger.Error(err))
		return err
	}
	return nil
}

// Time Logs
// ─────────────────────────────────────────────

func (r *Repository) CreateTimeLog(ctx context.Context, input domain.TimeLogInput) (int64, error) {
	res, err := r.queries.CreateTimeLog(ctx, sqlc.CreateTimeLogParams{
		OwnerID:         input.OwnerID,
		BusinessID:      input.BusinessID,
		TaskID:          input.TaskID,
		UserID:          input.UserID,
		DurationMinutes: int32(input.DurationMinutes), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Description:     sql.NullString{String: input.Description, Valid: input.Description != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateTimeLog failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetTimeLogsByTask(ctx context.Context, taskID int64) ([]domain.TimeLog, error) {
	rows, err := r.queries.GetTimeLogsByTask(ctx, taskID)
	if err != nil {
		logger.Ctx(ctx).Error("GetTimeLogsByTask failed", logger.Error(err))
		return nil, err
	}

	logs := make([]domain.TimeLog, 0, len(rows))
	for _, row := range rows {
		logs = append(logs, domain.TimeLog{
			ID:              row.ID,
			OwnerID:         row.OwnerID,
			BusinessID:      row.BusinessID,
			TaskID:          row.TaskID,
			UserID:          row.UserID,
			DurationMinutes: int(row.DurationMinutes),
			Description:     row.Description.String,
			LoggedAt:        row.LoggedAt.Time,
			CreatedAt:       row.CreatedAt.Time,
		})
	}
	return logs, nil
}

// Labels
// ─────────────────────────────────────────────

func (r *Repository) CreateLabel(ctx context.Context, input domain.LabelInput) (int64, error) {
	color := input.Color
	if color == "" {
		color = "#6366f1"
	}

	res, err := r.queries.CreateLabel(ctx, sqlc.CreateLabelParams{
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Color:       color,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateLabel failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) GetLabelsByWorkspace(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Label, int64, error) {
	total, err := r.queries.CountLabelsByWorkspace(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("CountLabelsByWorkspace failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.GetLabelsByWorkspace(ctx, sqlc.GetLabelsByWorkspaceParams{
		WorkspaceID: workspaceID,
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("GetLabelsByWorkspace failed", logger.Error(err))
		return nil, 0, err
	}

	labels := make([]domain.Label, 0, len(rows))
	for _, row := range rows {
		labels = append(labels, domain.Label{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			Name:        row.Name,
			Color:       row.Color,
			CreatedAt:   row.CreatedAt.Time,
		})
	}
	return labels, total, nil
}

func (r *Repository) UpdateLabel(ctx context.Context, id, actorID, businessID int64, input domain.LabelInput) error {
	color := input.Color
	if color == "" {
		color = "#6366f1"
	}
	res, err := r.queries.UpdateLabel(ctx, sqlc.UpdateLabelParams{
		Name:       input.Name,
		Color:      color,
		ID:         id,
		BusinessID: businessID,
		OwnerID:    actorID,
		UserID:     actorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateLabel failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) DeleteLabel(ctx context.Context, id, actorID, businessID int64) error {
	res, err := r.queries.DeleteLabel(ctx, sqlc.DeleteLabelParams{
		ID:         id,
		BusinessID: businessID,
		OwnerID:    actorID,
		UserID:     actorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteLabel failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) SetTaskLabels(ctx context.Context, taskID, actorID int64, labelIDs []int64) error {
	labelsJSON, err := json.Marshal(labelIDs)
	if err != nil {
		return err
	}
	res, err := r.queries.SetTaskLabels(ctx, sqlc.SetTaskLabelsParams{
		Labels: labelsJSON,
		ID:     taskID,
		// The same actor fills every branch: project member, project owner,
		// project creator, then the two legacy workspace branches.
		UserID:    actorID,
		OwnerID:   actorID,
		CreatedBy: actorID,
		OwnerID_2: actorID,
		UserID_2:  actorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("SetTaskLabels failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

// Time Logs (extended)
// ─────────────────────────────────────────────

func (r *Repository) GetTimeLogTaskID(ctx context.Context, id int64) (int64, error) {
	return r.queries.GetTimeLogTaskID(ctx, id)
}

func (r *Repository) DeleteTimeLog(ctx context.Context, id, userID int64) error {
	err := r.queries.DeleteTimeLog(ctx, sqlc.DeleteTimeLogParams{ID: id, UserID: userID})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteTimeLog failed", logger.Error(err))
	}
	return err
}

func (r *Repository) RecalcActualHours(ctx context.Context, taskID, actorID int64) error {
	// TaskID and TaskID_2 are the same task, matched against the two stores of
	// hours the query sums (tryrtylr/rtylr#903). The first is sql.NullInt64
	// because hr_time_entries.task_id is nullable; Valid is always true here,
	// since a null would match no row rather than every row.
	err := r.queries.RecalcActualHours(ctx, sqlc.RecalcActualHoursParams{
		TaskID:   sql.NullInt64{Int64: taskID, Valid: true},
		TaskID_2: taskID,
		ID:       taskID,
		// One actor across every branch: project member, project owner, project
		// creator, then the two legacy workspace branches. Without the project
		// clauses this UPDATE matched no row for a v2 task, so a task with hours
		// logged against it kept actual_hours NULL (tryrtylr/rtylr#903).
		UserID:    actorID,
		OwnerID:   actorID,
		CreatedBy: actorID,
		OwnerID_2: actorID,
		UserID_2:  actorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("RecalcActualHours failed", logger.Error(err))
	}
	return err
}
