package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/hr_setup/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{
		pool: dbclient.Pool(),
	}
}

func (r *Repository) GetSetupSteps(ctx context.Context, ownerID int64) ([]*domain.SetupStep, error) {
	rows, err := r.pool.QueryContext(ctx,
		"SELECT id, owner_id, step_key, completed_at FROM hr_setup_steps WHERE owner_id = ?",
		ownerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []*domain.SetupStep
	for rows.Next() {
		var s domain.SetupStep
		if err := rows.Scan(&s.ID, &s.OwnerID, &s.StepKey, &s.CompletedAt); err != nil {
			return nil, err
		}
		steps = append(steps, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return steps, nil
}

func (r *Repository) UpsertSetupStep(ctx context.Context, ownerID int64, stepKey string, completedAt time.Time) error {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO hr_setup_steps (owner_id, step_key, completed_at)
		 VALUES (?, ?, ?)
		 ON DUPLICATE KEY UPDATE completed_at = VALUES(completed_at)`,
		ownerID, stepKey, completedAt,
	)

	return err
}

func (r *Repository) GetHRConfig(ctx context.Context, ownerID int64) (*domain.HRConfig, error) {
	row := r.pool.QueryRowContext(ctx,
		"SELECT id, owner_id, setup_completed_at, created_at, updated_at FROM hr_configs WHERE owner_id = ?",
		ownerID,
	)

	var cfg domain.HRConfig
	err := row.Scan(&cfg.ID, &cfg.OwnerID, &cfg.SetupCompletedAt, &cfg.CreatedAt, &cfg.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrConfigNotFound
		}
		return nil, err
	}

	return &cfg, nil
}

func (r *Repository) UpsertHRConfig(ctx context.Context, ownerID int64, completedAt *time.Time) (*domain.HRConfig, error) {
	_, err := r.pool.ExecContext(ctx,
		`INSERT INTO hr_configs (owner_id, setup_completed_at)
		 VALUES (?, ?)
		 ON DUPLICATE KEY UPDATE setup_completed_at = VALUES(setup_completed_at), updated_at = NOW()`,
		ownerID, completedAt,
	)
	if err != nil {
		return nil, err
	}

	return r.GetHRConfig(ctx, ownerID)
}
