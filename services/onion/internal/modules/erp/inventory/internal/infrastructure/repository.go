package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"strconv"

	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/inventory/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

	// Verify tenancy so a movement can never create stock rows for another
	// business's catalog entry. No row lock here: ApplyMovementTx locks the
	// stock level before the product, and taking the product first would
	// invert that order and deadlock against concurrent sales.
	var ownedProductID int64
	err = tx.QueryRowContext(ctx, `
		SELECT id
		FROM erp_products
		WHERE id = ? AND owner_id = ? AND business_id = ?`, productID, ownerID, ownerID).Scan(&ownedProductID)
	if errors.Is(err, sql.ErrNoRows) {
		return sql.ErrNoRows
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to verify inventory product ownership", logger.Error(err))

		return err
	}

	// Route the stock change through the shared stock_core primitive so it lands
	// in stock_levels + stock_ledger and keeps erp_products.stock_quantity (the
	// legacy scalar) synced as the owner-wide total. ERP adjustments target the
	// tenant's default warehouse location.
	locID, err := stock.DefaultLocationID(ctx, tx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to resolve default location", logger.Error(err))

		return err
	}

	before, after, err := stock.ApplyMovementTx(ctx, tx, stock.Movement{
		OwnerID:    ownerID,
		LocationID: locID,
		ProductID:  productID,
		EntryType:  stock.EntryType(txType),
		Change:     change,
		Channel:    "erp",
		Note:       notes,
		CreatedBy:  &ownerID,
	})
	if errors.Is(err, stock.ErrInsufficientStock) {
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
		Notes:          sqlutil.NullString(notes),
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
	page, err := r.GetInventoryReportPage(ctx, ownerID, 100, 0)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (r *Repository) GetInventoryReportPage(ctx context.Context, ownerID int64, limit, offset int) (domain.InventoryReportPage, error) {
	rows, err := r.queries.GetInventoryReport(ctx, sqlc.GetInventoryReportParams{
		BusinessID: ownerID,
		Limit:      int32(limit),  // #nosec G115 -- pagination is normalized by the application service.
		Offset:     int32(offset), // #nosec G115 -- pagination is normalized by the application service.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get inventory report", logger.Error(err))
		return domain.InventoryReportPage{}, err
	}

	summary, err := r.queries.GetInventoryReportSummary(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to summarize inventory report", logger.Error(err))
		return domain.InventoryReportPage{}, err
	}

	return domain.InventoryReportPage{
		Items:      inventoryReportItems(rows),
		Total:      summary.Total,
		TotalValue: sqlutil.DecimalToFloat(summary.TotalValue),
		HasMore:    int64(len(rows))+int64(offset) < summary.Total,
		Limit:      int64(limit),
		Offset:     int64(offset),
	}, nil
}

func inventoryReportItems(rows []sqlc.GetInventoryReportRow) []domain.InventoryReportItem {
	report := make([]domain.InventoryReportItem, 0, len(rows))
	for _, row := range rows {
		report = append(report, inventoryReportItem(row.ID, row.Name, row.Sku, row.Unit, row.StockQuantity, row.LowStockThreshold, row.ReorderQuantity, row.CostPrice, row.SalePrice, row.StockValue, row.IsLowStock))
	}
	return report
}

func (r *Repository) GetLowStockProducts(ctx context.Context, ownerID int64) ([]domain.InventoryReportItem, error) {
	page, err := r.GetLowStockProductsPage(ctx, ownerID, 100, 0)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (r *Repository) GetLowStockProductsPage(ctx context.Context, ownerID int64, limit, offset int) (domain.LowStockPage, error) {
	rows, err := r.queries.GetLowStockProducts(ctx, sqlc.GetLowStockProductsParams{
		BusinessID: ownerID,
		Limit:      int32(limit),  // #nosec G115 -- pagination is normalized by the application service.
		Offset:     int32(offset), // #nosec G115 -- pagination is normalized by the application service.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get low stock products", logger.Error(err))
		return domain.LowStockPage{}, err
	}

	summary, err := r.queries.CountLowStockProducts(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to count low stock products", logger.Error(err))
		return domain.LowStockPage{}, err
	}

	items := make([]domain.InventoryReportItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, inventoryReportItem(row.ID, row.Name, row.Sku, row.Unit, row.StockQuantity, row.LowStockThreshold, row.ReorderQuantity, row.CostPrice, row.SalePrice, row.StockValue, row.IsLowStock))
	}

	return domain.LowStockPage{
		Items:   items,
		Total:   summary,
		HasMore: int64(len(rows))+int64(offset) < summary,
		Limit:   int64(limit),
		Offset:  int64(offset),
	}, nil
}

func inventoryReportItem(productID int64, name string, sku, unit sql.NullString, stockQuantity string, lowStockThreshold, reorderQuantity sql.NullString, costPrice, salePrice, stockValue string, isLowStock int32) domain.InventoryReportItem {
	item := domain.InventoryReportItem{
		ProductID:     productID,
		ProductName:   name,
		StockQuantity: sqlutil.ParseDecimal(stockQuantity),
		CostPrice:     sqlutil.ParseDecimal(costPrice),
		SalePrice:     sqlutil.ParseDecimal(salePrice),
		StockValue:    sqlutil.ParseDecimal(stockValue),
		IsLowStock:    isLowStock != 0,
	}
	if sku.Valid {
		item.SKU = sku.String
	}
	if unit.Valid {
		item.Unit = unit.String
	}
	if lowStockThreshold.Valid {
		item.LowStockThreshold = sqlutil.ParseDecimal(lowStockThreshold.String)
	}
	if reorderQuantity.Valid {
		item.ReorderQuantity = sqlutil.ParseDecimal(reorderQuantity.String)
	}
	return item
}

func (r *Repository) GetValuation(ctx context.Context, ownerID int64) ([]domain.ValuationItem, error) {
	page, err := r.GetValuationPage(ctx, ownerID, 100, 0)
	if err != nil {
		return nil, err
	}
	return page.Items, nil
}

func (r *Repository) GetValuationPage(ctx context.Context, ownerID int64, limit, offset int) (domain.ValuationPage, error) {
	rows, err := r.queries.GetValuation(ctx, sqlc.GetValuationParams{
		BusinessID:   ownerID,
		BusinessID_2: ownerID,
		Limit:        int32(limit),  // #nosec G115 -- pagination is normalized by the application service.
		Offset:       int32(offset), // #nosec G115 -- pagination is normalized by the application service.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get inventory valuation", logger.Error(err))
		return domain.ValuationPage{}, err
	}

	summary, err := r.queries.GetValuationSummary(ctx, sqlc.GetValuationSummaryParams{
		BusinessID:   ownerID,
		BusinessID_2: ownerID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to summarize inventory valuation", logger.Error(err))
		return domain.ValuationPage{}, err
	}

	items := make([]domain.ValuationItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.ValuationItem{
			CategoryName:  row.CategoryName,
			WarehouseName: row.WarehouseName,
			TotalItems:    row.TotalItems,
			TotalValue:    sqlutil.DecimalToFloat(row.TotalValue),
			TotalCost:     sqlutil.DecimalToFloat(row.TotalCost),
		})
	}

	return domain.ValuationPage{
		Items:      items,
		Total:      summary.Total,
		TotalValue: sqlutil.DecimalToFloat(summary.TotalValue),
		HasMore:    int64(len(rows))+int64(offset) < summary.Total,
		Limit:      int64(limit),
		Offset:     int64(offset),
	}, nil
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
