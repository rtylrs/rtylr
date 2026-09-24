package domain

import (
	"database/sql"
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
	ID              int64        `json:"id"`
	OwnerID         int64        `json:"owner_id"`
	BusinessID      int64        `json:"business_id"`
	SupplierID      int64        `json:"supplier_id"`
	SupplierName    string       `json:"supplier_name"`
	PurchaseOrderID *int64       `json:"purchase_order_id"`
	Amount          money.Amount `json:"amount"`
	Currency        string       `json:"currency"`
	DueDate         time.Time    `json:"due_date"`
	Status          string       `json:"status"`
	Description     string       `json:"description"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedAt       time.Time    `json:"updated_at"`
}

func (b *Bill) FromRow(row *sql.Rows) error {
	var (
		supplierID  sql.NullInt64
		currency    sql.NullString
		description sql.NullString
	)

	var supplierName sql.NullString
	err := row.Scan(
		&b.ID, &b.OwnerID, &b.BusinessID, &supplierID, &supplierName, &b.Amount, &currency,
		&b.DueDate, &b.Status, &description, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if supplierID.Valid {
		b.SupplierID = supplierID.Int64
	}
	if supplierName.Valid {
		b.SupplierName = supplierName.String
	}
	if currency.Valid {
		b.Currency = currency.String
	}
	if description.Valid {
		b.Description = description.String
	}

	return nil
}

type BillInput struct {
	SupplierID      int64        `json:"supplier_id"`
	SupplierName    string       `json:"supplier_name"`
	PurchaseOrderID *int64       `json:"purchase_order_id"`
	Amount          money.Amount `json:"amount"`
	Currency        string       `json:"currency"`
	DueDate         string       `json:"due_date"`
	Status          string       `json:"status"`
	Description     string       `json:"description"`
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
