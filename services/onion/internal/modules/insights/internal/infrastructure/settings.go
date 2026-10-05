package infrastructure

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/voxire/onion/internal/modules/insights/internal/domain"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type SettingsRow = domain.SettingsRow

var ErrSettingsNotFound = domain.ErrSettingsNotFound

func decodeStringList(value json.RawMessage) ([]string, error) {
	if len(value) == 0 || string(value) == "null" {
		return []string{}, nil
	}
	var out []string
	if err := json.Unmarshal(value, &out); err != nil {
		return nil, err
	}
	if out == nil {
		out = []string{}
	}
	return out, nil
}

func encodeStringList(values []string) json.RawMessage {
	if values == nil {
		values = []string{}
	}
	encoded, _ := json.Marshal(values)
	return encoded
}

func (r *Repository) GetSettings(ctx context.Context, businessID int64) (SettingsRow, error) {
	row, err := r.queries.GetSettings(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SettingsRow{}, ErrSettingsNotFound
		}
		logger.Ctx(ctx).Error("insights: get settings failed", logger.Error(err))
		return SettingsRow{}, err
	}
	widgets, err := decodeStringList(row.DashboardWidgets)
	if err != nil {
		return SettingsRow{}, err
	}
	reports, err := decodeStringList(row.FavoriteReports)
	if err != nil {
		return SettingsRow{}, err
	}
	return SettingsRow{
		OwnerID: row.OwnerID, BusinessID: row.BusinessID,
		DefaultCurrency: row.DefaultCurrency, FiscalYearStart: row.FiscalYearStart, Timezone: row.Timezone,
		DefaultPeriod: row.DefaultPeriod, ComparePrevious: row.ComparePrevious,
		DashboardWidgets: widgets, FavoriteReports: reports,
	}, nil
}

func (r *Repository) UpsertSettings(ctx context.Context, row SettingsRow) error {
	err := r.queries.UpsertSettings(ctx, sqlc.UpsertSettingsParams{
		OwnerID: row.OwnerID, BusinessID: row.BusinessID,
		DefaultCurrency: row.DefaultCurrency, FiscalYearStart: row.FiscalYearStart, Timezone: row.Timezone,
		DefaultPeriod: row.DefaultPeriod, ComparePrevious: row.ComparePrevious,
		DashboardWidgets: encodeStringList(row.DashboardWidgets), FavoriteReports: encodeStringList(row.FavoriteReports),
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: upsert settings failed", logger.Any("business_id", row.BusinessID), logger.Error(err))
		return err
	}
	return nil
}
