package api

import (
	"github.com/voxire/onion/internal/modules/purchase_orders/internal/domain"
	erppurchaseorderv1 "github.com/voxire/voxie/gen/go/erp/purchase_order/v1"
)

func newPurchaseOrderPayload(po domain.PurchaseOrder) *erppurchaseorderv1.PurchaseOrder {
	payload := &erppurchaseorderv1.PurchaseOrder{
		Id:         po.ID,
		OwnerId:    po.OwnerID,
		SupplierId: po.SupplierID,
		PoNumber:   po.PONumber,
		Status:     string(po.Status),
		Subtotal:   po.Subtotal,
		TaxAmount:  po.TaxAmount,
		Total:      po.Total,
		Notes:      po.Notes,
		CreatedBy:  po.CreatedBy,
		CreatedAt:  po.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  po.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		Items:      make([]*erppurchaseorderv1.PurchaseOrderItem, 0, len(po.Items)),
	}

	if po.ExpectedDelivery != nil {
		formatted := po.ExpectedDelivery.Format("2006-01-02")
		payload.ExpectedDelivery = &formatted
	}

	if po.ReceivedAt != nil {
		formatted := po.ReceivedAt.Format("2006-01-02T15:04:05Z07:00")
		payload.ReceivedAt = &formatted
	}

	for _, item := range po.Items {
		payload.Items = append(payload.Items, &erppurchaseorderv1.PurchaseOrderItem{
			Id:               item.ID,
			PurchaseOrderId:  item.PurchaseOrderID,
			ProductId:        item.ProductID,
			QuantityOrdered:  item.QuantityOrdered,
			QuantityReceived: item.QuantityReceived,
			UnitCost:         item.UnitCost,
			TotalCost:        item.TotalCost,
		})
	}

	return payload
}
