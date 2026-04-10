package infrastructure

import (
	"context"
	"database/sql"

	"github.com/voxire/onion/internal/modules/pos_kds/internal/domain"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	pool *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	return &Repository{pool: dbclient.Pool()}
}

func (r *Repository) ListActiveOrders(ctx context.Context, ownerID int64) ([]domain.KDSOrder, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, order_number, type, status, COALESCE(table_id,0),
		        COALESCE(customer_name,''), COALESCE(kitchen_notes,''), created_at
		 FROM pos_orders
		 WHERE owner_id = ? AND status IN ('confirmed','preparing','ready')
		 ORDER BY created_at ASC`, ownerID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list KDS orders", logger.Error(err))
		return nil, err
	}
	defer rows.Close()

	orders := make([]domain.KDSOrder, 0)
	for rows.Next() {
		var o domain.KDSOrder
		if err := rows.Scan(&o.ID, &o.OrderNumber, &o.Type, &o.Status, &o.TableID, &o.CustomerName, &o.KitchenNotes, &o.CreatedAt); err != nil {
			logger.Ctx(ctx).Error("Failed to scan KDS order", logger.Error(err))
			return nil, err
		}

		items, itemErr := r.getOrderItems(ctx, o.ID)
		if itemErr != nil {
			return nil, itemErr
		}
		o.Items = items
		orders = append(orders, o)
	}
	return orders, rows.Err()
}

func (r *Repository) getOrderItems(ctx context.Context, orderID int64) ([]domain.KDSItem, error) {
	rows, err := r.pool.QueryContext(ctx,
		`SELECT id, order_id, name, quantity, COALESCE(notes,''), status
		 FROM pos_order_items WHERE order_id = ? ORDER BY id`, orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.KDSItem, 0)
	for rows.Next() {
		var item domain.KDSItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.Name, &item.Quantity, &item.Notes, &item.Status); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateItemStatus(ctx context.Context, input domain.UpdateItemStatusInput) error {
	_, err := r.pool.ExecContext(ctx,
		`UPDATE pos_order_items SET status = ? WHERE order_id = ? AND id = ?`,
		input.Status, input.OrderID, input.ItemID,
	)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update KDS item status", logger.Error(err))
		return err
	}
	return nil
}
