package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/platform/logger"
)

// SettingsRow is the module-owned config for the insights surface.
// Unlike cross-schema reads above, this table lives in the insights
// schema — no annotation comment is needed because there is no
// other module owning it.
type SettingsRow struct {
	OwnerID          int64
	DefaultCurrency  string
	FiscalYearStart  string // "MM-DD"
	Timezone         string
}

// ErrSettingsNotFound matches the sentinel used in every other module.
var ErrSettingsNotFound = errors.New("insights settings not found")

// GetSettings reads the one-row-per-owner settings. Returns
// ErrSettingsNotFound when no row exists; the service layer is expected
// to fall back to a default envelope in that case (see CLAUDE.md §8).
func (r *Repository) GetSettings(ctx context.Context, ownerID int64) (SettingsRow, error) {
	const q = `
		SELECT owner_id, default_currency, fiscal_year_start, timezone
		FROM insights_settings
		WHERE owner_id = ?
	`
	row := r.pool.QueryRowContext(ctx, q, ownerID)
	var out SettingsRow
	if err := row.Scan(&out.OwnerID, &out.DefaultCurrency, &out.FiscalYearStart, &out.Timezone); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return SettingsRow{}, ErrSettingsNotFound
		}
		logger.Ctx(ctx).Error("insights: get settings failed", logger.Error(err))
		return SettingsRow{}, err
	}
	return out, nil
}
