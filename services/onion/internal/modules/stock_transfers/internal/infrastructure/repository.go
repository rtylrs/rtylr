package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_transfers/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListStockTransfers(ctx context.Context, ownerID int64, limit, offset int) ([]domain.StockTransfer, bool, error) {
	const q = `
		SELECT id, owner_id, from_warehouse_id, to_warehouse_id, status, notes, created_at, updated_at
		FROM erp_stock_transfers
		WHERE owner_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`

	rows, err := r.pool.QueryContext(ctx, q, ownerID, limit+1, offset)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list stock transfers", logger.Error(err))
		return nil, false, err
	}
	defer rows.Close()

	var transfers []domain.StockTransfer
	for rows.Next() {
		var t domain.StockTransfer
		var status, notes sql.NullString
		var createdAt, updatedAt sql.NullTime

		if err := rows.Scan(
			&t.ID, &t.OwnerID, &t.FromWarehouseID, &t.ToWarehouseID,
			&status, &notes, &createdAt, &updatedAt,
		); err != nil {
			logger.Ctx(ctx).Error("Failed to scan stock transfer row", logger.Error(err))
			return nil, false, err
		}

		if status.Valid {
			t.Status = domain.StockTransferStatus(status.String)
		}
		if notes.Valid {
			t.Notes = notes.String
		}
		if createdAt.Valid {
			t.CreatedAt = createdAt.Time
		}
		if updatedAt.Valid {
			t.UpdatedAt = updatedAt.Time
		}

		transfers = append(transfers, t)
	}

	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(transfers) > limit
	if hasMore {
		transfers = transfers[:limit]
	}

	for i := range transfers {
		items, err := r.getTransferItems(ctx, transfers[i].ID)
		if err != nil {
			return nil, false, err
		}
		transfers[i].Items = items
	}

	return transfers, hasMore, nil
}

func (r *Repository) getTransferItems(ctx context.Context, transferID int64) ([]domain.StockTransferItem, error) {
	rows, err := r.queries.GetTransferItems(ctx, transferID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get transfer items", logger.Error(err))

		return nil, err
	}

	var items []domain.StockTransferItem

	for _, row := range rows {
		var item domain.StockTransferItem
		item.ID = row.ID
		item.TransferID = row.TransferID
		item.ProductID = row.ProductID
		item.Quantity, _ = strconv.ParseFloat(row.Quantity, 64)

		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}

		items = append(items, item)
	}

	return items, nil
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

	qtx := r.queries.WithTx(tx)

	res, err := qtx.InsertStockTransfer(ctx, sqlc.InsertStockTransferParams{
		OwnerID:         ownerID,
		FromWarehouseID: input.FromWarehouseID,
		ToWarehouseID:   input.ToWarehouseID,
		Notes:           nullString(input.Notes),
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create stock transfer", logger.Error(err))

		return 0, err
	}

	transferID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, item := range input.Items {
		_, err = qtx.InsertStockTransferItem(ctx, sqlc.InsertStockTransferItemParams{
			TransferID: transferID,
			ProductID:  item.ProductID,
			Quantity:   strconv.FormatFloat(item.Quantity, 'f', -1, 64),
		})
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
	err := r.queries.UpdateStockTransferStatus(ctx, sqlc.UpdateStockTransferStatusParams{
		Status: sqlc.NullErpStockTransfersStatus{
			ErpStockTransfersStatus: sqlc.ErpStockTransfersStatus(status),
			Valid:                   true,
		},
		OwnerID: ownerID,
		ID:      id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update stock transfer status", logger.Error(err))

		return err
	}

	return nil
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
