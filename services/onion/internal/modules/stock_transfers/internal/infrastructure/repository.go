package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
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

func (r *Repository) ListStockTransfers(ctx context.Context, ownerID int64) ([]domain.StockTransfer, error) {
	query := `
		SELECT id, owner_id, from_warehouse_id, to_warehouse_id, status, notes, created_at, updated_at
		FROM erp_stock_transfers
		WHERE owner_id = ?
		ORDER BY created_at DESC`

	rows, err := r.pool.QueryContext(ctx, query, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list stock transfers", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var transfers []domain.StockTransfer

	for rows.Next() {
		var t domain.StockTransfer
		if err := t.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan stock transfer row", logger.Error(err))

			return nil, err
		}

		transfers = append(transfers, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range transfers {
		items, err := r.getTransferItems(ctx, transfers[i].ID)
		if err != nil {
			return nil, err
		}

		transfers[i].Items = items
	}

	return transfers, nil
}

func (r *Repository) getTransferItems(ctx context.Context, transferID int64) ([]domain.StockTransferItem, error) {
	query := `
		SELECT id, transfer_id, product_id, quantity, created_at
		FROM erp_stock_transfer_items
		WHERE transfer_id = ?
		ORDER BY id ASC`

	rows, err := r.pool.QueryContext(ctx, query, transferID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get transfer items", logger.Error(err))

		return nil, err
	}

	defer rows.Close()

	var items []domain.StockTransferItem

	for rows.Next() {
		var item domain.StockTransferItem
		if err := item.FromRow(rows); err != nil {
			logger.Ctx(ctx).Error("Failed to scan transfer item row", logger.Error(err))

			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (r *Repository) CreateStockTransfer(ctx context.Context, ownerID int64, input domain.StockTransferInput) (int64, error) {
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
		INSERT INTO erp_stock_transfers (owner_id, from_warehouse_id, to_warehouse_id, status, notes)
		VALUES (?, ?, ?, 'draft', ?)`

	res, err := tx.ExecContext(ctx, query,
		ownerID,
		input.FromWarehouseID,
		input.ToWarehouseID,
		nullString(input.Notes),
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create stock transfer", logger.Error(err))

		return 0, err
	}

	transferID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	itemQuery := `
		INSERT INTO erp_stock_transfer_items (transfer_id, product_id, quantity)
		VALUES (?, ?, ?)`

	for _, item := range input.Items {
		_, err = tx.ExecContext(ctx, itemQuery, transferID, item.ProductID, item.Quantity)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to create stock transfer item", logger.Error(err))

			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		logger.Ctx(ctx).Error("Failed to commit stock transfer", logger.Error(err))

		return 0, err
	}

	return transferID, nil
}

func (r *Repository) UpdateStockTransferStatus(ctx context.Context, ownerID, id int64, status domain.StockTransferStatus) error {
	query := `
		UPDATE erp_stock_transfers
		SET status = ?
		WHERE owner_id = ? AND id = ?`

	_, err := r.pool.ExecContext(ctx, query, string(status), ownerID, id)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update stock transfer status", logger.Error(err))

		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
