package infrastructure

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"golang.org/x/sync/errgroup"
)

// GetDashboard aggregates the counts + recent records used by the ERP
// dashboard. Counts run as three small queries rather than one UNION
// because MySQL optimises simple COUNT(*) scans extremely well.
func (r *Repository) GetDashboard(ctx context.Context, ownerID int64) (domain.DashboardStats, error) {
	var (
		totalProducts  int64
		lowStock       int64
		pendingOrders  int64
		totalSuppliers int64
		invValue       float64
		items          []domain.LowStockItem
		recent         []domain.RecentPO
		cats           []domain.TopCategory
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		totalProducts, err = r.queries.CountActiveProducts(groupCtx, ownerID)
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to count products", logger.Error(err))
		}
		return err
	})
	group.Go(func() error {
		var err error
		lowStock, err = r.queries.CountLowStockProducts(groupCtx, ownerID)
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to count low stock", logger.Error(err))
		}
		return err
	})
	group.Go(func() error {
		var err error
		pendingOrders, err = r.queries.CountPendingPurchaseOrders(groupCtx, ownerID)
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to count pending POs", logger.Error(err))
		}
		return err
	})
	group.Go(func() error {
		var err error
		totalSuppliers, err = r.queries.CountActiveSuppliers(groupCtx, ownerID)
		if err != nil {
			logger.Ctx(groupCtx).Error("Failed to count suppliers", logger.Error(err))
		}
		return err
	})
	group.Go(func() error {
		var err error
		invValue, err = r.inventoryValue(groupCtx, ownerID)
		return err
	})
	group.Go(func() error {
		var err error
		items, err = r.topLowStockItems(groupCtx, ownerID, 5)
		return err
	})
	group.Go(func() error {
		var err error
		recent, err = r.recentPurchaseOrders(groupCtx, ownerID, 5)
		return err
	})
	group.Go(func() error {
		var err error
		cats, err = r.topCategories(groupCtx, ownerID, 5)
		return err
	})
	if err := group.Wait(); err != nil {
		return domain.DashboardStats{}, err
	}

	return domain.DashboardStats{
		TotalProducts:  int(totalProducts),
		LowStockCount:  int(lowStock),
		PendingOrders:  int(pendingOrders),
		TotalSuppliers: int(totalSuppliers),
		InventoryValue: invValue,
		LowStockItems:  items,
		RecentOrders:   recent,
		TopCategories:  cats,
	}, nil
}

// inventoryValue sums stock_quantity * cost_price for tracked, active
// products. Built with squirrel because the COALESCE(SUM(...)) projection
// scans into sql.NullFloat64 — a shape sqlc would not reproduce cleanly.
func (r *Repository) inventoryValue(ctx context.Context, ownerID int64) (float64, error) {
	query, args, err := squirrel.
		Select("COALESCE(SUM(stock_quantity * cost_price), 0)").
		From("erp_products").
		Where(squirrel.Eq{"business_id": ownerID, "is_active": true, "track_inventory": true}).
		PlaceholderFormat(squirrel.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build inventory value query", logger.Error(err))
		return 0, err
	}

	var invValue sql.NullFloat64
	if err := r.pool.QueryRowContext(ctx, query, args...).Scan(&invValue); err != nil {
		logger.Ctx(ctx).Error("Failed to sum inventory value", logger.Error(err))
		return 0, err
	}

	if invValue.Valid {
		return invValue.Float64, nil
	}

	return 0, nil
}

func (r *Repository) topLowStockItems(ctx context.Context, ownerID int64, limit int) ([]domain.LowStockItem, error) {
	query, args, err := squirrel.
		Select(
			"p.id", "p.name", "COALESCE(p.sku, '')",
			"p.stock_quantity", "COALESCE(p.low_stock_threshold, 0)",
			"''",
		).
		From("erp_products p").
		Where(squirrel.Eq{"p.business_id": ownerID, "p.is_active": true, "p.track_inventory": true}).
		Where("p.low_stock_threshold IS NOT NULL").
		Where("p.stock_quantity <= p.low_stock_threshold").
		OrderBy("(p.stock_quantity - p.low_stock_threshold) ASC", "p.stock_quantity ASC").
		Limit(uint64(limit)).
		PlaceholderFormat(squirrel.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build low stock items query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list low stock items", logger.Error(err))
		return nil, err
	}
	defer closeRows(ctx, rows)

	items := make([]domain.LowStockItem, 0, limit)
	for rows.Next() {
		var it domain.LowStockItem
		if err := rows.Scan(&it.ProductID, &it.Name, &it.SKU,
			&it.CurrentStock, &it.ReorderPoint, &it.WarehouseName); err != nil {
			logger.Ctx(ctx).Error("Failed to scan low stock item", logger.Error(err))
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *Repository) recentPurchaseOrders(ctx context.Context, ownerID int64, limit int) ([]domain.RecentPO, error) {
	// cross-schema read from erp_suppliers — join for display name
	query, args, err := squirrel.
		Select(
			"po.id", "po.po_number", "COALESCE(s.name, 'Unknown')",
			"po.status", "po.total", "po.created_at",
		).
		From("erp_purchase_orders po").
		LeftJoin("erp_suppliers s ON s.id = po.supplier_id").
		Where(squirrel.Eq{"po.business_id": ownerID}).
		OrderBy("po.created_at DESC").
		Limit(uint64(limit)).
		PlaceholderFormat(squirrel.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build recent POs query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list recent POs", logger.Error(err))
		return nil, err
	}
	defer closeRows(ctx, rows)

	pos := make([]domain.RecentPO, 0, limit)
	for rows.Next() {
		var p domain.RecentPO
		var createdAt sql.NullTime
		if err := rows.Scan(&p.ID, &p.PONumber, &p.SupplierName,
			&p.Status, &p.Total, &createdAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan recent PO", logger.Error(err))
			return nil, err
		}
		if createdAt.Valid {
			p.CreatedAt = createdAt.Time
		}
		pos = append(pos, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pos, nil
}

func (r *Repository) topCategories(ctx context.Context, ownerID int64, limit int) ([]domain.TopCategory, error) {
	// cross-schema read from erp_categories — join for display name
	query, args, err := squirrel.
		Select(
			"COALESCE(c.name, 'Uncategorized') AS name",
			"COUNT(p.id) AS product_count",
		).
		From("erp_products p").
		LeftJoin("erp_categories c ON c.id = p.category_id").
		Where(squirrel.Eq{"p.business_id": ownerID, "p.is_active": true}).
		GroupBy("c.id", "c.name").
		OrderBy("product_count DESC").
		Limit(uint64(limit)).
		PlaceholderFormat(squirrel.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build top categories query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list top categories", logger.Error(err))
		return nil, err
	}
	defer closeRows(ctx, rows)

	cats := make([]domain.TopCategory, 0, limit)
	for rows.Next() {
		var c domain.TopCategory
		if err := rows.Scan(&c.Name, &c.ProductCount); err != nil {
			logger.Ctx(ctx).Error("Failed to scan top category", logger.Error(err))
			return nil, err
		}
		cats = append(cats, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cats, nil
}

// ListReorderSuggestions returns products that have fallen to or below
// their reorder threshold, with a suggested order quantity. Without
// sales-history integration the suggested qty is just the recorded
// reorder_quantity; confidence is "low" to flag this to the UI.
func (r *Repository) ListReorderSuggestions(ctx context.Context, ownerID int64) ([]domain.ReorderSuggestion, error) {
	query, args, err := squirrel.
		Select(
			"id", "name", "COALESCE(sku, '')",
			"stock_quantity",
			"COALESCE(low_stock_threshold, 0)",
			"COALESCE(reorder_quantity, 0)",
		).
		From("erp_products").
		Where(squirrel.Eq{"business_id": ownerID, "is_active": true, "track_inventory": true}).
		Where("low_stock_threshold IS NOT NULL").
		Where("stock_quantity <= low_stock_threshold").
		OrderBy("(stock_quantity - low_stock_threshold) ASC").
		Limit(100).
		PlaceholderFormat(squirrel.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("Failed to build reorder suggestions query", logger.Error(err))
		return nil, err
	}

	rows, err := r.pool.QueryContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list reorder suggestions", logger.Error(err))
		return nil, err
	}
	defer closeRows(ctx, rows)

	out := make([]domain.ReorderSuggestion, 0, 100)
	for rows.Next() {
		var s domain.ReorderSuggestion
		if err := rows.Scan(&s.ProductID, &s.Name, &s.SKU,
			&s.CurrentStock, &s.ReorderPoint, &s.SuggestedQty); err != nil {
			logger.Ctx(ctx).Error("Failed to scan reorder suggestion", logger.Error(err))
			return nil, err
		}
		if s.SuggestedQty == 0 {
			// Fall back to doubling the reorder point if no reorder_quantity is set.
			s.SuggestedQty = s.ReorderPoint * 2
		}
		s.Confidence = "low"
		out = append(out, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

// ListExpiringItems is stubbed to return an empty slice. Batch-level
// expiry tracking isn't modelled yet — once an erp_product_batches table
// exists, this method should join and filter by expiry_date <= NOW() + N days.
func (r *Repository) ListExpiringItems(_ context.Context, _ int64, _ int) ([]domain.ExpiryItem, error) {
	return []domain.ExpiryItem{}, nil
}

func closeRows(ctx context.Context, rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		logger.Ctx(ctx).Warn("Failed to close rows", logger.Error(err))
	}
}
