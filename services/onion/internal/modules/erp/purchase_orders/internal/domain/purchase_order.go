package domain

import (
	"errors"
	"time"
)

var (
	ErrInvalidPurchaseOrder           = errors.New("invalid purchase order")
	ErrPurchaseOrderNotFound          = errors.New("purchase order not found")
	ErrPurchaseOrderAlreadyApproved   = errors.New("purchase order already approved")
	ErrPurchaseOrderAlreadyReceived   = errors.New("purchase order already received")
	ErrPurchaseOrderSelfApproval      = errors.New("purchase order creator cannot approve the purchase order")
	ErrPurchaseOrderInvalidTransition = errors.New("purchase order cannot be approved from its current status")
	ErrPurchaseOrderReferenceNotFound = errors.New("purchase order reference not found in business")
	ErrInvalidReceivedQuantity        = errors.New("invalid received quantity")
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
	ID               int64
	PurchaseOrderID  int64
	ProductID        int64
	QuantityOrdered  float64
	QuantityReceived float64
	UnitCost         float64
	TotalCost        float64
}

type PurchaseOrderItemInput struct {
	ProductID       int64
	QuantityOrdered float64
	UnitCost        float64
}

type PurchaseOrder struct {
	ID               int64
	OwnerID          int64
	BusinessID       int64
	SupplierID       int64
	PONumber         string
	Status           PurchaseOrderStatus
	Subtotal         float64
	TaxAmount        float64
	Total            float64
	ExpectedDelivery *time.Time
	ReceivedAt       *time.Time
	Notes            string
	CreatedBy        int64
	CreatedAt        time.Time
	UpdatedAt        time.Time
	Items            []PurchaseOrderItem
}

type PurchaseOrderInput struct {
	SupplierID       int64
	PONumber         string
	TaxAmount        float64
	ExpectedDelivery string
	Notes            string
	Items            []PurchaseOrderItemInput
}
