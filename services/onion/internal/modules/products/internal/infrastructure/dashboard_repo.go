package infrastructure

import (
	"context"
	"fmt"
	"strconv"

	"golang.org/x/sync/errgroup"

	"github.com/voxire/onion/internal/modules/products/internal/domain"
	"github.com/voxire/onion/internal/modules/products/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) inventoryValue(ctx context.Context, ownerID int64) (float64, error) {
	raw, err := r.queries.SumInventoryValue(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to sum inventory value", logger.Error(err))
		return 0, err
	}
	return interfaceToFloat(raw), nil
}

func (r *Repository) topLowStockItems(ctx context.Context, ownerID int64, limit int) ([]domain.LowStockItem, error) {
	rows, err := r.queries.ListTopLowStockItems(ctx, sqlc.ListTopLowStockItemsParams{
		BusinessID: ownerID,
		Limit:      int32(limit), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list low stock items", logger.Error(err))
		return nil, err
	}

	items := make([]domain.LowStockItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.LowStockItem{
			ProductID:     row.ID,
			Name:          row.Name,
			SKU:           row.Sku,
			CurrentStock:  sqlutil.ParseDecimal(row.StockQuantity),
			ReorderPoint:  sqlutil.ParseDecimal(row.LowStockThreshold),
			WarehouseName: "",
		})
	}
	return items, nil
}

func (r *Repository) recentPurchaseOrders(ctx context.Context, ownerID int64, limit int) ([]domain.RecentPO, error) {
	// cross-schema read from erp_suppliers — join for display name
	rows, err := r.queries.ListRecentPurchaseOrders(ctx, sqlc.ListRecentPurchaseOrdersParams{
		BusinessID: ownerID,
		Limit:      int32(limit), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list recent POs", logger.Error(err))
		return nil, err
	}

	pos := make([]domain.RecentPO, 0, len(rows))
	for _, row := range rows {
		p := domain.RecentPO{
			ID:           row.ID,
			PONumber:     row.PoNumber,
			SupplierName: row.SupplierName,
			Total:        sqlutil.ParseDecimal(row.Total.String),
		}
		if row.Status.Valid {
			p.Status = string(row.Status.ErpPurchaseOrdersStatus)
		}
		if row.CreatedAt.Valid {
			p.CreatedAt = row.CreatedAt.Time
		}
		pos = append(pos, p)
	}
	return pos, nil
}

func (r *Repository) topCategories(ctx context.Context, ownerID int64, limit int) ([]domain.TopCategory, error) {
	// cross-schema read from erp_categories — join for display name
	rows, err := r.queries.ListTopCategories(ctx, sqlc.ListTopCategoriesParams{
		BusinessID: ownerID,
		Limit:      int32(limit), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list top categories", logger.Error(err))
		return nil, err
	}

	cats := make([]domain.TopCategory, 0, len(rows))
	for _, row := range rows {
		cats = append(cats, domain.TopCategory{
			Name:         row.Name,
			ProductCount: int(row.ProductCount),
		})
	}
	return cats, nil
}

// ListReorderSuggestions returns products that have fallen to or below
// their reorder threshold, with a suggested order quantity. Without
// sales-history integration the suggested qty is just the recorded
// reorder_quantity; confidence is "low" to flag this to the UI.
func (r *Repository) ListReorderSuggestions(ctx context.Context, ownerID int64) (domain.ReorderSuggestionsResult, error) {
	rows, err := r.queries.ListReorderSuggestions(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list reorder suggestions", logger.Error(err))
		return domain.ReorderSuggestionsResult{}, err
	}

	truncated := len(rows) > 100
	if truncated {
		rows = rows[:100]
	}
	out := make([]domain.ReorderSuggestion, 0, len(rows))
	for _, row := range rows {
		s := domain.ReorderSuggestion{
			ProductID:    row.ID,
			Name:         row.Name,
			SKU:          row.Sku,
			CurrentStock: sqlutil.ParseDecimal(row.StockQuantity),
			ReorderPoint: sqlutil.ParseDecimal(row.LowStockThreshold),
			SuggestedQty: sqlutil.ParseDecimal(row.ReorderQuantity),
		}
		if s.SuggestedQty == 0 {
			s.SuggestedQty = s.ReorderPoint * 2
		}
		s.Confidence = "low"
		out = append(out, s)
	}
	return domain.ReorderSuggestionsResult{Suggestions: out, Truncated: truncated}, nil
}

// ListExpiringItems is stubbed to return an empty slice. Batch-level
// expiry tracking isn't modelled yet — once an erp_product_batches table
// exists, this method should join and filter by expiry_date <= NOW() + N days.
func (r *Repository) ListExpiringItems(_ context.Context, _ int64, _ int) ([]domain.ExpiryItem, error) {
	return []domain.ExpiryItem{}, nil
}

func interfaceToFloat(v interface{}) float64 {
	if v == nil {
		return 0
	}
	if f, ok := v.(float64); ok {
		return f
	}
	var str string
	switch val := v.(type) {
	case []byte:
		str = string(val)
	case string:
		str = val
	default:
		str = fmt.Sprintf("%v", val)
	}
	if str == "" {
		return 0
	}
	f, err := strconv.ParseFloat(str, 64)
	if err != nil {
		logger.Global().Warn("failed to parse dashboard decimal from database",
			logger.String("value", str), logger.Error(err))
		return 0
	}
	return f
}
