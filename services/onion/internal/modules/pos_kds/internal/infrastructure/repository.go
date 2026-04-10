package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/pos_kds/internal/domain"
	"github.com/voxire/onion/internal/modules/pos_kds/internal/infrastructure/persistence/sqlc"
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

func (r *Repository) ListActiveOrders(ctx context.Context, ownerID int64) ([]domain.KDSOrder, error) {
	rows, err := r.queries.ListActiveKDSOrders(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list KDS orders", logger.Error(err))
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
		}

		items, itemErr := r.getOrderItems(ctx, o.ID)
		if itemErr != nil {
			return nil, itemErr
		}
		o.Items = items
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *Repository) getOrderItems(ctx context.Context, orderID int64) ([]domain.KDSItem, error) {
	rows, err := r.queries.GetKDSOrderItems(ctx, orderID)
	if err != nil {
		return nil, err
	}

	items := make([]domain.KDSItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, domain.KDSItem{
			ID:       row.ID,
			OrderID:  row.OrderID,
			Name:     row.Name,
			Quantity: decimalToFloat(row.Quantity),
			Notes:    row.Notes.String,
			Status:   string(row.Status.PosOrderItemsStatus),
		})
	}
	return items, nil
}

func (r *Repository) UpdateItemStatus(ctx context.Context, input domain.UpdateItemStatusInput) error {
	err := r.queries.UpdateKDSItemStatus(ctx, sqlc.UpdateKDSItemStatusParams{
		Status:  sqlc.NullPosOrderItemsStatus{PosOrderItemsStatus: sqlc.PosOrderItemsStatus(input.Status), Valid: true},
		OrderID: input.OrderID,
		ID:      input.ItemID,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update KDS item status", logger.Error(err))
		return err
	}
	return nil
}

func decimalToFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
