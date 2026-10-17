package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/voxire/onion/internal/modules/erp/internal/outbox"

	"github.com/voxire/onion/internal/contracts/audit"
	"github.com/voxire/onion/internal/modules/erp/internal/stock"
	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/modules/erp/stock_takes/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListStockTakes(ctx context.Context, businessID int64, limit, offset int) ([]domain.StockTake, bool, error) {
	rows, err := r.queries.ListStockTakesPaged(ctx, sqlc.ListStockTakesPagedParams{
		BusinessID: businessID,
		Limit:      int32(limit + 1), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:     int32(offset),    // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list stock takes", logger.Error(err))
		return nil, false, err
	}

	stockTakes := make([]domain.StockTake, 0, limit+1)
	for _, row := range rows {
		s := domain.StockTake{
			ID:         row.ID,
			OwnerID:    row.OwnerID,
			BusinessID: row.BusinessID,
		}

		if row.WarehouseID.Valid {
			s.WarehouseID = row.WarehouseID.Int64
		}
		if row.Status.Valid {
			s.Status = domain.StockTakeStatus(row.Status.ErpStockTakesStatus)
		}
		if row.Notes.Valid {
			s.Notes = row.Notes.String
		}
		if row.StartedAt.Valid {
			s.StartedAt = row.StartedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if row.CompletedAt.Valid {
			s.CompletedAt = row.CompletedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if row.CreatedAt.Valid {
			s.CreatedAt = row.CreatedAt.Time
		}

		stockTakes = append(stockTakes, s)
	}

	hasMore := len(stockTakes) > limit
	if hasMore {
		stockTakes = stockTakes[:limit]
	}

	itemsByStockTake, err := r.getStockTakeItemsBatch(ctx, stockTakes)
	if err != nil {
		return nil, false, err
	}
	for i := range stockTakes {
		stockTakes[i].Items = itemsByStockTake[stockTakes[i].ID]
	}

	return stockTakes, hasMore, nil
}

// getStockTakeItemsBatch loads all items for the given stock takes in a single
// query (WHERE stock_take_id IN (...)) and groups them by stock take ID,
// replacing the previous per-stock-take N+1 query pattern.
func (r *Repository) getStockTakeItemsBatch(ctx context.Context, stockTakes []domain.StockTake) (map[int64][]domain.StockTakeItem, error) {
	if len(stockTakes) == 0 {
		return nil, nil
	}

	ids := make([]int64, 0, len(stockTakes))
	for i := range stockTakes {
		ids = append(ids, stockTakes[i].ID)
	}

	rows, err := r.queries.GetStockTakeItemsByStockTakeIDs(ctx, ids)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch get stock take items", logger.Error(err))
		return nil, err
	}

	result := make(map[int64][]domain.StockTakeItem, len(stockTakes))
	for _, row := range rows {
		item := domain.StockTakeItem{
			ID:          row.ID,
			StockTakeID: row.StockTakeID,
			ProductID:   row.ProductID,
			ExpectedQty: sqlutil.ParseDecimal(row.ExpectedQty),
		}
		if row.CountedQty.Valid {
			item.CountedQty = sqlutil.ParseDecimal(row.CountedQty.String)
		}
		if row.Variance.Valid {
			item.Variance = sqlutil.ParseDecimal(row.Variance.String)
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		result[item.StockTakeID] = append(result[item.StockTakeID], item)
	}

	return result, nil
}

func (r *Repository) CreateStockTake(ctx context.Context, businessID int64, input domain.StockTakeInput, events ...audit.Event) (int64, error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transaction", logger.Error(err))
		return 0, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)
	if err = validateStockTakeReferences(ctx, qtx, businessID, input); err != nil {
		return 0, err
	}
	items, err := snapshotStockTakeItems(ctx, tx, businessID, input.WarehouseID, input.Items)
	if err != nil {
		return 0, err
	}

	res, err := qtx.InsertStockTake(ctx, sqlc.InsertStockTakeParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		WarehouseID: sqlutil.NullInt64(input.WarehouseID),
		Notes:       sqlutil.NullString(input.Notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create stock take", logger.Error(err))
		return 0, err
	}

	stockTakeID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range items {
		// Draft creation snapshots expected stock only; physical counts remain
		// absent until an explicit in-progress or completion update arrives.
		_, err = qtx.InsertStockTakeItem(ctx, sqlc.InsertStockTakeItemParams{
			StockTakeID: stockTakeID,
			ProductID:   item.ProductID,
			ExpectedQty: strconv.FormatFloat(item.ExpectedQty, 'f', -1, 64),
			CountedQty:  sql.NullString{},
			Variance:    sql.NullString{},
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to create stock take item", logger.Error(err))
			return 0, err
		}
	}

	if err = appendStockTakeAuditEvents(ctx, tx, events, stockTakeID); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit stock take", logger.Error(err))
		return 0, err
	}

	return stockTakeID, nil
}

func snapshotStockTakeItems(ctx context.Context, tx *sql.Tx, businessID, warehouseID int64, requested []domain.StockTakeItemInput) ([]domain.StockTakeItemInput, error) {
	var (
		locationID int64
		err        error
	)
	if warehouseID > 0 {
		locationID, err = stock.WarehouseLocationID(ctx, tx, businessID, warehouseID)
	} else {
		locationID, err = stock.DefaultLocationID(ctx, tx, businessID)
	}
	if err != nil {
		return nil, err
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT product_id, on_hand
		FROM stock_levels
		WHERE owner_id = ? AND location_id = ?`, businessID, locationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	onHand := make(map[int64]float64, len(requested))
	for rows.Next() {
		var productID int64
		var quantity string
		if err := rows.Scan(&productID, &quantity); err != nil {
			return nil, err
		}
		onHand[productID] = sqlutil.ParseDecimal(quantity)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	items := make([]domain.StockTakeItemInput, 0, len(requested))
	for _, item := range requested {
		item.ExpectedQty = math.Max(0, onHand[item.ProductID])
		item.CountedQty = 0
		items = append(items, item)
	}
	return items, nil
}

func validateStockTakeReferences(ctx context.Context, qtx *sqlc.Queries, businessID int64, input domain.StockTakeInput) error {
	if input.WarehouseID > 0 {
		if _, err := qtx.GetWarehouseReference(ctx, sqlc.GetWarehouseReferenceParams{BusinessID: businessID, ID: input.WarehouseID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrStockTakeReferenceNotFound
			}
			return err
		}
	}
	for _, item := range input.Items {
		if _, err := qtx.GetProductReference(ctx, sqlc.GetProductReferenceParams{BusinessID: businessID, ID: item.ProductID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrStockTakeReferenceNotFound
			}
			return err
		}
	}
	return nil
}

func (r *Repository) UpdateStockTake(ctx context.Context, businessID int64, input domain.StockTakeUpdateInput, events ...audit.Event) (err error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transaction", logger.Error(err))
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)
	prior, err := qtx.GetStockTakeStatusForUpdate(ctx, sqlc.GetStockTakeStatusForUpdateParams{
		BusinessID: businessID,
		ID:         input.ID,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrStockTakeNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to lock stock take", logger.Error(err))
		return err
	}
	alreadyCompleted := prior.Status.Valid && prior.Status.ErpStockTakesStatus == sqlc.ErpStockTakesStatus(domain.StockTakeStatusCompleted)
	if alreadyCompleted {
		if input.Status == domain.StockTakeStatusCompleted {
			return nil
		}
		return domain.ErrStockTakeAlreadyCompleted
	}

	if !alreadyCompleted {
		if err = updateStockTakeCounts(ctx, qtx, input.ID, input.Items, input.Status == domain.StockTakeStatusCompleted); err != nil {
			return err
		}
	}

	statusVal := sqlc.NullErpStockTakesStatus{
		ErpStockTakesStatus: sqlc.ErpStockTakesStatus(input.Status),
		Valid:               true,
	}
	switch input.Status {
	case domain.StockTakeStatusInProgress:
		err = qtx.UpdateStockTakeStatusInProgress(ctx, sqlc.UpdateStockTakeStatusInProgressParams{
			Status: statusVal, Notes: sqlutil.NullString(input.Notes), BusinessID: businessID, ID: input.ID,
		})
	case domain.StockTakeStatusCompleted:
		err = qtx.UpdateStockTakeStatusCompleted(ctx, sqlc.UpdateStockTakeStatusCompletedParams{
			Status: statusVal, Notes: sqlutil.NullString(input.Notes), BusinessID: businessID, ID: input.ID,
		})
	default:
		err = qtx.UpdateStockTakeStatus(ctx, sqlc.UpdateStockTakeStatusParams{
			Status: statusVal, Notes: sqlutil.NullString(input.Notes), BusinessID: businessID, ID: input.ID,
		})
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update stock take", logger.Error(err))
		return err
	}

	if input.Status == domain.StockTakeStatusCompleted && !alreadyCompleted {
		if err = r.reconcileStockTake(ctx, tx, qtx, businessID, input.ID, prior.WarehouseID); err != nil {
			logger.Ctx(ctx).Error("Failed to reconcile stock take", logger.Error(err))
			return err
		}
	}
	if err = appendStockTakeAuditEvents(ctx, tx, events, input.ID); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit stock take update", logger.Error(err))
		return err
	}
	return nil
}

func updateStockTakeCounts(ctx context.Context, qtx *sqlc.Queries, stockTakeID int64, submitted []domain.StockTakeItemInput, requireAll bool) error {
	existing, err := qtx.GetStockTakeItems(ctx, stockTakeID)
	if err != nil {
		return err
	}
	if requireAll && len(submitted) != len(existing) {
		return domain.ErrStockTakeIncomplete
	}
	if len(submitted) == 0 {
		return nil
	}

	byProduct := make(map[int64]sqlc.ErpStockTakeItem, len(existing))
	for _, item := range existing {
		byProduct[item.ProductID] = item
	}
	matched := make(map[int64]struct{}, len(submitted))
	for _, item := range submitted {
		if item.CountedQty < 0 || math.IsNaN(item.CountedQty) || math.IsInf(item.CountedQty, 0) {
			return domain.ErrStockTakeIncomplete
		}
		if _, duplicate := matched[item.ProductID]; duplicate {
			return domain.ErrStockTakeIncomplete
		}
		stored, ok := byProduct[item.ProductID]
		if !ok {
			return domain.ErrStockTakeReferenceNotFound
		}
		expected := sqlutil.ParseDecimal(stored.ExpectedQty)
		counted := strconv.FormatFloat(item.CountedQty, 'f', -1, 64)
		variance := strconv.FormatFloat(item.CountedQty-expected, 'f', -1, 64)
		updated, err := qtx.UpdateStockTakeItemCount(ctx, sqlc.UpdateStockTakeItemCountParams{
			CountedQty:  sql.NullString{String: counted, Valid: true},
			Variance:    sql.NullString{String: variance, Valid: true},
			StockTakeID: stockTakeID,
			ProductID:   item.ProductID,
		})
		if err != nil {
			return err
		}
		if updated != 1 {
			return domain.ErrStockTakeReferenceNotFound
		}
		matched[item.ProductID] = struct{}{}
	}
	if requireAll && len(matched) != len(existing) {
		return domain.ErrStockTakeIncomplete
	}
	return nil
}

func appendStockTakeAuditEvents(ctx context.Context, tx *sql.Tx, events []audit.Event, entityID int64) error {
	for _, event := range events {
		event.EntityID = entityID
		switch value := event.OldValues.(type) {
		case domain.StockTakeInput:
			event.OldValues = stockTakeInputSnapshotOf(value)
		case domain.StockTakeUpdateInput:
			event.OldValues = stockTakeUpdateInputSnapshotOf(value)
		}
		switch value := event.NewValues.(type) {
		case domain.StockTakeInput:
			event.NewValues = stockTakeInputSnapshotOf(value)
		case domain.StockTakeUpdateInput:
			event.NewValues = stockTakeUpdateInputSnapshotOf(value)
		}
		if err := outbox.AppendAuditTx(ctx, tx, event); err != nil {
			return err
		}
	}
	return nil
}

// reconcileStockTake sets each counted product's on-hand stock to the physical
// count recorded on the stock take and writes an 'adjustment' inventory
// transaction for the variance. Runs inside the caller's transaction.
func (r *Repository) reconcileStockTake(ctx context.Context, tx *sql.Tx, qtx *sqlc.Queries, businessID, stockTakeID int64, warehouseID sql.NullInt64) error {
	items, err := qtx.GetStockTakeItems(ctx, stockTakeID)
	if err != nil {
		return err
	}
	// Lock stock rows in product order, like every other multi-product writer.
	sort.Slice(items, func(i, j int) bool { return items[i].ProductID < items[j].ProductID })

	var locID int64
	if warehouseID.Valid && warehouseID.Int64 > 0 {
		locID, err = stock.WarehouseLocationID(ctx, tx, businessID, warehouseID.Int64)
	} else {
		// Legacy stock takes without a selected warehouse continue to use the
		// tenant default location.
		locID, err = stock.DefaultLocationID(ctx, tx, businessID)
	}
	if err != nil {
		return err
	}

	for _, it := range items {
		if !it.CountedQty.Valid {
			continue
		}
		counted, perr := strconv.ParseFloat(it.CountedQty.String, 64)
		if perr != nil {
			return perr
		}
		if counted < 0 {
			continue // a physical count can't be negative; skip bad rows
		}

		// Set the location's on-hand to the physical count; stock_core ledgers
		// the 'count' entry and syncs the scalar cache.
		before, delta, serr := stock.SetLevelTx(ctx, tx, stock.Movement{
			OwnerID:       businessID,
			LocationID:    locID,
			ProductID:     it.ProductID,
			EntryType:     stock.EntryCount,
			ReferenceID:   &stockTakeID,
			ReferenceType: "stock_take",
			Channel:       "erp",
			CreatedBy:     &businessID,
			Note:          fmt.Sprintf("stock take %d reconciliation", stockTakeID),
		}, counted)
		if serr != nil {
			return serr
		}
		if delta == 0 {
			continue // counted matched on-hand; no legacy variance row needed
		}

		// Legacy ledger row (read-compat).
		if _, err = qtx.InsertInventoryTransactionWithRef(ctx, sqlc.InsertInventoryTransactionWithRefParams{
			OwnerID:        businessID,
			BusinessID:     businessID,
			ProductID:      it.ProductID,
			Type:           sqlc.ErpInventoryTransactionsType("adjustment"),
			QuantityChange: strconv.FormatFloat(delta, 'f', -1, 64),
			QuantityBefore: strconv.FormatFloat(before, 'f', -1, 64),
			QuantityAfter:  strconv.FormatFloat(counted, 'f', -1, 64),
			ReferenceID:    sql.NullInt64{Int64: stockTakeID, Valid: true},
			ReferenceType:  sql.NullString{String: "stock_take", Valid: true},
			CreatedBy:      sql.NullInt64{Int64: businessID, Valid: true},
			Notes:          sql.NullString{String: fmt.Sprintf("stock take %d reconciliation", stockTakeID), Valid: true},
		}); err != nil {
			return err
		}
	}

	return nil
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
