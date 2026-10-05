package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/activity/internal/domain"
	"github.com/voxire/onion/internal/modules/activity/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries  *sqlc.Queries
	userName func(context.Context, int64, int64) (string, error)
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

// ListActivity returns one page of activity for a business (optionally filtered
// to a single app — pass "" for all apps) plus the total count for that same
// filter, so callers can paginate. limit is clamped to [1,200]; offset is floored at 0.
func (r *Repository) ListActivity(ctx context.Context, businessID int64, app string, limit, offset int) ([]domain.ActivityLog, int64, error) {
	if limit <= 0 {
		limit = 50
	}

	if limit > 200 {
		limit = 200
	}

	if offset < 0 {
		offset = 0
	}

	rows, err := r.queries.ListActivity(ctx, sqlc.ListActivityParams{
		BusinessID: businessID,
		Column2:    app, // matches the `? = ''` short-circuit; empty app = all apps
		App:        app,
		Limit:      int32(limit),
		Offset:     int32(offset),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list activity", logger.Error(err))

		return nil, 0, err
	}

	activities := make([]domain.ActivityLog, 0, len(rows))

	for _, row := range rows {
		var a domain.ActivityLog
		a.ID = row.ID
		a.OwnerID = row.OwnerID
		a.BusinessID = row.BusinessID
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

		if a.UserID > 0 && r.userName != nil {
			name, err := r.userName(ctx, businessID, a.UserID)
			if err != nil {
				return nil, 0, err
			}
			a.UserName = name
		}

		if row.CreatedAt.Valid {
			a.CreatedAt = row.CreatedAt.Time
		}

		activities = append(activities, a)
	}

	total, err := r.queries.CountActivity(ctx, sqlc.CountActivityParams{
		BusinessID: businessID,
		Column2:    app,
		App:        app,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count activity", logger.Error(err))

		return nil, 0, err
	}

	return activities, total, nil
}

func (r *Repository) WithUserNameReader(reader func(context.Context, int64, int64) (string, error)) *Repository {
	r.userName = reader
	return r
}
