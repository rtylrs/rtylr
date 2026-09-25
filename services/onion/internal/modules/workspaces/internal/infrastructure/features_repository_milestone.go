package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

func (r *Repository) CreateMilestone(ctx context.Context, input domain.MilestoneInput) (int64, error) {
	res, err := r.queries.CreateMilestone(ctx, sqlc.CreateMilestoneParams{
		OwnerID:     input.CreatedBy,
		BusinessID:  input.BusinessID,
		ProjectID:   input.ProjectID,
		WorkspaceID: sql.NullInt64{Int64: input.WorkspaceID, Valid: true},
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		DueDate:     sql.NullTime{Time: input.DueDate, Valid: true},
		CreatedBy:   sql.NullInt64{Int64: input.CreatedBy, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateMilestone failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

// updateMilestone writes the milestone columns both APIs share. dateColumn
// differs by generation: v1 stores its target date in due_date, while the v2
// list reads target_date, so a v2 update that wrote due_date would not be
// visible to the client that sent it.
func (r *Repository) updateMilestone(ctx context.Context, input domain.UpdateMilestoneInput, dateColumn string) error {
	parts := []string{"name = ?"}
	args := []interface{}{input.Name}

	if input.Description != "" {
		parts = append(parts, "description = ?")
		args = append(args, input.Description)
	}
	if input.Status != "" {
		parts = append(parts, "status = ?")
		args = append(args, input.Status)
		if input.Status == "completed" {
			parts = append(parts, "completed_at = NOW()")
		}
	}
	if !input.DueDate.IsZero() {
		parts = append(parts, dateColumn+" = ?")
		args = append(args, input.DueDate.Format("2006-01-02"))
	}
	// Trailing binds: id, then the five placeholders inside memberPredicate.
	args = append(args, input.ID, input.ActorID, input.ActorID, input.ActorID, input.ActorID, input.ActorID)

	res, err := r.pool.ExecContext(ctx,
		fmt.Sprintf("UPDATE flow_milestones SET %s WHERE id = ? AND %s", strings.Join(parts, ", "), memberPredicate("flow_milestones")),
		args...,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateMilestone failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) UpdateMilestone(ctx context.Context, input domain.UpdateMilestoneInput) error {
	return r.updateMilestone(ctx, input, "due_date")
}

func (r *Repository) UpdateMilestoneV2(ctx context.Context, input domain.UpdateMilestoneInput) error {
	return r.updateMilestone(ctx, input, "target_date")
}

// ─────────────────────────────────────────────
// Views
// ─────────────────────────────────────────────

// ListViews returns the shared views plus the caller's own private ones.
func (r *Repository) ListViews(ctx context.Context, businessID, workspaceID, callerID, limit, offset int64) ([]domain.View, int64, error) {
	total, err := r.queries.CountViewsByWorkspace(ctx, sqlc.CountViewsByWorkspaceParams{
		BusinessID:  businessID,
		WorkspaceID: workspaceID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("CountViewsByWorkspace failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListViews(ctx, sqlc.ListViewsParams{
		BusinessID:  businessID,
		WorkspaceID: workspaceID,
		// Drives the private-view filter. Valid: true unconditionally, because a
		// NULL here would compare against created_by as NULL and hide every
		// private view from its own creator too.
		CreatedBy: sql.NullInt64{Int64: callerID, Valid: true},
		Limit:     int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:    int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListViews failed", logger.Error(err))
		return nil, 0, err
	}

	views := make([]domain.View, 0, len(rows))
	for _, row := range rows {
		sortDir := ""
		if row.SortDir.Valid {
			sortDir = string(row.SortDir.FlowViewsSortDir)
		}
		v := domain.View{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			Name:        row.Name,
			Icon:        row.Icon.String,
			GroupBy:     row.GroupBy.String,
			SortBy:      row.SortBy.String,
			SortDir:     sortDir,
			Visibility:  string(row.Visibility),
			CreatedBy:   row.CreatedBy.Int64,
			CreatedAt:   row.CreatedAt.Time,
		}
		if len(row.Filters) > 0 {
			if unmarshalErr := json.Unmarshal(row.Filters, &v.Filters); unmarshalErr != nil {
				logger.Ctx(ctx).Warn("failed to unmarshal view filters", logger.Error(unmarshalErr))
			}
		}
		if v.Filters == nil {
			v.Filters = map[string]interface{}{}
		}
		views = append(views, v)
	}
	return views, total, nil
}

// visibilityOrShared normalises an unset visibility to "shared". Sending "" to
// an ENUM column is rejected, and defaulting to private would hide views that
// used to be visible to everyone.
func visibilityOrShared(v string) string {
	if v == "" {
		return "shared"
	}
	return v
}

func (r *Repository) CreateView(ctx context.Context, input domain.ViewInput) (int64, error) {
	filtersJSON, err := json.Marshal(input.Filters)
	if err != nil {
		return 0, err
	}
	res, err := r.queries.CreateView(ctx, sqlc.CreateViewParams{
		OwnerID:     input.BusinessID,
		BusinessID:  input.BusinessID,
		WorkspaceID: input.WorkspaceID,
		Name:        input.Name,
		Icon:        sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		Filters:     filtersJSON,
		GroupBy:     sql.NullString{String: input.GroupBy, Valid: input.GroupBy != ""},
		SortBy:      sql.NullString{String: input.SortBy, Valid: input.SortBy != ""},
		SortDir:     sqlc.NullFlowViewsSortDir{FlowViewsSortDir: sqlc.FlowViewsSortDir(input.SortDir), Valid: input.SortDir != ""},
		// Empty means shared, matching the column default and how every view
		// behaved before the field existed.
		Visibility: sqlc.FlowViewsVisibility(visibilityOrShared(input.Visibility)),
		CreatedBy:  sql.NullInt64{Int64: input.CreatedBy, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateView failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateView(ctx context.Context, input domain.UpdateViewInput) error {
	filtersJSON, err := json.Marshal(input.Filters)
	if err != nil {
		return err
	}
	res, err := r.queries.UpdateView(ctx, sqlc.UpdateViewParams{
		Name:       input.Name,
		Icon:       sql.NullString{String: input.Icon, Valid: input.Icon != ""},
		Filters:    filtersJSON,
		GroupBy:    sql.NullString{String: input.GroupBy, Valid: input.GroupBy != ""},
		SortBy:     sql.NullString{String: input.SortBy, Valid: input.SortBy != ""},
		SortDir:    sqlc.NullFlowViewsSortDir{FlowViewsSortDir: sqlc.FlowViewsSortDir(input.SortDir), Valid: input.SortDir != ""},
		Visibility: sqlc.FlowViewsVisibility(visibilityOrShared(input.Visibility)),
		ID:         input.ID,
		BusinessID: input.BusinessID,
		CreatedBy:  sql.NullInt64{Int64: input.ActorID, Valid: true},
		OwnerID:    input.ActorID,
		UserID:     input.ActorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("UpdateView failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func (r *Repository) DeleteView(ctx context.Context, id, actorID, businessID int64) error {
	res, err := r.queries.DeleteView(ctx, sqlc.DeleteViewParams{
		ID:         id,
		BusinessID: businessID,
		CreatedBy:  sql.NullInt64{Int64: actorID, Valid: true},
		OwnerID:    actorID,
		UserID:     actorID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("DeleteView failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}
