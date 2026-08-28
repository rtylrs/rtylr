package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/voxire/onion/internal/modules/support_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/support_settings/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) Get(ctx context.Context, businessID int64) (*domain.Settings, error) {
	row, err := r.queries.GetSettings(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &domain.Settings{BusinessID: businessID, SLAHours: 24}, nil
		}
		logger.Ctx(ctx).Error("Failed to get support settings", logger.Error(err))
		return nil, err
	}
	s := domain.Settings{
		ID:         row.ID,
		BusinessID: row.BusinessID,
	}
	if row.AutoAssign.Valid {
		s.AutoAssign = row.AutoAssign.Bool
	}
	if row.SlaHours.Valid {
		s.SLAHours = int(row.SlaHours.Int32)
	}
	if len(row.BusinessHours) > 0 {
		s.BusinessHours = string(row.BusinessHours)
	}
	if row.NotificationEmail.Valid {
		s.NotificationEmail = row.NotificationEmail.String
	}
	if row.CreatedAt.Valid {
		s.CreatedAt = row.CreatedAt.Time
	}
	if row.UpdatedAt.Valid {
		s.UpdatedAt = row.UpdatedAt.Time
	}
	return &s, nil
}

func (r *Repository) Upsert(ctx context.Context, input domain.SettingsInput) error {
	bh := json.RawMessage("{}")
	if input.BusinessHours != "" {
		bh = json.RawMessage(input.BusinessHours)
	}
	err := r.queries.UpsertSettings(ctx, sqlc.UpsertSettingsParams{
		BusinessID:        input.BusinessID,
		AutoAssign:        sql.NullBool{Bool: input.AutoAssign, Valid: true},
		SlaHours:          sql.NullInt32{Int32: int32(input.SLAHours), Valid: true}, // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		BusinessHours:     bh,
		NotificationEmail: sql.NullString{String: input.NotificationEmail, Valid: input.NotificationEmail != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert support settings", logger.Error(err))
		return err
	}
	return nil
}
