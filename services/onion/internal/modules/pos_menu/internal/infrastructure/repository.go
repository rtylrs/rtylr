package infrastructure

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/voxire/onion/internal/modules/pos_menu/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_menu/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
	"github.com/voxire/onion/internal/platform/stock_core"
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
			SalePrice:  sqlutil.DecimalToFloat(row.SalePrice),
			CostPrice:  sqlutil.DecimalToFloat(row.CostPrice),
			TaxRate:    sqlutil.DecimalToFloat(row.TaxRate),
			Stock:      stockToFloat(row.StockQuantity),
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

func (r *Repository) IsOnlineOrderingEnabled(ctx context.Context, businessID int64) (bool, error) {
	var enabled int
	err := r.pool.QueryRowContext(ctx,
		"SELECT COUNT(*) FROM business_addons WHERE business_id = ? AND addon_id = 'online_ordering' AND is_enabled = TRUE",
		businessID,
	).Scan(&enabled)
	return enabled > 0, err
}

func (r *Repository) QuickCreate(ctx context.Context, ownerID, businessID int64, input domain.MenuItemInput) (int64, error) {
	result, err := r.queries.QuickCreateMenuItem(ctx, sqlc.QuickCreateMenuItemParams{
		OwnerID:    ownerID,
		BusinessID: businessID,
		Name:       input.Name,
		Sku:        sql.NullString{String: input.SKU, Valid: input.SKU != ""},
		SalePrice:  sqlutil.FloatToDecimal(input.SalePrice),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to quick-create menu item", logger.Error(err))
		return 0, err
	}
	return result.LastInsertId()
}

func (r *Repository) Update(ctx context.Context, businessID, id int64, input domain.MenuItemUpdateInput) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var ownerID int64
	if err := tx.QueryRowContext(ctx,
		"SELECT owner_id FROM businesses WHERE id = ? AND is_active = TRUE",
		businessID,
	).Scan(&ownerID); err != nil {
		return err
	}

	locationID, err := stock_core.DefaultLocationID(ctx, tx, ownerID)
	if err != nil {
		return err
	}
	if _, _, err := stock_core.SetTotalTx(ctx, tx, stock_core.Movement{
		OwnerID:       ownerID,
		BusinessID:    businessID,
		LocationID:    locationID,
		ProductID:     id,
		EntryType:     stock_core.EntryAdjustment,
		ReferenceType: "pos_menu",
		Channel:       "pos",
		Note:          "POS menu stock edit",
	}, input.Stock); err != nil {
		return err
	}

	qtx := r.queries.WithTx(tx)
	result, err := qtx.UpdateMenuItem(ctx, sqlc.UpdateMenuItemParams{
		Name:       input.Name,
		SalePrice:  sqlutil.FloatToDecimal(input.SalePrice),
		IsActive:   sql.NullBool{Bool: input.IsActive, Valid: true},
		BusinessID: businessID,
		ID:         id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update menu item", logger.Error(err))
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

func (r *Repository) BusinessOwnerID(ctx context.Context, businessID int64) (int64, error) {
	ownerID, err := r.queries.GetMenuBusinessOwnerID(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve menu business owner", logger.Error(err))
		return 0, err
	}
	return ownerID, nil
}

// stockToFloat converts the value returned by the COALESCE(pst.total_quantity, p.stock_quantity)
// aggregate (sqlc emits interface{} for a DECIMAL sourced from the product_stock_total view) into
// a float64, handling the driver representations MySQL may return for a DECIMAL column.
func stockToFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case []byte:
		return sqlutil.DecimalToFloat(string(val))
	case string:
		return sqlutil.DecimalToFloat(val)
	default:
		return sqlutil.DecimalToFloat(fmt.Sprintf("%v", val))
	}
}
