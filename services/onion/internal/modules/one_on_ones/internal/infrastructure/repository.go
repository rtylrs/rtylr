package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/voxire/onion/internal/modules/one_on_ones/internal/domain"
	"github.com/voxire/onion/internal/modules/one_on_ones/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.OneOnOne, error) {
	rows, err := r.queries.ListOneOnOnes(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list one-on-ones", logger.Error(err))
		return nil, err
	}

	items := make([]domain.OneOnOne, 0, len(rows))
	for _, row := range rows {
		var item domain.OneOnOne
		item.ID = row.ID
		item.OwnerID = row.OwnerID
		item.ManagerID = row.ManagerID
		item.StaffID = row.StaffID
		item.ScheduledAt = row.ScheduledAt
		if row.Notes.Valid {
			item.Notes = row.Notes.String
		}
		if row.ActionItems != nil {
			item.ActionItems = string(row.ActionItems)
		}
		if row.Status.Valid {
			item.Status = string(row.Status.HrOneOnOnesStatus)
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.OneOnOneInput) (int64, error) {
	status := input.Status
	if status == "" {
		status = "scheduled"
	}

	var actionItems json.RawMessage
	if input.ActionItems != "" {
		actionItems = json.RawMessage(input.ActionItems)
	}

	res, err := r.queries.CreateOneOnOne(ctx, sqlc.CreateOneOnOneParams{
		OwnerID:     ownerID,
		ManagerID:   input.ManagerID,
		StaffID:     input.StaffID,
		ScheduledAt: parseTime(input.ScheduledAt),
		Notes:       sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		ActionItems: actionItems,
		Status:      sqlc.NullHrOneOnOnesStatus{HrOneOnOnesStatus: sqlc.HrOneOnOnesStatus(status), Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create one-on-one", logger.Error(err))
		return 0, err
	}
	return res.LastInsertId()
}

func parseTime(s string) time.Time {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		if parsed, err := time.Parse(layout, s); err == nil {
			return parsed
		}
	}
	return time.Time{}
}
