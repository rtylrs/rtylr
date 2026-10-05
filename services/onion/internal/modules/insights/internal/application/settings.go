package application

import (
	"context"
	"errors"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

type SettingsPayload struct {
	DefaultPeriod    string   `json:"default_period"`
	DefaultCurrency  string   `json:"default_currency"`
	ComparePrevious  bool     `json:"compare_previous"`
	DashboardWidgets []string `json:"dashboard_widgets"`
	FavoriteReports  []string `json:"favorite_reports"`
	FiscalYearStart  string   `json:"fiscal_year_start"`
	Timezone         string   `json:"timezone"`
}

func defaultSettings() SettingsPayload {
	return SettingsPayload{
		DefaultPeriod: "this_month", DefaultCurrency: "USD", ComparePrevious: false,
		DashboardWidgets: []string{}, FavoriteReports: []string{},
		FiscalYearStart: "01-01", Timezone: "UTC",
	}
}

func (s *Service) GetSettings(ctx context.Context, businessID int64) (SettingsPayload, error) {
	row, err := s.repo.GetSettings(ctx, businessID)
	if err != nil {
		if errors.Is(err, domain.ErrSettingsNotFound) {
			logger.Ctx(ctx).Debug("insights: no settings row, returning defaults")
			return defaultSettings(), nil
		}
		return SettingsPayload{}, err
	}
	return SettingsPayload{
		DefaultPeriod: row.DefaultPeriod, DefaultCurrency: row.DefaultCurrency, ComparePrevious: row.ComparePrevious,
		DashboardWidgets: row.DashboardWidgets, FavoriteReports: row.FavoriteReports,
		FiscalYearStart: row.FiscalYearStart, Timezone: row.Timezone,
	}, nil
}

func (s *Service) UpdateSettings(ctx context.Context, businessID int64, in SettingsPayload) (SettingsPayload, error) {
	defaults := defaultSettings()
	if in.DefaultPeriod == "" {
		in.DefaultPeriod = defaults.DefaultPeriod
	}
	if in.DefaultCurrency == "" {
		in.DefaultCurrency = defaults.DefaultCurrency
	}
	if in.DashboardWidgets == nil {
		in.DashboardWidgets = []string{}
	}
	if in.FavoriteReports == nil {
		in.FavoriteReports = []string{}
	}
	if in.FiscalYearStart == "" {
		in.FiscalYearStart = defaults.FiscalYearStart
	}
	if in.Timezone == "" {
		in.Timezone = defaults.Timezone
	}
	err := s.repo.UpsertSettings(ctx, domain.SettingsRow{
		OwnerID: businessID, BusinessID: businessID,
		DefaultCurrency: in.DefaultCurrency, FiscalYearStart: in.FiscalYearStart, Timezone: in.Timezone,
		DefaultPeriod: in.DefaultPeriod, ComparePrevious: in.ComparePrevious,
		DashboardWidgets: in.DashboardWidgets, FavoriteReports: in.FavoriteReports,
	})
	if err != nil {
		return SettingsPayload{}, err
	}
	return in, nil
}
