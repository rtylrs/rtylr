package infrastructure

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/voxire/onion/internal/modules/workspaces/internal/domain"
	"github.com/voxire/onion/internal/modules/workspaces/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

// ─────────────────────────────────────────────
// Cycles
// ─────────────────────────────────────────────

func (r *Repository) ListCycles(ctx context.Context, workspaceID, limit, offset int64) ([]domain.Cycle, int64, error) {
	total, err := r.queries.CountCycles(ctx, workspaceID)
	if err != nil {
		logger.Ctx(ctx).Error("CountCycles failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListCycles(ctx, sqlc.ListCyclesParams{
		WorkspaceID: workspaceID,
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListCycles failed", logger.Error(err))
		return nil, 0, err
	}

	cycles := make([]domain.Cycle, 0, len(rows))
	now := time.Now().UTC()
	for _, row := range rows {
		c := domain.Cycle{
			ID:          row.ID,
			WorkspaceID: row.WorkspaceID,
			Name:        row.Name,
			Description: row.Description.String,
			Status:      string(row.Status.FlowCyclesStatus),
			StartDate:   row.StartDate.Time,
			EndDate:     row.EndDate.Time,
			Scope:       int(row.Scope),
			Completed:   int(row.Completed),
			CreatedBy:   row.CreatedBy.Int64,
			CreatedAt:   row.CreatedAt.Time,
		}
		if row.ProjectID.Valid {
			c.ProjectID = &row.ProjectID.Int64
		}
		if c.Status != "completed" {
			if now.Before(c.StartDate) {
				c.Status = "upcoming"
			} else if now.After(c.EndDate) {
				c.Status = "completed"
			} else {
				c.Status = "active"
			}
		}
		cycles = append(cycles, c)
	}
	return cycles, total, nil
}

func (r *Repository) CreateCycle(ctx context.Context, input domain.CycleInput) (int64, error) {
	res, err := r.queries.CreateCycle(ctx, sqlc.CreateCycleParams{
		OwnerID:     input.BusinessID,
		BusinessID:  input.BusinessID,
		WorkspaceID: input.WorkspaceID,
		ProjectID:   nullInt64PtrFromPtr(input.ProjectID),
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		StartDate:   sql.NullTime{Time: input.StartDate, Valid: true},
		EndDate:     sql.NullTime{Time: input.EndDate, Valid: true},
		CreatedBy:   sql.NullInt64{Int64: input.CreatedBy, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("CreateCycle failed", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func (r *Repository) UpdateCycle(ctx context.Context, input domain.UpdateCycleInput) error {
	parts := []string{"name = ?"}
	args := []interface{}{input.Name}

	if input.Status != "" {
		parts = append(parts, "status = ?")
		args = append(args, input.Status)
	}
	if !input.EndDate.IsZero() {
		parts = append(parts, "end_date = ?")
		args = append(args, input.EndDate.Format("2006-01-02"))
	}
	// Trailing binds: id, then the five placeholders inside memberPredicate - all
	// the acting user, so the workspace owner branch stops comparing the owner
	// id against the business id.
	args = append(args, input.ID, input.ActorID, input.ActorID, input.ActorID, input.ActorID, input.ActorID)

	res, err := r.pool.ExecContext(ctx,
		fmt.Sprintf("UPDATE flow_cycles SET %s WHERE id = ? AND %s", strings.Join(parts, ", "), memberPredicate("flow_cycles")),
		args...,
	)
	if err != nil {
		logger.Ctx(ctx).Error("UpdateCycle failed", logger.Error(err))
		return err
	}
	return requireAffected(res)
}

func nullInt64PtrFromPtr(v *int64) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *v, Valid: true}
}

// ─────────────────────────────────────────────
// Milestones
// ─────────────────────────────────────────────

func (r *Repository) ListMilestones(ctx context.Context, workspaceID int64, projectID *int64, limit, offset int64) ([]domain.Milestone, int64, error) {
	if projectID != nil {
		total, countErr := r.queries.CountMilestonesByProject(ctx, sqlc.CountMilestonesByProjectParams{
			WorkspaceID: sql.NullInt64{Int64: workspaceID, Valid: true},
			ProjectID:   *projectID,
		})
		if countErr != nil {
			logger.Ctx(ctx).Error("CountMilestonesByProject failed", logger.Error(countErr))
			return nil, 0, countErr
		}
		projectRows, listErr := r.queries.ListMilestonesByProject(ctx, sqlc.ListMilestonesByProjectParams{
			WorkspaceID: sql.NullInt64{Int64: workspaceID, Valid: true},
			ProjectID:   *projectID,
			Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
			Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		})
		if listErr != nil {
			logger.Ctx(ctx).Error("ListMilestones failed", logger.Error(listErr))
			return nil, 0, listErr
		}
		return mapMilestoneProjectRows(projectRows), total, nil
	}

	total, err := r.queries.CountMilestonesByWorkspace(ctx, sql.NullInt64{Int64: workspaceID, Valid: true})
	if err != nil {
		logger.Ctx(ctx).Error("CountMilestonesByWorkspace failed", logger.Error(err))
		return nil, 0, err
	}

	rows, err := r.queries.ListMilestonesByWorkspace(ctx, sqlc.ListMilestonesByWorkspaceParams{
		WorkspaceID: sql.NullInt64{Int64: workspaceID, Valid: true},
		Limit:       int32(limit),  // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:      int32(offset), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("ListMilestones failed", logger.Error(err))
		return nil, 0, err
	}
	return mapMilestoneWorkspaceRows(rows), total, nil
}

func mapMilestoneWorkspaceRows(rows []sqlc.ListMilestonesByWorkspaceRow) []domain.Milestone {
	milestones := make([]domain.Milestone, 0, len(rows))
	now := time.Now().UTC()
	for _, row := range rows {
		m := domain.Milestone{
			ID:          row.ID,
			ProjectID:   row.ProjectID,
			WorkspaceID: row.WorkspaceID.Int64,
			Name:        row.Name,
			Description: row.Description.String,
			Status:      string(row.Status.FlowMilestonesStatus),
			DueDate:     row.DueDate.Time,
			CreatedBy:   row.CreatedBy.Int64,
			CreatedAt:   row.CreatedAt.Time,
			TaskCount:   int(row.TaskCount),
			DoneCount:   int(row.DoneCount),
		}
		if row.CompletedAt.Valid {
			m.CompletedAt = &row.CompletedAt.Time
		}
		if row.ProjectName.Valid {
			m.Project = &domain.MilestoneProject{Name: row.ProjectName.String, Color: row.ProjectColor.String}
		}
		if m.TaskCount > 0 {
			m.Progress = int(float64(m.DoneCount) / float64(m.TaskCount) * 100)
		}
		if m.Status != "completed" && m.DueDate.Before(now) {
			m.Status = "at_risk"
		}
		milestones = append(milestones, m)
	}
	return milestones
}

func mapMilestoneProjectRows(rows []sqlc.ListMilestonesByProjectRow) []domain.Milestone {
	milestones := make([]domain.Milestone, 0, len(rows))
	now := time.Now().UTC()
	for _, row := range rows {
		m := domain.Milestone{
			ID:          row.ID,
			ProjectID:   row.ProjectID,
			WorkspaceID: row.WorkspaceID.Int64,
			Name:        row.Name,
			Description: row.Description.String,
			Status:      string(row.Status.FlowMilestonesStatus),
			DueDate:     row.DueDate.Time,
			CreatedBy:   row.CreatedBy.Int64,
			CreatedAt:   row.CreatedAt.Time,
			TaskCount:   int(row.TaskCount),
			DoneCount:   int(row.DoneCount),
		}
		if row.CompletedAt.Valid {
			m.CompletedAt = &row.CompletedAt.Time
		}
		if row.ProjectName.Valid {
			m.Project = &domain.MilestoneProject{Name: row.ProjectName.String, Color: row.ProjectColor.String}
		}
		if m.TaskCount > 0 {
			m.Progress = int(float64(m.DoneCount) / float64(m.TaskCount) * 100)
		}
		if m.Status != "completed" && m.DueDate.Before(now) {
			m.Status = "at_risk"
		}
		milestones = append(milestones, m)
	}
	return milestones
}
