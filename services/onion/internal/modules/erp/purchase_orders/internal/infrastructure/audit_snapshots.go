package infrastructure

import (
	"time"

	"github.com/voxire/onion/internal/modules/erp/purchase_orders/internal/domain"
)

type purchaseOrderInputSnapshot struct {
	SupplierID       int64                            `json:"supplier_id"`
	PONumber         string                           `json:"po_number"`
	TaxAmount        float64                          `json:"tax_amount"`
	ExpectedDelivery string                           `json:"expected_delivery"`
	Notes            string                           `json:"notes"`
	Items            []purchaseOrderItemInputSnapshot `json:"items"`
}

type purchaseOrderItemInputSnapshot struct {
	ProductID       int64   `json:"product_id"`
	QuantityOrdered float64 `json:"quantity_ordered"`
	UnitCost        float64 `json:"unit_cost"`
}

type purchaseOrderSnapshot struct {
	ID               int64                       `json:"id"`
	OwnerID          int64                       `json:"owner_id"`
	BusinessID       int64                       `json:"business_id"`
	SupplierID       int64                       `json:"supplier_id"`
	PONumber         string                      `json:"po_number"`
	Status           domain.PurchaseOrderStatus  `json:"status"`
	Subtotal         float64                     `json:"subtotal"`
	TaxAmount        float64                     `json:"tax_amount"`
	Total            float64                     `json:"total"`
	ExpectedDelivery *time.Time                  `json:"expected_delivery"`
	ReceivedAt       *time.Time                  `json:"received_at"`
	Notes            string                      `json:"notes"`
	CreatedBy        int64                       `json:"created_by"`
	CreatedAt        time.Time                   `json:"created_at"`
	UpdatedAt        time.Time                   `json:"updated_at"`
	Items            []purchaseOrderItemSnapshot `json:"items"`
}

type purchaseOrderItemSnapshot struct {
	ID               int64   `json:"id"`
	PurchaseOrderID  int64   `json:"purchase_order_id"`
	ProductID        int64   `json:"product_id"`
	QuantityOrdered  float64 `json:"quantity_ordered"`
	QuantityReceived float64 `json:"quantity_received"`
	UnitCost         float64 `json:"unit_cost"`
	TotalCost        float64 `json:"total_cost"`
}

func purchaseOrderInputSnapshotOf(v domain.PurchaseOrderInput) purchaseOrderInputSnapshot {
	out := purchaseOrderInputSnapshot{SupplierID: v.SupplierID, PONumber: v.PONumber, TaxAmount: v.TaxAmount, ExpectedDelivery: v.ExpectedDelivery, Notes: v.Notes}
	if v.Items != nil {
		out.Items = make([]purchaseOrderItemInputSnapshot, len(v.Items))
		for i, item := range v.Items {
			out.Items[i] = purchaseOrderItemInputSnapshot(item)
		}
	}
	return out
}

func purchaseOrderSnapshotOf(v domain.PurchaseOrder) purchaseOrderSnapshot {
	out := purchaseOrderSnapshot{ID: v.ID, OwnerID: v.OwnerID, BusinessID: v.BusinessID, SupplierID: v.SupplierID, PONumber: v.PONumber, Status: v.Status, Subtotal: v.Subtotal, TaxAmount: v.TaxAmount, Total: v.Total, ExpectedDelivery: v.ExpectedDelivery, ReceivedAt: v.ReceivedAt, Notes: v.Notes, CreatedBy: v.CreatedBy, CreatedAt: v.CreatedAt, UpdatedAt: v.UpdatedAt}
	if v.Items != nil {
		out.Items = make([]purchaseOrderItemSnapshot, len(v.Items))
		for i, item := range v.Items {
			out.Items[i] = purchaseOrderItemSnapshot(item)
		}
	}
	return out
}
