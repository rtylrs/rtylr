package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/milestones/internal/domain"
	"github.com/voxire/onion/internal/modules/milestones/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64, projectID *int64) ([]domain.Milestone, error) {
	var rows []sqlc.FlowMilestone
	var err error

	if projectID != nil {
		rows, err = r.queries.ListMilestonesByProject(ctx, sqlc.ListMilestonesByProjectParams{
			OwnerID:   ownerID,
			ProjectID: *projectID,
		})
	} else {
		rows, err = r.queries.ListMilestonesByOwner(ctx, ownerID)
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list milestones", logger.Error(err))
		return nil, err
	}

	milestones := make([]domain.Milestone, 0, len(rows))
	for _, row := range rows {
		m := domain.Milestone{
			ID:          row.ID,
			OwnerID:     row.OwnerID,
			ProjectID:   row.ProjectID,
			Name:        row.Name,
			Description: row.Description.String,
			Status:      string(row.Status.FlowMilestonesStatus),
			CreatedAt:   row.CreatedAt.Time,
			UpdatedAt:   row.UpdatedAt.Time,
		}
		if row.TargetDate.Valid {
			m.TargetDate = &row.TargetDate.Time
		}
		milestones = append(milestones, m)
	}
	return milestones, nil
}

func (r *Repository) GetProject(ctx context.Context, projectID int64, ownerID int64) (int64, error) {
	id, err := r.queries.GetProjectByIDAndOwner(ctx, sqlc.GetProjectByIDAndOwnerParams{
		ID:      projectID,
		OwnerID: ownerID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, domain.ErrProjectNotFound
		}
		logger.Ctx(ctx).Error("Failed to verify project ownership",
			logger.Any("project_id", projectID),
			logger.Any("owner_id", ownerID),
			logger.Error(err),
		)
		return 0, err
	}
	return id, nil
}

func (r *Repository) Create(ctx context.Context, input domain.MilestoneInput) (int64, error) {
	result, err := r.queries.CreateMilestone(ctx, sqlc.CreateMilestoneParams{
		OwnerID:     input.OwnerID,
		ProjectID:   input.ProjectID,
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		TargetDate:  nullTimePtrToSql(input.TargetDate),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create milestone", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, input domain.MilestoneUpdateInput) error {
	status := input.Status
	if status == "" {
		status = "upcoming"
	}
	result, err := r.queries.UpdateMilestone(ctx, sqlc.UpdateMilestoneParams{
		Name:        input.Name,
		Description: sql.NullString{String: input.Description, Valid: input.Description != ""},
		TargetDate:  nullTimePtrToSql(input.TargetDate),
		Status:      sqlc.NullFlowMilestonesStatus{FlowMilestonesStatus: sqlc.FlowMilestonesStatus(status), Valid: true},
		ID:          input.ID,
		OwnerID:     input.OwnerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update milestone", logger.Error(err))
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrMilestoneNotFound
	}
	return nil
}

func nullTimePtrToSql(v *time.Time) sql.NullTime {
	if v == nil {
		return sql.NullTime{}
	}
	return sql.NullTime{Time: *v, Valid: true}
}
