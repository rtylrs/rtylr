package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/addons/internal/domain"
	"github.com/voxire/onion/internal/modules/addons/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/money"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{
		queries: sqlc.New(db),
	}
}

func (r *Repository) ListAll(ctx context.Context) ([]domain.Addon, error) {
	rows, err := r.queries.ListAllAddons(ctx)
	if err != nil {
		return nil, err
	}

	addons := make([]domain.Addon, 0, len(rows))
	for _, row := range rows {
		addons = append(addons, domain.Addon{
			ID:               row.ID,
			Name:             row.Name,
			Description:      row.Description.String,
			Icon:             row.Icon.String,
			Price:            parseDecimal(row.Price.String),
			IsFree:           row.IsFree.Bool,
			IsDefaultEnabled: row.IsDefaultEnabled.Bool,
			SortOrder:        int(row.SortOrder.Int32),
		})
	}
	return addons, nil
}

func (r *Repository) ListForBusiness(ctx context.Context, businessID int64) ([]domain.BusinessAddon, error) {
	rows, err := r.queries.ListBusinessAddons(ctx, businessID)
	if err != nil {
		return nil, err
	}

	addons := make([]domain.BusinessAddon, 0, len(rows))
	for _, row := range rows {
		ba := domain.BusinessAddon{
			AddonID:     row.AddonID,
			Name:        row.Name,
			Description: row.Description.String,
			Icon:        row.Icon.String,
			Price:       parseDecimal(row.Price.String),
			IsFree:      row.IsFree.Bool,
			IsEnabled:   row.IsEnabled.Bool,
		}
		if row.EnabledAt.Valid {
			ba.EnabledAt = &row.EnabledAt.Time
		}
		if row.DisabledAt.Valid {
			ba.DisabledAt = &row.DisabledAt.Time
		}
		addons = append(addons, ba)
	}
	return addons, nil
}

func (r *Repository) HasActiveGrant(ctx context.Context, businessID int64, addonID string) (bool, error) {
	count, err := r.queries.HasActiveAddonGrant(ctx, sqlc.HasActiveAddonGrantParams{
		BusinessID: businessID,
		AddonID:    addonID,
	})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) Enable(ctx context.Context, businessID int64, addonID string) error {
	return r.queries.EnableAddon(ctx, sqlc.EnableAddonParams{
		BusinessID: businessID,
		AddonID:    addonID,
	})
}

func (r *Repository) Disable(ctx context.Context, businessID int64, addonID string) error {
	return r.queries.DisableAddon(ctx, sqlc.DisableAddonParams{
		BusinessID: businessID,
		AddonID:    addonID,
	})
}

// parseDecimal converts a decimal string from the database into exact minor units.
// On a corrupt (non-numeric) value it logs at Warn and returns 0 rather than
// silently discarding the error — see CLAUDE.md §4 Error Handling Pattern.
func parseDecimal(s string) money.Amount {
	if s == "" {
		return 0
	}
	v, err := money.FromString(s)
	if err != nil {
		logger.Global().Warn("failed to parse decimal value from database",
			logger.String("value", s),
			logger.Error(err),
		)
		return 0
	}
	return v
}
