package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrPurchaseOrderNotFound = errors.New("purchase order not found")

type PurchaseOrderStatus string

const (
	POStatusDraft     PurchaseOrderStatus = "draft"
	POStatusSent      PurchaseOrderStatus = "sent"
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
	PurchaseOrderID  int64   `json:"purchaseOrderId"`
	ProductID        int64   `json:"productId"`
	QuantityOrdered  float64 `json:"quantityOrdered"`
	QuantityReceived float64 `json:"quantityReceived"`
	UnitCost         float64 `json:"unitCost"`
	TotalCost        float64 `json:"totalCost"`
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
	ProductID       int64   `json:"productId"`
	QuantityOrdered float64 `json:"quantityOrdered"`
	UnitCost        float64 `json:"unitCost"`
}

type PurchaseOrder struct {
	ID               int64               `json:"id"`
	OwnerID          int64               `json:"ownerId"`
	SupplierID       int64               `json:"supplierId"`
	PONumber         string              `json:"poNumber"`
	Status           PurchaseOrderStatus `json:"status"`
	Subtotal         float64             `json:"subtotal"`
	TaxAmount        float64             `json:"taxAmount"`
	Total            float64             `json:"total"`
	ExpectedDelivery *time.Time          `json:"expectedDelivery"`
	ReceivedAt       *time.Time          `json:"receivedAt"`
	Notes            string              `json:"notes"`
	CreatedBy        int64               `json:"createdBy"`
	CreatedAt        time.Time           `json:"createdAt"`
	UpdatedAt        time.Time           `json:"updatedAt"`
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
	SupplierID       int64                    `json:"supplierId"`
	PONumber         string                   `json:"poNumber"`
	TaxAmount        float64                  `json:"taxAmount"`
	ExpectedDelivery string                   `json:"expectedDelivery"`
	Notes            string                   `json:"notes"`
	Items            []PurchaseOrderItemInput `json:"items"`
}
