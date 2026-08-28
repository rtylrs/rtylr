package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"sort"
	"strconv"

	"github.com/voxire/onion/internal/modules/stock_transfers/internal/domain"
	"github.com/voxire/onion/internal/modules/stock_transfers/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/sqlutil"
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

func (r *Repository) ListStockTransfers(ctx context.Context, businessID int64, limit, offset int) ([]domain.StockTransfer, bool, error) {
	rows, err := r.queries.ListStockTransfersPaged(ctx, sqlc.ListStockTransfersPagedParams{
		BusinessID: businessID,
		Limit:      int32(limit + 1), // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
		Offset:     int32(offset),    // #nosec G115 -- Reviewed width conversion uses validated application limits and the established database or contract field width.
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list stock transfers", logger.Error(err))
		return nil, false, err
	}

	transfers := make([]domain.StockTransfer, 0, limit+1)
	for _, row := range rows {
		transfers = append(transfers, stockTransferFromRow(row))
	}

	hasMore := len(transfers) > limit
	if hasMore {
		transfers = transfers[:limit]
	}

	itemsByTransfer, err := r.getTransferItemsBatch(ctx, transfers)
	if err != nil {
		return nil, false, err
	}
	for i := range transfers {
		transfers[i].Items = itemsByTransfer[transfers[i].ID]
	}

	return transfers, hasMore, nil
}

func (r *Repository) GetByID(ctx context.Context, businessID, id int64) (domain.StockTransfer, error) {
	row, err := r.queries.GetStockTransferByID(ctx, sqlc.GetStockTransferByIDParams{
		BusinessID: businessID,
		ID:         id,
	})
	if errors.Is(err, sql.ErrNoRows) {
		return domain.StockTransfer{}, domain.ErrStockTransferNotFound
	}
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get stock transfer by ID", logger.Error(err))
		return domain.StockTransfer{}, err
	}

	transfer := stockTransferFromRow(row)
	itemsByTransfer, err := r.getTransferItemsBatch(ctx, []domain.StockTransfer{transfer})
	if err != nil {
		return domain.StockTransfer{}, err
	}
	transfer.Items = itemsByTransfer[transfer.ID]
	return transfer, nil
}

func stockTransferFromRow(row sqlc.ErpStockTransfer) domain.StockTransfer {
	t := domain.StockTransfer{
		ID:              row.ID,
		OwnerID:         row.OwnerID,
		BusinessID:      row.BusinessID,
		FromWarehouseID: row.FromWarehouseID,
		ToWarehouseID:   row.ToWarehouseID,
	}

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

// getTransferItemsBatch loads all items for the given transfers in a single
// query (WHERE transfer_id IN (...)) and groups them by transfer ID.
func (r *Repository) getTransferItemsBatch(ctx context.Context, transfers []domain.StockTransfer) (map[int64][]domain.StockTransferItem, error) {
	if len(transfers) == 0 {
		return nil, nil
	}

	ids := make([]int64, 0, len(transfers))
	for i := range transfers {
		ids = append(ids, transfers[i].ID)
	}

	rows, err := r.queries.ListTransferItemsByTransferIDs(ctx, ids)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch get transfer items", logger.Error(err))
		return nil, err
	}

	result := make(map[int64][]domain.StockTransferItem, len(transfers))
	for _, row := range rows {
		item := domain.StockTransferItem{
			ID:         row.ID,
			TransferID: row.TransferID,
			ProductID:  row.ProductID,
			Quantity:   sqlutil.ParseDecimal(row.Quantity),
		}
		if row.CreatedAt.Valid {
			item.CreatedAt = row.CreatedAt.Time
		}
		result[item.TransferID] = append(result[item.TransferID], item)
	}

	return result, nil
}

func (r *Repository) CreateStockTransfer(ctx context.Context, businessID int64, input domain.StockTransferInput) (int64, error) {
	if err := domain.ValidateStockTransferInput(input); err != nil {
		return 0, fmt.Errorf("%w: %w", domain.ErrInvalidStockTransfer, err)
	}
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
	if err := validateStockTransferReferences(ctx, qtx, businessID, input); err != nil {
		return 0, err
	}

	res, err := qtx.InsertStockTransfer(ctx, sqlc.InsertStockTransferParams{
		OwnerID:         businessID,
		BusinessID:      businessID,
		FromWarehouseID: input.FromWarehouseID,
		ToWarehouseID:   input.ToWarehouseID,
		Notes:           sqlutil.NullString(input.Notes),
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

func validateStockTransferReferences(ctx context.Context, qtx *sqlc.Queries, businessID int64, input domain.StockTransferInput) error {
	for _, warehouseID := range []int64{input.FromWarehouseID, input.ToWarehouseID} {
		if _, err := qtx.GetWarehouseReference(ctx, sqlc.GetWarehouseReferenceParams{BusinessID: businessID, ID: warehouseID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrStockTransferReferenceNotFound
			}
			return err
		}
	}
	for _, item := range input.Items {
		if _, err := qtx.GetProductReference(ctx, sqlc.GetProductReferenceParams{BusinessID: businessID, ID: item.ProductID}); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrStockTransferReferenceNotFound
			}
			return err
		}
	}
	return nil
}

func (r *Repository) UpdateStockTransferStatus(ctx context.Context, businessID, id int64, status domain.StockTransferStatus) error {
	// Only 'received' moves stock; other transitions are a plain status update.
	if status != domain.StockTransferStatusReceived {
		result, err := r.queries.UpdateStockTransferStatus(ctx, sqlc.UpdateStockTransferStatusParams{
			Status:     sqlc.NullErpStockTransfersStatus{ErpStockTransfersStatus: sqlc.ErpStockTransfersStatus(status), Valid: true},
			BusinessID: businessID,
			ID:         id,
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to update stock transfer status", logger.Error(err))
			return err
		}
		changed, err := statusUpdateChanged(result)
		if err != nil {
			logger.Ctx(ctx).Error("Failed to inspect stock transfer status update", logger.Error(err))
			return err
		}
		if !changed {
			_, err = r.queries.GetStockTransferForUpdate(ctx, sqlc.GetStockTransferForUpdateParams{
				BusinessID: businessID,
				ID:         id,
			})
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrStockTransferNotFound
			}
			if err != nil {
				return err
			}
		}
		return nil
	}

	return r.receiveTransfer(ctx, businessID, id)
}

// receiveTransfer atomically moves each line item's quantity from the source
// warehouse location to the destination location (transfer_out + transfer_in,
// paired by transfer_group), then marks the transfer received. Owner-verified
// and row-locked; idempotent on the (stock_transfer, transfer_out) ledger key.
func (r *Repository) receiveTransfer(ctx context.Context, businessID, id int64) (err error) {
	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	qtx := r.queries.WithTx(tx)

	row, err := qtx.GetStockTransferForUpdate(ctx, sqlc.GetStockTransferForUpdateParams{
		BusinessID: businessID,
		ID:         id,
	})
	if errors.Is(err, sql.ErrNoRows) {
		err = domain.ErrStockTransferNotFound
		return err
	}
	if err != nil {
		return err
	}
	fromWH := row.FromWarehouseID
	toWH := row.ToWarehouseID
	curStatus := string(row.Status.ErpStockTransfersStatus)
	if curStatus == string(domain.StockTransferStatusReceived) {
		err = nil
		return nil // idempotent: already received
	}
	for _, warehouseID := range []int64{fromWH, toWH} {
		if _, lookupErr := qtx.GetWarehouseReference(ctx, sqlc.GetWarehouseReferenceParams{
			BusinessID: businessID,
			ID:         warehouseID,
		}); lookupErr != nil {
			if errors.Is(lookupErr, sql.ErrNoRows) {
				return domain.ErrStockTransferReferenceNotFound
			}
			return lookupErr
		}
	}

	items, err := qtx.GetTransferItems(ctx, id)
	if err != nil {
		return err
	}

	fromLoc, err := stock_core.WarehouseLocationID(ctx, tx, businessID, fromWH)
	if err != nil {
		return err
	}
	toLoc, err := stock_core.WarehouseLocationID(ctx, tx, businessID, toWH)
	if err != nil {
		return err
	}

	type line struct {
		productID int64
		qty       float64
	}
	lines := make([]line, 0, len(items))
	for _, it := range items {
		if it.ProductID <= 0 {
			return domain.ErrInvalidStockTransfer
		}
		q, perr := strconv.ParseFloat(it.Quantity, 64)
		if perr != nil {
			err = perr
			return err
		}
		if q <= 0 || math.IsNaN(q) || math.IsInf(q, 0) {
			return domain.ErrInvalidStockTransfer
		}
		lines = append(lines, line{productID: it.ProductID, qty: q})
	}
	sort.Slice(lines, func(i, j int) bool { return lines[i].productID < lines[j].productID })

	group := id
	for _, l := range lines {
		out := stock_core.Movement{
			OwnerID: businessID, LocationID: fromLoc, ProductID: l.productID,
			EntryType: stock_core.EntryTransferOut, Change: -l.qty,
			ReferenceID: &id, ReferenceType: "stock_transfer", TransferGroup: &group,
			Channel: "erp", Note: "stock transfer",
		}
		in := stock_core.Movement{
			OwnerID: businessID, LocationID: toLoc, ProductID: l.productID,
			EntryType: stock_core.EntryTransferIn, Change: l.qty,
			ReferenceID: &id, ReferenceType: "stock_transfer", TransferGroup: &group,
			Channel: "erp", Note: "stock transfer", AllowNegative: true,
		}
		// Acquire location row locks in ascending location_id order to avoid deadlocks.
		first, second := out, in
		if toLoc < fromLoc {
			first, second = in, out
		}
		if _, _, err = stock_core.ApplyMovementTx(ctx, tx, first); err != nil {
			return err
		}
		if _, _, err = stock_core.ApplyMovementTx(ctx, tx, second); err != nil {
			return err
		}
	}

	if _, err = qtx.UpdateStockTransferStatus(ctx, sqlc.UpdateStockTransferStatusParams{
		Status:     sqlc.NullErpStockTransfersStatus{ErpStockTransfersStatus: sqlc.ErpStockTransfersStatus(domain.StockTransferStatusReceived), Valid: true},
		BusinessID: businessID,
		ID:         id,
	}); err != nil {
		return err
	}

	return tx.Commit()
}

// parseDecimal converts a DB decimal string to float64, logging corrupt
// values instead of silently coercing them to 0.
