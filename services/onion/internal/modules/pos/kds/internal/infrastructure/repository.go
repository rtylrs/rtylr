package infrastructure

import (
	"context"
	"database/sql"
	"errors"

	"github.com/voxire/onion/internal/modules/pos/kds/internal/domain"
	"github.com/voxire/onion/internal/modules/pos/kds/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListActiveOrders(ctx context.Context, businessID int64) ([]domain.KDSOrder, error) {
	rows, err := r.queries.ListActiveKDSOrders(ctx, businessID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list KDS orders", logger.Error(err))
		return nil, err
	}

	orderIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		orderIDs = append(orderIDs, row.ID)
	}

	// Batch-fetch all items for the active orders in a single query and group
	// them in memory, avoiding one GetKDSOrderItems round-trip per order (N+1).
	itemsByOrder, err := r.getItemsForOrders(ctx, orderIDs)
	if err != nil {
		return nil, err
	}

	orders := make([]domain.KDSOrder, 0, len(rows))
	for _, row := range rows {
		o := domain.KDSOrder{
			ID:           row.ID,
			OrderNumber:  row.OrderNumber,
			Type:         string(row.Type.PosOrdersType),
			Status:       string(row.Status.PosOrdersStatus),
			TableID:      row.TableID.Int64,
			CustomerName: row.CustomerName.String,
			KitchenNotes: row.KitchenNotes.String,
			CreatedAt:    row.CreatedAt.Time,
			OwnerID:      row.OwnerID,
			BusinessID:   row.BusinessID,
		}
		o.Items = itemsByOrder[o.ID]
		orders = append(orders, o)
	}
	return orders, nil
}

// getItemsForOrders fetches order items for many orders in one query, returning
// them grouped by order id. Mirrors the column set of GetKDSOrderItems.
func (r *Repository) getItemsForOrders(ctx context.Context, orderIDs []int64) (map[int64][]domain.KDSItem, error) {
	itemsByOrder := make(map[int64][]domain.KDSItem, len(orderIDs))
	if len(orderIDs) == 0 {
		return itemsByOrder, nil
	}

	rows, err := r.queries.GetKDSOrderItemsForOrders(ctx, orderIDs)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to batch fetch KDS order items", logger.Error(err))
		return nil, err
	}

	for _, row := range rows {
		itemsByOrder[row.OrderID] = append(itemsByOrder[row.OrderID], domain.KDSItem{
			ID:       row.ID,
			OrderID:  row.OrderID,
			Name:     row.Name,
			Quantity: sqlutil.DecimalToFloat(row.Quantity),
			Notes:    row.Notes.String,
			Status:   string(row.Status.PosOrderItemsStatus),
		})
	}
	return itemsByOrder, nil
}

func (r *Repository) UpdateItemStatus(ctx context.Context, input domain.UpdateItemStatusInput) error {
	normStatus, err := domain.NormalizeItemStatus(input.Status)
	if err != nil {
		return err
	}

	tx, err := r.pool.BeginTx(ctx, nil)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to begin transaction for KDS item update", logger.Error(err))
		return err
	}
	defer tx.Rollback()

	var orderID int64
	if err = tx.QueryRowContext(ctx, `
		SELECT oi.order_id FROM pos_order_items oi
		JOIN pos_orders o ON o.id = oi.order_id
		WHERE oi.id = ? AND o.business_id = ?
		FOR UPDATE`, input.ItemID, input.BusinessID).Scan(&orderID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrItemNotFound
		}
		logger.Ctx(ctx).Error("Failed to resolve KDS item order", logger.Error(err))
		return err
	}

	if _, err = tx.ExecContext(ctx,
		`UPDATE pos_order_items SET status = ? WHERE id = ? AND order_id = ?`,
		normStatus, input.ItemID, orderID); err != nil {
		logger.Ctx(ctx).Error("Failed to update KDS item status", logger.Error(err))
		return err
	}

	rows, err := tx.QueryContext(ctx, `SELECT status FROM pos_order_items WHERE order_id = ?`, orderID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to read KDS item statuses", logger.Error(err))
		return err
	}
	var statuses []string
	for rows.Next() {
		var st sql.NullString
		if scanErr := rows.Scan(&st); scanErr != nil {
			rows.Close()
			return scanErr
		}
		if st.Valid {
			statuses = append(statuses, st.String)
		}
	}
	if err = rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()

	if _, err = tx.ExecContext(ctx, `
		UPDATE pos_orders
		SET status = ?, updated_at = NOW()
		WHERE id = ? AND business_id = ?
		  AND status NOT IN ('paid', 'refunded', 'voided', 'cancelled', 'held')`,
		domain.DeriveOrderStatusFromItemStatuses(statuses), orderID, input.BusinessID); err != nil {
		logger.Ctx(ctx).Error("Failed to sync order status from KDS items", logger.Error(err))
		return err
	}

	return tx.Commit()
}

// UpdateItemStatusByItemID updates a KDS item status by item id alone, resolving the
// owning order via join. BusinessID-scoped to prevent cross-tenant mutation.
func (r *Repository) UpdateItemStatusByItemID(ctx context.Context, businessID, itemID int64, status string) error {
	return r.UpdateItemStatus(ctx, domain.UpdateItemStatusInput{
		BusinessID: businessID,
		ItemID:     itemID,
		Status:     status,
	})
}
