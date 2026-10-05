package application

import (
	"context"

	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/domain"
)

type repository interface {
	GetByID(ctx context.Context, businessID, id int64) (domain.PurchaseOrder, error)
	ListPurchaseOrders(ctx context.Context, businessID int64, status string, limit, offset int32) ([]domain.PurchaseOrder, error)
	CreatePurchaseOrder(ctx context.Context, businessID int64, input domain.PurchaseOrderInput, createdBy int64) (int64, error)
	GetPurchaseOrderStatus(ctx context.Context, businessID, id int64) (string, error)
	ApprovePurchaseOrder(ctx context.Context, businessID, id int64) error
	UpdatePurchaseOrderStatus(ctx context.Context, businessID, id int64, status string) error
	ReceivePurchaseOrder(ctx context.Context, businessID, poID, createdBy int64, submitted []domain.PurchaseOrderItem) error
}
