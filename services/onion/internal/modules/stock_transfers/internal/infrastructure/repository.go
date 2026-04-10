package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_transfers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()

	return &Repository{
		queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (r *Repository) ListStockTransfers(ctx context.Context, ownerID int64) ([]domain.StockTransfer, error) {
	rows, err := r.queries.ListStockTransfers(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list stock transfers", logger.Error(err))

		return nil, err
	}

	var transfers []domain.StockTransfer

	for _, row := range rows {
		transfers = append(transfers, stockTransferFromRow(row))
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

func stockTransferFromRow(row sqlc.ErpStockTransfer) domain.StockTransfer {
	var t domain.StockTransfer
	t.ID = row.ID
	t.OwnerID = row.OwnerID
	t.FromWarehouseID = row.FromWarehouseID
	t.ToWarehouseID = row.ToWarehouseID

	if row.Status.Valid {
		t.Status = domain.StockTransferStatus(row.Status.ErpStockTransfersStatus)
	}

	if row.Notes.Valid {
		t.Notes = row.Notes.String
	}

	if row.CreatedAt.Valid {
		t.CreatedAt = row.CreatedAt.Time
	}

	if row.UpdatedAt.Valid {
		t.UpdatedAt = row.UpdatedAt.Time
	}

	return t
}

func nullString(s string) sql.NullString {
	return sql.NullString{String: s, Valid: s != ""}
}
