package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/crm_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/crm_settings/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) Get(ctx context.Context, businessID int64) (domain.Settings, error) {
	row, err := r.queries.GetSettings(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Settings{
				OwnerID:        businessID,
				BusinessID:     businessID,
				DefaultConsent: true,
				Timezone:       "UTC",
			}, nil
		}

		logger.Ctx(ctx).Error("Failed to get CRM settings", logger.Error(err))

		return domain.Settings{}, err
	}

	s := domain.Settings{
		ID:             row.ID,
		OwnerID:        row.OwnerID,
		BusinessID:     row.BusinessID,
		DefaultConsent: row.DefaultConsent.Bool,
		FromEmail:      row.FromEmail.String,
		FromName:       row.FromName.String,
		SmsSenderID:    row.SmsSenderID.String,
		Timezone:       row.Timezone.String,
		CreatedAt:      row.CreatedAt.Time,
		UpdatedAt:      row.UpdatedAt.Time,
	}

	if s.Timezone == "" {
		s.Timezone = "UTC"
	}

	return s, nil
}

func (r *Repository) Upsert(ctx context.Context, input domain.SettingsInput) error {
	err := r.queries.UpsertSettings(ctx, sqlc.UpsertSettingsParams{
		OwnerID:        input.OwnerID,
		BusinessID:     input.BusinessID,
		DefaultConsent: sql.NullBool{Bool: input.DefaultConsent, Valid: true},
		FromEmail:      sql.NullString{String: input.FromEmail, Valid: input.FromEmail != ""},
		FromName:       sql.NullString{String: input.FromName, Valid: input.FromName != ""},
		SmsSenderID:    sql.NullString{String: input.SmsSenderID, Valid: input.SmsSenderID != ""},
		Timezone:       sql.NullString{String: input.Timezone, Valid: input.Timezone != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert CRM settings", logger.Error(err))
	}

	return err
}
