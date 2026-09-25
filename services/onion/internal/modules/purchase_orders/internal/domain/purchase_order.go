package domain

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrInvalidPurchaseOrder           = errors.New("invalid purchase order")
	ErrPurchaseOrderNotFound          = errors.New("purchase order not found")
	ErrPurchaseOrderAlreadyApproved   = errors.New("purchase order already approved")
	ErrPurchaseOrderAlreadyReceived   = errors.New("purchase order already received")
	ErrPurchaseOrderSelfApproval      = errors.New("purchase order creator cannot approve the purchase order")
	ErrPurchaseOrderApprovalForbidden = errors.New("user is not authorized to approve purchase orders")
	ErrPurchaseOrderInvalidTransition = errors.New("purchase order cannot be approved from its current status")
	ErrPurchaseOrderReferenceNotFound = errors.New("purchase order reference not found in business")
)

type PurchaseOrderStatus string

const (
	POStatusDraft     PurchaseOrderStatus = "draft"
	POStatusSent      PurchaseOrderStatus = "sent"
	POStatusApproved  PurchaseOrderStatus = "approved"
	POStatusPartial   PurchaseOrderStatus = "partial"
	POStatusReceived  PurchaseOrderStatus = "received"
	POStatusCancelled PurchaseOrderStatus = "cancelled"
)

type InventoryTransactionType string

const (
	InventoryTypePurchase InventoryTransactionType = "purchase"
)

type PurchaseOrderItem struct {
	ID               int64   `json:"id"`
	PurchaseOrderID  int64   `json:"purchase_order_id"`
	ProductID        int64   `json:"product_id"`
	QuantityOrdered  float64 `json:"quantity_ordered"`
	QuantityReceived float64 `json:"quantity_received"`
	UnitCost         float64 `json:"unit_cost"`
	TotalCost        float64 `json:"total_cost"`
}

func (i *PurchaseOrderItem) FromRow(row *sql.Rows) error {
	return row.Scan(
		&i.ID,
		&i.PurchaseOrderID,
		&i.ProductID,
		&i.QuantityOrdered,
		&i.QuantityReceived,
		&i.UnitCost,
		&i.TotalCost,
	)
}

type PurchaseOrderItemInput struct {
	ProductID       int64   `json:"product_id"`
	QuantityOrdered float64 `json:"quantity_ordered"`
	UnitCost        float64 `json:"unit_cost"`
}

type PurchaseOrder struct {
	ID               int64               `json:"id"`
	OwnerID          int64               `json:"owner_id"`
	BusinessID       int64               `json:"business_id"`
	SupplierID       int64               `json:"supplier_id"`
	PONumber         string              `json:"po_number"`
	Status           PurchaseOrderStatus `json:"status"`
	Subtotal         float64             `json:"subtotal"`
	TaxAmount        float64             `json:"tax_amount"`
	Total            float64             `json:"total"`
	ExpectedDelivery *time.Time          `json:"expected_delivery"`
	ReceivedAt       *time.Time          `json:"received_at"`
	Notes            string              `json:"notes"`
	CreatedBy        int64               `json:"created_by"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`
	Items            []PurchaseOrderItem `json:"items"`
}

func (po *PurchaseOrder) FromRow(row *sql.Rows) error {
	var (
		expectedDelivery sql.NullTime
		receivedAt       sql.NullTime
		notes            sql.NullString
		createdBy        sql.NullInt64
	)

	err := row.Scan(
		&po.ID,
		&po.OwnerID,
		&po.SupplierID,
		&po.PONumber,
		&po.Status,
		&po.Subtotal,
		&po.TaxAmount,
		&po.Total,
		&expectedDelivery,
		&receivedAt,
		&notes,
		&createdBy,
		&po.CreatedAt,
		&po.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if expectedDelivery.Valid {
		po.ExpectedDelivery = &expectedDelivery.Time
	}
	if receivedAt.Valid {
		po.ReceivedAt = &receivedAt.Time
	}
	if notes.Valid {
		po.Notes = notes.String
	}
	if createdBy.Valid {
		po.CreatedBy = createdBy.Int64
	}

	return nil
}

type PurchaseOrderInput struct {
	SupplierID       int64                    `json:"supplier_id"`
	PONumber         string                   `json:"po_number"`
	TaxAmount        float64                  `json:"tax_amount"`
	ExpectedDelivery string                   `json:"expected_delivery"`
	Notes            string                   `json:"notes"`
	Items            []PurchaseOrderItemInput `json:"items"`
}
