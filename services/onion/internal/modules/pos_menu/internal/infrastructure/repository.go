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

func (r *Repository) List(ctx context.Context, businessID int64) ([]domain.MenuItem, error) {
	rows, err := r.queries.ListMenuItemsWithDetails(ctx, sqlc.ListMenuItemsWithDetailsParams{
		BusinessID:   businessID,
		BusinessID_2: businessID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list menu items", logger.Error(err))
		return nil, err
	}

	items := make([]domain.MenuItem, 0, len(rows))
	for _, row := range rows {
		item := domain.MenuItem{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
			Name:       row.Name,
			SalePrice:  decimalToFloat(row.SalePrice),
			TaxRate:    decimalToFloat(row.TaxRate),
			Stock:      decimalToFloat(row.StockQuantity),
			IsActive:   row.IsActive.Valid && row.IsActive.Bool,
			Category:   row.CategoryName,
		}
		if row.CategoryID.Valid {
			item.CategoryID = row.CategoryID.Int64
		}
		if row.Sku.Valid {
			item.SKU = row.Sku.String
		}
		if row.Barcode.Valid {
			item.Barcode = row.Barcode.String
		}
		if row.Description.Valid {
			item.Description = row.Description.String
		}
		if row.ImageUrl != "" {
			item.ImageURL = row.ImageUrl
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}
		if row.UpdatedAt.Valid {
			item.UpdatedAt = row.UpdatedAt.Time
		}
		items = append(items, item)
	}
	return items, nil
}

func (r *Repository) QuickCreate(ctx context.Context, ownerID, businessID int64, input domain.MenuItemInput) (int64, error) {
	result, err := r.queries.QuickCreateMenuItem(ctx, sqlc.QuickCreateMenuItemParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		Name:       input.Name,
		Sku:        sql.NullString{String: input.SKU, Valid: input.SKU != ""},
		SalePrice:  floatToDecimal(input.SalePrice),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to quick-create menu item", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.MenuItemUpdateInput) error {
	err := r.queries.UpdateMenuItem(ctx, sqlc.UpdateMenuItemParams{
		Name:          input.Name,
		SalePrice:     floatToDecimal(input.SalePrice),
		StockQuantity: floatToDecimal(input.Stock),
		IsActive:      sql.NullBool{Bool: input.IsActive, Valid: true},
		BusinessID:    businessID,
		ID:            id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update menu item", logger.Error(err))
		return err
	}
	return nil
}

func (r *Repository) BusinessOwnerID(ctx context.Context, businessID int64) (int64, error) {
	ownerID, err := r.queries.GetMenuBusinessOwnerID(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve menu business owner", logger.Error(err))
		return 0, err
	}
	return ownerID, nil
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
