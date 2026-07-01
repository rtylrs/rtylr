package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type SettingsRow struct {
	OwnerID         int64
	BusinessID      int64
	DefaultCurrency string
	FiscalYearStart string
	Timezone        string
}

var ErrSettingsNotFound = errors.New("insights settings not found")

func (r *Repository) GetSettings(ctx context.Context, businessID int64) (SettingsRow, error) {
	row, err := r.queries.GetSettings(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SettingsRow{}, ErrSettingsNotFound
		}
		logger.Ctx(ctx).Error("insights: get settings failed", logger.Error(err))
		return SettingsRow{}, err
	}
	return SettingsRow{
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		DefaultCurrency: row.DefaultCurrency,
		FiscalYearStart: row.FiscalYearStart,
		Timezone:        row.Timezone,
	}, nil
}

func (r *Repository) UpsertSettings(ctx context.Context, row SettingsRow) error {
	err := r.queries.UpsertSettings(ctx, sqlc.UpsertSettingsParams{
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		DefaultCurrency: row.DefaultCurrency,
		FiscalYearStart: row.FiscalYearStart,
		Timezone:        row.Timezone,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: upsert settings failed",
			logger.Any("business_id", row.BusinessID),
			logger.Error(err),
		)
		return err
	}
	return nil
}
