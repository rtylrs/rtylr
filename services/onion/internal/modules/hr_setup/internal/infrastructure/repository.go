package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_setup/internal/infrastructure/persistence/sqlc"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) GetSetupSteps(ctx context.Context, businessID int64) ([]*domain.SetupStep, error) {
	rows, err := r.queries.ListSetupSteps(ctx, businessID)
	if err != nil {
		return nil, err
	}

	steps := make([]*domain.SetupStep, 0, len(rows))
	for _, row := range rows {
		s := &domain.SetupStep{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			StepKey:    row.StepKey,
		}
		if row.CompletedAt.Valid {
			t := row.CompletedAt.Time
			s.CompletedAt = &t
		}
		steps = append(steps, s)
	}
	return steps, nil
}

func (r *Repository) UpsertSetupStep(ctx context.Context, businessID int64, stepKey string, completedAt time.Time) error {
	return r.queries.UpsertSetupStep(ctx, sqlc.UpsertSetupStepParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		StepKey:     stepKey,
		CompletedAt: sql.NullTime{Time: completedAt, Valid: true},
	})
}

func (r *Repository) GetHRConfig(ctx context.Context, ownerID int64) (*domain.HRConfig, error) {
	row, err := r.queries.GetHRConfig(ctx, ownerID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConfigNotFound
		}
		return nil, err
	}

	cfg := &domain.HRConfig{
		ID:      row.ID,
		OwnerID: row.OwnerID,
	}
	if row.SetupCompletedAt.Valid {
		cfg.SetupCompletedAt = &row.SetupCompletedAt.Time
	}
	cfg.CreatedAt = row.CreatedAt
	cfg.UpdatedAt = row.UpdatedAt
	return cfg, nil
}

func (r *Repository) UpsertHRConfig(ctx context.Context, ownerID int64, completedAt *time.Time) (*domain.HRConfig, error) {
	var completed sql.NullTime
	if completedAt != nil {
		completed = sql.NullTime{Time: *completedAt, Valid: true}
	}
	if err := r.queries.UpsertHRConfig(ctx, sqlc.UpsertHRConfigParams{
		OwnerID:          ownerID,
		SetupCompletedAt: completed,
	}); err != nil {
		return nil, err
	}
	return r.GetHRConfig(ctx, ownerID)
}
