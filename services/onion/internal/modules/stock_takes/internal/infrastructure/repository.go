package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/stock_takes/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		pool: pool,
	}
}

func (r *Repository) ListStockTakes(ctx context.Context, ownerID int64) ([]domain.StockTake, error) {
	query := `
		SELECT id, owner_id, warehouse_id, status, notes, started_at, completed_at, created_at
		FROM erp_stock_takes
		WHERE owner_id = ?
		ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list stock takes", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var stockTakes []domain.StockTake

	for rows.Next() {
		var s domain.StockTake
		if err := s.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan stock take row", logger.Error(err))

			return nil, err
		}

		stockTakes = append(stockTakes, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range stockTakes {
		items, err := r.getStockTakeItems(ctx, stockTakes[i].ID)
		if err != nil {
			return nil, err
		}

		stockTakes[i].Items = items
	}

	return stockTakes, nil
}

func (r *Repository) getStockTakeItems(ctx context.Context, stockTakeID int64) ([]domain.StockTakeItem, error) {
	query := `
		SELECT id, stock_take_id, product_id, expected_qty, counted_qty, variance, created_at
		FROM erp_stock_take_items
		WHERE stock_take_id = ?
		ORDER BY id ASC`

	rows, err := r.pool.QueryContext(ctx, query, stockTakeID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get stock take items", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var items []domain.StockTakeItem

	for rows.Next() {
		var item domain.StockTakeItem
		if err := item.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan stock take item row", logger.Error(err))

			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
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

	query := `
		INSERT INTO erp_stock_takes (owner_id, warehouse_id, status, notes)
		VALUES (?, ?, 'draft', ?)`

	res, err := tx.ExecContext(ctx, query,
		ownerID,
		nullInt64(input.WarehouseID),
		nullString(input.Notes),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create stock take", logger.Error(err))

		return 0, err
	}

	stockTakeID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	itemQuery := `
		INSERT INTO erp_stock_take_items (stock_take_id, product_id, expected_qty, counted_qty, variance)
		VALUES (?, ?, ?, ?, ?)`

	for _, item := range input.Items {
		variance := item.CountedQty - item.ExpectedQty
		_, err = tx.ExecContext(ctx, itemQuery, stockTakeID, item.ProductID, item.ExpectedQty, item.CountedQty, variance)
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

	query := `
		UPDATE erp_stock_takes
		SET status = ?, notes = ?`

	args := []interface{}{string(input.Status), nullString(input.Notes)}

	if input.Status == domain.StockTakeStatusInProgress {
		query += `, started_at = NOW()`
	} else if input.Status == domain.StockTakeStatusCompleted {
		query += `, completed_at = NOW()`
	}

	query += ` WHERE owner_id = ? AND id = ?`
	args = append(args, ownerID, input.ID)

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update stock take", logger.Error(err))

		return err
	}

	if len(input.Items) > 0 {
		_, err = tx.ExecContext(ctx, `DELETE FROM erp_stock_take_items WHERE stock_take_id = ?`, input.ID)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to clear stock take items", logger.Error(err))

			return err
		}

		itemQuery := `
			INSERT INTO erp_stock_take_items (stock_take_id, product_id, expected_qty, counted_qty, variance)
			VALUES (?, ?, ?, ?, ?)`

		for _, item := range input.Items {
			variance := item.CountedQty - item.ExpectedQty
			_, err = tx.ExecContext(ctx, itemQuery, input.ID, item.ProductID, item.ExpectedQty, item.CountedQty, variance)
			if err != nil {
				logger.Ctx(ctx).Error("Failed to insert stock take item", logger.Error(err))

				return err
			}
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit stock take update", logger.Error(err))

		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}

func nullInt64(v int64) sql.NullInt64 {
	return sql.NullInt64{Int64: v, Valid: v != 0}
}
