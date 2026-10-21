package domain

import (
	"errors"
	"time"

	"github.com/voxire/onion/internal/platform/money"
)

var (
	ErrBillNotFound            = errors.New("bill not found")
	ErrBillReferenceNotFound   = errors.New("bill reference not found in business")
	ErrInvalidStatusTransition = errors.New("invalid bill status transition")
	ErrPaidBillImmutable       = errors.New("paid bill is immutable")
)

type Bill struct {
	ID              int64
	OwnerID         int64
	BusinessID      int64
	SupplierID      int64
	SupplierName    string
	PurchaseOrderID *int64
	Amount          money.Amount
	Currency        string
	DueDate         time.Time
	Status          string
	Description     string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
type BillInput struct {
	SupplierID      int64
	SupplierName    string
	PurchaseOrderID *int64
	Amount          money.Amount
	Currency        string
	DueDate         string
	Status          string
	Description     string
}

// IsPaid reports whether the bill has been fully paid.
func (b *Bill) IsPaid() bool {
	return b.Status == "paid"
}

// ValidateBillStatusTransition enforces the bill lifecycle. Bills start
// pending, may become overdue or paid, and overdue bills may subsequently be
// marked paid. Paid is terminal.
func ValidateBillStatusTransition(current, next string) error {
	if current == next {
		return nil
	}
	if current == "pending" && (next == "overdue" || next == "paid") {
		return nil
	}
	if current == "overdue" && next == "paid" {
		return nil
	}
	return ErrInvalidStatusTransition
}

// IsOverdue reports whether the bill is past its due date and not paid.
func (b *Bill) IsOverdue() bool {
	if b.IsPaid() {
		return false
	}
	return time.Now().After(b.DueDate)
}

// DaysUntilDue returns the number of full days remaining until the bill's due
// date. Negative values mean the bill is already past due.
func (b *Bill) DaysUntilDue() int {
	return int(time.Until(b.DueDate).Hours() / 24)
}
