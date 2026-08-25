package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	"github.com/voxire/onion/internal/modules/inventory/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
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

func (r *Repository) AdjustInventory(ctx context.Context, ownerID, productID int64, change float64, txType domain.InventoryTransactionType, notes string) error {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin inventory transaction", logger.Error(err))

		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	// Route the stock change through the shared stock_core primitive so it lands
	// in stock_levels + stock_ledger and keeps erp_products.stock_quantity (the
	// legacy scalar) synced as the owner-wide total. ERP adjustments target the
	// tenant's default warehouse location.
	locID, err := stock_core.DefaultLocationID(ctx, tx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve default location", logger.Error(err))

		return err
	}

	before, after, err := stock_core.ApplyMovementTx(ctx, tx, stock_core.Movement{
		OwnerID:    ownerID,
		LocationID: locID,
		ProductID:  productID,
		EntryType:  stock_core.EntryType(txType),
		Change:     change,
		Channel:    "erp",
		Note:       notes,
		CreatedBy:  &ownerID,
	})
	if errors.Is(err, stock_core.ErrInsufficientStock) {
		err = domain.ErrInsufficientStock
		logger.Ctx(ctx).Warn("Rejected inventory adjustment that would go negative",
			logger.Any("owner_id", ownerID), logger.Any("product_id", productID))

		return err
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to apply inventory movement", logger.Error(err))

		return err
	}

	// Legacy ledger row for read-compatibility, mirroring the movement.
	_, err = qtx.InsertInventoryTransaction(ctx, sqlc.InsertInventoryTransactionParams{
		OwnerID:        ownerID,
		BusinessID:     ownerID,
		ProductID:      productID,
		Type:           sqlc.ErpInventoryTransactionsType(txType),
		QuantityChange: strconv.FormatFloat(change, 'f', -1, 64),
		QuantityBefore: strconv.FormatFloat(before, 'f', -1, 64),
		QuantityAfter:  strconv.FormatFloat(after, 'f', -1, 64),
		Notes:          nullString(notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to insert inventory transaction", logger.Error(err))

		return err
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit inventory transaction", logger.Error(err))

		return err
	}

	return nil
}

func (r *Repository) GetInventoryReport(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	rows, err := r.queries.GetInventoryReport(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get inventory report", logger.Error(err))

		return nil, err
	}

	report := make([]domain.InventoryReportItem, 0, len(rows))

	for _, row := range rows {
		var item domain.InventoryReportItem
		item.ProductID = row.ID
		item.ProductName = row.Name

		if row.Sku.Valid {
			item.SKU = row.Sku.String
		}

		if row.Unit.Valid {
			item.Unit = row.Unit.String
		}

		item.StockQuantity = parseDecimal(row.StockQuantity)
		item.CostPrice = parseDecimal(row.CostPrice)
		item.SalePrice = parseDecimal(row.SalePrice)
		item.StockValue = parseDecimal(row.StockValue)
		item.IsLowStock = row.IsLowStock != 0

		if row.LowStockThreshold.Valid {
			item.LowStockThreshold = parseDecimal(row.LowStockThreshold.String)
		}

		if row.ReorderQuantity.Valid {
			item.ReorderQuantity = parseDecimal(row.ReorderQuantity.String)
		}

		report = append(report, item)
	}

	return report, nil
}

func (r *Repository) GetLowStockProducts(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	rows, err := r.queries.GetLowStockProducts(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get low stock products", logger.Error(err))

		return nil, err
	}

	report := make([]domain.InventoryReportItem, 0, len(rows))

	for _, row := range rows {
		var item domain.InventoryReportItem
		item.ProductID = row.ID
		item.ProductName = row.Name

		if row.Sku.Valid {
			item.SKU = row.Sku.String
		}

		if row.Unit.Valid {
			item.Unit = row.Unit.String
		}

		item.StockQuantity = parseDecimal(row.StockQuantity)
		item.CostPrice = parseDecimal(row.CostPrice)
		item.SalePrice = parseDecimal(row.SalePrice)
		item.StockValue = parseDecimal(row.StockValue)
		item.IsLowStock = row.IsLowStock != 0

		if row.LowStockThreshold.Valid {
			item.LowStockThreshold = parseDecimal(row.LowStockThreshold.String)
		}

		if row.ReorderQuantity.Valid {
			item.ReorderQuantity = parseDecimal(row.ReorderQuantity.String)
		}

		report = append(report, item)
	}

	return report, nil
}

func (r *Repository) GetValuation(ctx context.Context, ownerID int64) ([]domain.ValuationItem, error) {
	rows, err := r.queries.GetValuation(ctx, sqlc.GetValuationParams{
		BusinessID:   ownerID,
		BusinessID_2: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get inventory valuation", logger.Error(err))

		return nil, err
	}

	items := make([]domain.ValuationItem, 0, len(rows))

	for _, row := range rows {
		var item domain.ValuationItem
		item.CategoryName = row.CategoryName
		item.WarehouseName = row.WarehouseName
		item.TotalItems = row.TotalItems

		if row.TotalValue != nil {
			switch v := row.TotalValue.(type) {
			case float64:
				item.TotalValue = v
			case []byte:
				item.TotalValue = parseDecimal(string(v))
			}
		}

		if row.TotalCost != nil {
			switch v := row.TotalCost.(type) {
			case float64:
				item.TotalCost = v
			case []byte:
				item.TotalCost = parseDecimal(string(v))
			}
		}

		items = append(items, item)
	}

	return items, nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
func parseDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse inventory decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}
