package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_takes/internal/infrastructure/persistence/sqlc"
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
			ExpectedQty: parseDecimal(row.ExpectedQty),
			CountedQty:  parseDecimal(row.CountedQty),
			Variance:    parseDecimal(row.Variance),
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		result[item.StockTakeID] = append(result[item.StockTakeID], item)
	}

	return result, nil
}

func (r *Repository) CreateStockTake(ctx context.Context, businessID int64, input domain.StockTakeInput) (int64, error) {
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
	if err := validateStockTakeReferences(ctx, qtx, businessID, input); err != nil {
		return 0, err
	}

	res, err := qtx.InsertStockTake(ctx, sqlc.InsertStockTakeParams{
		OwnerID:     businessID,
		BusinessID:  businessID,
		WarehouseID: nullInt64(input.WarehouseID),
		Notes:       nullString(input.Notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create stock take", logger.Error(err))

		return 0, err
	}

	stockTakeID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range input.Items {
		variance := item.CountedQty - item.ExpectedQty
		_, err = qtx.InsertStockTakeItem(ctx, sqlc.InsertStockTakeItemParams{
			StockTakeID: stockTakeID,
			ProductID:   item.ProductID,
			ExpectedQty: strconv.FormatFloat(item.ExpectedQty, 'f', -1, 64),
			CountedQty:  strconv.FormatFloat(item.CountedQty, 'f', -1, 64),
			Variance:    strconv.FormatFloat(variance, 'f', -1, 64),
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to create stock take item", logger.Error(err))

			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit stock take", logger.Error(err))

		return 0, err
	}

	return stockTakeID, nil
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

func (r *Repository) UpdateStockTake(ctx context.Context, businessID int64, input domain.StockTakeUpdateInput) error {
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

	// Lock the row before any status update so unknown IDs fail closed. This
	// also lets repeated completion requests remain idempotent when MySQL
	// reports zero changed rows for an unchanged status.
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
	warehouseID := prior.WarehouseID

	statusVal := sqlc.NullErpStockTakesStatus{
		ErpStockTakesStatus: sqlc.ErpStockTakesStatus(input.Status),
		Valid:               true,
	}

	switch input.Status {
	case domain.StockTakeStatusInProgress:
		err = qtx.UpdateStockTakeStatusInProgress(ctx, sqlc.UpdateStockTakeStatusInProgressParams{
			Status:     statusVal,
			Notes:      nullString(input.Notes),
			BusinessID: businessID,
			ID:         input.ID,
		})
	case domain.StockTakeStatusCompleted:
		err = qtx.UpdateStockTakeStatusCompleted(ctx, sqlc.UpdateStockTakeStatusCompletedParams{
			Status:     statusVal,
			Notes:      nullString(input.Notes),
			BusinessID: businessID,
			ID:         input.ID,
		})
	default:
		err = qtx.UpdateStockTakeStatus(ctx, sqlc.UpdateStockTakeStatusParams{
			Status:     statusVal,
			Notes:      nullString(input.Notes),
			BusinessID: businessID,
			ID:         input.ID,
		})
	}

	if err != nil {
		logger.Ctx(ctx).Error("Failed to update stock take", logger.Error(err))

		return err
	}

	if len(input.Items) > 0 {
		err = qtx.DeleteStockTakeItems(ctx, input.ID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to clear stock take items", logger.Error(err))

			return err
		}

		for _, item := range input.Items {
			variance := item.CountedQty - item.ExpectedQty
			_, err = qtx.InsertStockTakeItem(ctx, sqlc.InsertStockTakeItemParams{
				StockTakeID: input.ID,
				ProductID:   item.ProductID,
				ExpectedQty: strconv.FormatFloat(item.ExpectedQty, 'f', -1, 64),
				CountedQty:  strconv.FormatFloat(item.CountedQty, 'f', -1, 64),
				Variance:    strconv.FormatFloat(variance, 'f', -1, 64),
			})
			if err != nil {
				logger.Ctx(ctx).Error("Failed to insert stock take item", logger.Error(err))

				return err
			}
		}
	}

	// On the draft/in_progress -> completed transition, reconcile on-hand stock
	// to the counted quantities and ledger the variance.
	if input.Status == domain.StockTakeStatusCompleted && !alreadyCompleted {
		if err = r.reconcileStockTake(ctx, tx, qtx, businessID, input.ID, warehouseID); err != nil {
			logger.Ctx(ctx).Error("Failed to reconcile stock take", logger.Error(err))
			return err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit stock take update", logger.Error(err))

		return err
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

	var locID int64
	if warehouseID.Valid && warehouseID.Int64 > 0 {
		locID, err = stock_core.WarehouseLocationID(ctx, tx, businessID, warehouseID.Int64)
	} else {
		// Legacy stock takes without a selected warehouse continue to use the
		// tenant default location.
		locID, err = stock_core.DefaultLocationID(ctx, tx, businessID)
	}
	if err != nil {
		return err
	}

	for _, it := range items {
		counted, perr := strconv.ParseFloat(it.CountedQty, 64)
		if perr != nil {
			return perr
		}
		if counted < 0 {
			continue // a physical count can't be negative; skip bad rows
		}

		// Set the location's on-hand to the physical count; stock_core ledgers
		// the 'count' entry and syncs the scalar cache.
		before, delta, serr := stock_core.SetLevelTx(ctx, tx, stock_core.Movement{
			OwnerID:       businessID,
			LocationID:    locID,
			ProductID:     it.ProductID,
			EntryType:     stock_core.EntryCount,
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

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
func parseDecimal(s string) float64 {
	if s == "" {
		return 0
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		logger.Global().Warn("failed to parse stock-take decimal from database",
			logger.String("value", s), logger.Error(err))
		return 0
	}
	return f
}
