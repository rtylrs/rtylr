package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		pool: db,
	}
}

func (r *Repository) GetSetupSteps(ctx context.Context, businessID int64) ([]*domain.SetupStep, error) {
	listSQL, args, err := sq.Select("id", "owner_id", "business_id", "step_key", "completed_at").
		From("hr_setup_steps").
		Where(sq.Eq{"business_id": businessID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, listSQL, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	steps := make([]*domain.SetupStep, 0)
	for rows.Next() {
		var s domain.SetupStep
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.BusinessID, &s.StepKey, &s.CompletedAt); err != nil {
			return nil, err
		}
		steps = append(steps, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return steps, nil
}

func (r *Repository) UpsertSetupStep(ctx context.Context, businessID int64, stepKey string, completedAt time.Time) error {
	upsertSQL, args, err := sq.Insert("hr_setup_steps").
		Columns("owner_id", "business_id", "step_key", "completed_at").
		Values(businessID, businessID, stepKey, completedAt).
		Suffix("ON DUPLICATE KEY UPDATE completed_at = VALUES(completed_at)").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return err
	}

	_, err = r.pool.ExecContext(ctx, upsertSQL, args...)

	return err
}

func (r *Repository) GetHRConfig(ctx context.Context, ownerID int64) (*domain.HRConfig, error) {
	selectSQL, args, err := sq.Select("id", "owner_id", "setup_completed_at", "created_at", "updated_at").
		From("hr_configs").
		Where(sq.Eq{"owner_id": ownerID}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := r.pool.QueryRowContext(ctx, selectSQL, args...)

	var cfg domain.HRConfig
	err = row.Scan(&cfg.ID, &cfg.OwnerID, &cfg.SetupCompletedAt, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrConfigNotFound
		}
		return nil, err
	}

	return &cfg, nil
}

func (r *Repository) UpsertHRConfig(ctx context.Context, ownerID int64, completedAt *time.Time) (*domain.HRConfig, error) {
	upsertSQL, args, err := sq.Insert("hr_configs").
		Columns("owner_id", "setup_completed_at").
		Values(ownerID, completedAt).
		Suffix("ON DUPLICATE KEY UPDATE setup_completed_at = VALUES(setup_completed_at), updated_at = NOW()").
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		return nil, err
	}

	_, err = r.pool.ExecContext(ctx, upsertSQL, args...)
	if err != nil {
		return nil, err
	}

	return r.GetHRConfig(ctx, ownerID)
}
