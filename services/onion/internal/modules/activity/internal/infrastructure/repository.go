package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/activity/internal/domain"
	"github.com/voxire/onion/internal/modules/activity/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) ListActivity(ctx context.Context, ownerID int64, limit int) ([]domain.ActivityLog, error) {
	if limit <= 0 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	rows, err := r.queries.ListActivity(ctx, sqlc.ListActivityParams{
		OwnerID: ownerID,
		Limit:   int32(limit),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list activity", logger.Error(err))

		return nil, err
	}

	activities := make([]domain.ActivityLog, 0, len(rows))

	for _, row := range rows {
		var a domain.ActivityLog
		a.ID = row.ID
		a.OwnerID = row.OwnerID
		a.App = row.App
		a.EntityType = row.EntityType
		a.Action = row.Action

		if row.EntityID.Valid {
			a.EntityID = row.EntityID.Int64
		}

		if row.Description.Valid {
			a.Description = row.Description.String
		}

		if row.UserID.Valid {
			a.UserID = row.UserID.Int64
		}

		if row.UserName.Valid {
			a.UserName = row.UserName.String
		}

		if row.CreatedAt.Valid {
			a.CreatedAt = row.CreatedAt.Time
		}

		activities = append(activities, a)
	}

	return activities, nil
}
