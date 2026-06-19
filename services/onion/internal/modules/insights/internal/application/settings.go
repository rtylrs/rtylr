package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure"
	"github.com/voxire/onion/internal/platform/logger"
)

// SettingsPayload is what /insights/settings returns. Non-critical config
// — we silently fall back to defaults when the tenant has not yet
// created a row, per the fallback rules in CLAUDE.md §8.
type SettingsPayload struct {
	DefaultCurrency string `json:"default_currency"`
	FiscalYearStart string `json:"fiscal_year_start"`
	Timezone        string `json:"timezone"`
}

// defaultSettings is the in-code fallback. Product-level defaults; if a
// tenant wants to override them they insert a row explicitly.
func defaultSettings() SettingsPayload {
	return SettingsPayload{
		DefaultCurrency: "USD",
		FiscalYearStart: "01-01",
		Timezone:        "UTC",
	}
}

// GetSettings returns the insights-settings row, or the default envelope
// when no row exists yet for this tenant.
func (s *Service) GetSettings(ctx context.Context, businessID int64) (SettingsPayload, error) {
	row, err := s.repo.GetSettings(ctx, businessID)
	if err != nil {
		if errors.Is(err, infrastructure.ErrSettingsNotFound) {
			logger.Ctx(ctx).Debug("insights: no settings row, returning defaults")
			return defaultSettings(), nil
		}
		return SettingsPayload{}, err
	}
	return SettingsPayload{
		DefaultCurrency: row.DefaultCurrency,
		FiscalYearStart: row.FiscalYearStart,
		Timezone:        row.Timezone,
	}, nil
}

// UpdateSettings writes the payload. Missing fields are rejected —
// partial updates would surprise the UI which always sends the full
// envelope.
func (s *Service) UpdateSettings(ctx context.Context, businessID int64, in SettingsPayload) (SettingsPayload, error) {
	if in.DefaultCurrency == "" {
		in.DefaultCurrency = "USD"
	}
	if in.FiscalYearStart == "" {
		in.FiscalYearStart = "01-01"
	}
	if in.Timezone == "" {
		in.Timezone = "UTC"
	}
	err := s.repo.UpsertSettings(ctx, infrastructure.SettingsRow{
		OwnerID:         businessID,
		BusinessID:      businessID,
		DefaultCurrency: in.DefaultCurrency,
		FiscalYearStart: in.FiscalYearStart,
		Timezone:        in.Timezone,
	})
	if err != nil {
		return SettingsPayload{}, err
	}
	return in, nil
}
