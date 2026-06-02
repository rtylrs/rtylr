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

func (r *Repository) ListStockTakes(ctx context.Context, ownerID int64, limit, offset int) ([]domain.StockTake, bool, error) {
	const q = `
		SELECT id, owner_id, warehouse_id, status, notes, started_at, completed_at, created_at
		FROM erp_stock_takes
		WHERE owner_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ? OFFSET ?`

	rows, err := r.pool.QueryContext(ctx, q, ownerID, limit+1, offset)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list stock takes", logger.Error(err))
		return nil, false, err
	}
	defer rows.Close()

	stockTakes := make([]domain.StockTake, 0, limit+1)
	for rows.Next() {
		var s domain.StockTake
		var warehouseID sql.NullInt64
		var status, notes sql.NullString
		var startedAt, completedAt, createdAt sql.NullTime

		if err := rows.Scan(
			&s.ID, &s.OwnerID, &warehouseID, &status, &notes,
			&startedAt, &completedAt, &createdAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan stock take row", logger.Error(err))
			return nil, false, err
		}

		if warehouseID.Valid {
			s.WarehouseID = warehouseID.Int64
		}
		if status.Valid {
			s.Status = domain.StockTakeStatus(status.String)
		}
		if notes.Valid {
			s.Notes = notes.String
		}
		if startedAt.Valid {
			s.StartedAt = startedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if completedAt.Valid {
			s.CompletedAt = completedAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		if createdAt.Valid {
			s.CreatedAt = createdAt.Time
		}

		stockTakes = append(stockTakes, s)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
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

	placeholders := make([]byte, 0, len(stockTakes)*2)
	args := make([]any, 0, len(stockTakes))
	for i := range stockTakes {
		if i > 0 {
			placeholders = append(placeholders, ',', '?')
		} else {
			placeholders = append(placeholders, '?')
		}
		args = append(args, stockTakes[i].ID)
	}

	q := `SELECT id, stock_take_id, product_id, expected_qty, counted_qty, variance, created_at
		FROM erp_stock_take_items
		WHERE stock_take_id IN (` + string(placeholders) + `)
		ORDER BY id ASC`

	rows, err := r.pool.QueryContext(ctx, q, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch get stock take items", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]domain.StockTakeItem, len(stockTakes))
	for rows.Next() {
		var item domain.StockTakeItem
		var expectedQty, countedQty, variance string
		var createdAt sql.NullTime
		if err := rows.Scan(
			&item.ID, &item.StockTakeID, &item.ProductID,
			&expectedQty, &countedQty, &variance, &createdAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan stock take item row", logger.Error(err))
			return nil, err
		}
		item.ExpectedQty, _ = strconv.ParseFloat(expectedQty, 64)
		item.CountedQty, _ = strconv.ParseFloat(countedQty, 64)
		item.Variance, _ = strconv.ParseFloat(variance, 64)
		if createdAt.Valid {
			item.CreatedAt = createdAt.Time.Format("2006-01-02T15:04:05Z07:00")
		}
		result[item.StockTakeID] = append(result[item.StockTakeID], item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (r *Repository) CreateStockTake(ctx context.Context, ownerID int64, input domain.StockTakeInput) (int64, error) {
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

	res, err := qtx.InsertStockTake(ctx, sqlc.InsertStockTakeParams{
		OwnerID:     ownerID,
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

func (r *Repository) UpdateStockTake(ctx context.Context, ownerID int64, input domain.StockTakeUpdateInput) error {
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

	// When completing a stock take we reconcile on-hand stock to the counted
	// quantities (and ledger the variance). Lock the row first and capture the
	// prior status so a repeated "complete" is a no-op rather than re-applying
	// (and double-ledgering) the same counts.
	alreadyCompleted := false
	if input.Status == domain.StockTakeStatusCompleted {
		var prior sqlc.NullErpStockTakesStatus
		prior, err = qtx.GetStockTakeStatusForUpdate(ctx, sqlc.GetStockTakeStatusForUpdateParams{
			OwnerID: ownerID,
			ID:      input.ID,
		})
		if errors.Is(err, sql.ErrNoRows) {
			err = domain.ErrStockTakeNotFound
			return err
		}
		if err != nil {
			logger.Ctx(ctx).Error("Failed to lock stock take for completion", logger.Error(err))
			return err
		}
		alreadyCompleted = prior.Valid && prior.ErpStockTakesStatus == sqlc.ErpStockTakesStatus(domain.StockTakeStatusCompleted)
	}

	statusVal := sqlc.NullErpStockTakesStatus{
		ErpStockTakesStatus: sqlc.ErpStockTakesStatus(input.Status),
		Valid:               true,
	}

	switch input.Status {
	case domain.StockTakeStatusInProgress:
		err = qtx.UpdateStockTakeStatusInProgress(ctx, sqlc.UpdateStockTakeStatusInProgressParams{
			Status:  statusVal,
			Notes:   nullString(input.Notes),
			OwnerID: ownerID,
			ID:      input.ID,
		})
	case domain.StockTakeStatusCompleted:
		err = qtx.UpdateStockTakeStatusCompleted(ctx, sqlc.UpdateStockTakeStatusCompletedParams{
			Status:  statusVal,
			Notes:   nullString(input.Notes),
			OwnerID: ownerID,
			ID:      input.ID,
		})
	default:
		err = qtx.UpdateStockTakeStatus(ctx, sqlc.UpdateStockTakeStatusParams{
			Status:  statusVal,
			Notes:   nullString(input.Notes),
			OwnerID: ownerID,
			ID:      input.ID,
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
		if err = r.reconcileStockTake(ctx, qtx, ownerID, input.ID); err != nil {
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
func (r *Repository) reconcileStockTake(ctx context.Context, qtx *sqlc.Queries, ownerID, stockTakeID int64) error {
	items, err := qtx.GetStockTakeItems(ctx, stockTakeID)
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

		stockStr, serr := qtx.GetProductStockForUpdate(ctx, sqlc.GetProductStockForUpdateParams{
			OwnerID: ownerID,
			ID:      it.ProductID,
		})
		if errors.Is(serr, sql.ErrNoRows) {
			continue // product gone / not owned; skip rather than fail the whole take
		}
		if serr != nil {
			return serr
		}

		current, perr := strconv.ParseFloat(stockStr, 64)
		if perr != nil {
			return perr
		}

		delta := counted - current
		if delta == 0 {
			continue
		}

		if err = qtx.UpdateProductStock(ctx, sqlc.UpdateProductStockParams{
			StockQuantity: strconv.FormatFloat(counted, 'f', -1, 64),
			OwnerID:       ownerID,
			ID:            it.ProductID,
		}); err != nil {
			return err
		}

		if _, err = qtx.InsertInventoryTransactionWithRef(ctx, sqlc.InsertInventoryTransactionWithRefParams{
			OwnerID:        ownerID,
			ProductID:      it.ProductID,
			Type:           sqlc.ErpInventoryTransactionsType("adjustment"),
			QuantityChange: strconv.FormatFloat(delta, 'f', -1, 64),
			QuantityBefore: strconv.FormatFloat(current, 'f', -1, 64),
			QuantityAfter:  strconv.FormatFloat(counted, 'f', -1, 64),
			ReferenceID:    sql.NullInt64{Int64: stockTakeID, Valid: true},
			ReferenceType:  sql.NullString{String: "stock_take", Valid: true},
			CreatedBy:      sql.NullInt64{Int64: ownerID, Valid: true},
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
