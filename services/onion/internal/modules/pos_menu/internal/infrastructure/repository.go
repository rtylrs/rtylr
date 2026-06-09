package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	pool := db
	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.MenuItem, error) {
	rows, err := r.queries.ListMenuItems(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list menu items", logger.Error(err))
		return nil, err
	}

	items := make([]domain.MenuItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.MenuItem{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			CategoryID: row.CategoryID.Int64,
			Name:       row.Name,
			SKU:        row.Sku.String,
			SalePrice:  decimalToFloat(row.SalePrice),
			IsActive:   row.IsActive.Bool,
			CreatedAt:  row.CreatedAt.Time,
			UpdatedAt:  row.UpdatedAt.Time,
		})
	}
	return items, nil
}

func (r *Repository) QuickCreate(ctx context.Context, ownerID int64, input domain.MenuItemInput) (int64, error) {
	result, err := r.queries.QuickCreateMenuItem(ctx, sqlc.QuickCreateMenuItemParams{
		OwnerID:   ownerID,
		Name:      input.Name,
		Sku:       sql.NullString{String: input.SKU, Valid: input.SKU != ""},
		SalePrice: floatToDecimal(input.SalePrice),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to quick-create menu item", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func decimalToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse menu decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}

func floatToDecimal(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}
