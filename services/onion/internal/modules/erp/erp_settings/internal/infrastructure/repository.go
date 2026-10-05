package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/erp/erp_settings/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/erp_settings/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{queries: sqlc.New(db)}
}

func (r *Repository) Get(ctx context.Context, businessID int64) (domain.Settings, error) {
	row, err := r.queries.GetErpSettings(ctx, businessID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Settings{}, domain.ErrSettingsNotFound
		}
		logger.Ctx(ctx).Error("Failed to get ERP settings", logger.Error(err))
		return domain.Settings{}, err
	}

	return domain.Settings{
		ID:                 row.ID,
		BusinessID:         row.BusinessID,
		Currency:           row.Currency,
		TaxRate:            row.TaxRate,
		LowStockThreshold:  row.LowStockThreshold,
		DefaultWarehouseID: row.DefaultWarehouseID,
		NotifyOnLowStock:   row.NotifyOnLowStock,
		NotifyOnExpiry:     row.NotifyOnExpiry,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}, nil
}

func (r *Repository) Upsert(ctx context.Context, input domain.SettingsInput) error {
	err := r.queries.UpsertErpSettings(ctx, sqlc.UpsertErpSettingsParams{
		BusinessID:         input.BusinessID,
		Currency:           input.Currency,
		TaxRate:            input.TaxRate,
		LowStockThreshold:  input.LowStockThreshold,
		DefaultWarehouseID: input.DefaultWarehouseID,
		NotifyOnLowStock:   input.NotifyOnLowStock,
		NotifyOnExpiry:     input.NotifyOnExpiry,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to upsert ERP settings", logger.Error(err))
	}
	return err
}
