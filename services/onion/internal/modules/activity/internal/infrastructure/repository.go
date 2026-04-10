package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/activity/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		pool: pool,
	}
}

func (r *Repository) ListActivity(ctx context.Context, ownerID int64, limit int) ([]domain.ActivityLog, error) {
	if limit <= 0 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	query := `
		SELECT id, owner_id, entity_type, entity_id, action, description, user_id, created_at
		FROM erp_activity_log
		WHERE owner_id = ?
		ORDER BY created_at DESC
		LIMIT ?`

	rows, err := r.pool.QueryContext(ctx, query, ownerID, limit)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list activity", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var activities []domain.ActivityLog

	for rows.Next() {
		var a domain.ActivityLog
		if err := a.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan activity row", logger.Error(err))

			return nil, err
		}

		activities = append(activities, a)
	}

	return activities, rows.Err()
}
