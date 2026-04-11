package infrastructure

import (
	"context"
	"database/sql"
	"strconv"

	"github.com/voxire/onion/internal/modules/store_orders/internal/domain"
	"github.com/voxire/onion/internal/modules/store_orders/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/database/dbclient"
	"github.com/voxire/onion/internal/platform/logger"
)

type Repository struct {
	queries *sqlc.Queries
	pool    *sql.DB
}

func NewRepository(_ context.Context) *Repository {
	pool := dbclient.Pool()
	return &Repository{queries: sqlc.New(pool), pool: pool}
}

func (r *Repository) List(ctx context.Context, ownerID int64) ([]domain.StoreOrder, error) {
	rows, err := r.queries.ListStoreOrders(ctx, ownerID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list store orders", logger.Error(err))
		return nil, err
	}
	items := make([]domain.StoreOrder, 0, len(rows))
	for _, row := range rows {
		items = append(items, orderFromRow(row))
	}
	return items, nil
}

func (r *Repository) GetDetail(ctx context.Context, ownerID, orderID int64) (domain.StoreOrderDetail, error) {
	order, err := r.queries.GetStoreOrder(ctx, sqlc.GetStoreOrderParams{OwnerID: ownerID, ID: orderID})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to get store order", logger.Error(err))
		return domain.StoreOrderDetail{}, err
	}
	itemRows, err := r.queries.ListStoreOrderItems(ctx, orderID)
	if err != nil {
		logger.Ctx(ctx).Error("Failed to list store order items", logger.Error(err))
		return domain.StoreOrderDetail{}, err
	}
	items := make([]domain.StoreOrderItem, 0, len(itemRows))
	for _, row := range itemRows {
		items = append(items, itemFromRow(row))
	}
	return domain.StoreOrderDetail{Order: orderFromRow(order), Items: items}, nil
}

func (r *Repository) Create(ctx context.Context, ownerID int64, input domain.StoreOrderInput) (int64, error) {
	result, err := r.queries.CreateStoreOrder(ctx, sqlc.CreateStoreOrderParams{
		OwnerID:         ownerID,
		CustomerEmail:   sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerName:    sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerPhone:   sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		Status:          sqlc.NullStoreOrdersStatus{StoreOrdersStatus: sqlc.StoreOrdersStatus(input.Status), Valid: input.Status != ""},
		Subtotal:        strconv.FormatFloat(input.Subtotal, 'f', 2, 64),
		Tax:             strconv.FormatFloat(input.Tax, 'f', 2, 64),
		ShippingCost:    strconv.FormatFloat(input.ShippingCost, 'f', 2, 64),
		Total:           strconv.FormatFloat(input.Total, 'f', 2, 64),
		ShippingAddress: nullJSON(input.ShippingAddress),
		Notes:           sql.NullString{String: input.Notes, Valid: input.Notes != ""},
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to create store order", logger.Error(err))
		return 0, err
	}
	id, _ := result.LastInsertId()
	for _, item := range input.Items {
		_, err := r.queries.CreateStoreOrderItem(ctx, sqlc.CreateStoreOrderItemParams{
			OrderID:   id,
			ProductID: sql.NullInt64{Int64: item.ProductID, Valid: item.ProductID != 0},
			Name:      item.Name,
			Quantity:  int32(item.Quantity),
			Price:     strconv.FormatFloat(item.Price, 'f', 2, 64),
			Total:     strconv.FormatFloat(item.Total, 'f', 2, 64),
		})
		if err != nil {
			logger.Ctx(ctx).Error("Failed to create store order item", logger.Error(err))
			return 0, err
		}
	}
	return id, nil
}

func (r *Repository) Update(ctx context.Context, ownerID, id int64, input domain.StoreOrderInput) error {
	err := r.queries.UpdateStoreOrder(ctx, sqlc.UpdateStoreOrderParams{
		CustomerEmail:   sql.NullString{String: input.CustomerEmail, Valid: input.CustomerEmail != ""},
		CustomerName:    sql.NullString{String: input.CustomerName, Valid: input.CustomerName != ""},
		CustomerPhone:   sql.NullString{String: input.CustomerPhone, Valid: input.CustomerPhone != ""},
		Status:          sqlc.NullStoreOrdersStatus{StoreOrdersStatus: sqlc.StoreOrdersStatus(input.Status), Valid: input.Status != ""},
		Subtotal:        strconv.FormatFloat(input.Subtotal, 'f', 2, 64),
		Tax:             strconv.FormatFloat(input.Tax, 'f', 2, 64),
		ShippingCost:    strconv.FormatFloat(input.ShippingCost, 'f', 2, 64),
		Total:           strconv.FormatFloat(input.Total, 'f', 2, 64),
		ShippingAddress: nullJSON(input.ShippingAddress),
		Notes:           sql.NullString{String: input.Notes, Valid: input.Notes != ""},
		OwnerID:         ownerID,
		ID:              id,
	})
	if err != nil {
		logger.Ctx(ctx).Error("Failed to update store order", logger.Error(err))
		return err
	}
	return nil
}

func orderFromRow(row sqlc.StoreOrder) domain.StoreOrder {
	o := domain.StoreOrder{ID: row.ID, OwnerID: row.OwnerID}
	if row.CustomerEmail.Valid { o.CustomerEmail = row.CustomerEmail.String }
	if row.CustomerName.Valid { o.CustomerName = row.CustomerName.String }
	if row.CustomerPhone.Valid { o.CustomerPhone = row.CustomerPhone.String }
	if row.Status.Valid { o.Status = string(row.Status.StoreOrdersStatus) }
	o.Subtotal, _ = strconv.ParseFloat(row.Subtotal, 64)
	o.Tax, _ = strconv.ParseFloat(row.Tax, 64)
	o.ShippingCost, _ = strconv.ParseFloat(row.ShippingCost, 64)
	o.Total, _ = strconv.ParseFloat(row.Total, 64)
	if row.ShippingAddress != nil { o.ShippingAddress = string(row.ShippingAddress) }
	if row.Notes.Valid { o.Notes = row.Notes.String }
	if row.CreatedAt.Valid { o.CreatedAt = row.CreatedAt.Time }
	if row.UpdatedAt.Valid { o.UpdatedAt = row.UpdatedAt.Time }
	return o
}

func itemFromRow(row sqlc.StoreOrderItem) domain.StoreOrderItem {
	i := domain.StoreOrderItem{ID: row.ID, OrderID: row.OrderID, Name: row.Name, Quantity: int(row.Quantity)}
	if row.ProductID.Valid { i.ProductID = row.ProductID.Int64 }
	i.Price, _ = strconv.ParseFloat(row.Price, 64)
	i.Total, _ = strconv.ParseFloat(row.Total, 64)
	if row.CreatedAt.Valid { i.CreatedAt = row.CreatedAt.Time }
	return i
}

func nullJSON(s string) []byte {
	if s == "" { return nil }
	return []byte(s)
}
