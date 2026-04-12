package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrBillNotFound = errors.New("bill not found")

type Bill struct {
	ID          int64     `json:"id"`
	OwnerID     int64     `json:"owner_id"`
	SupplierID  int64     `json:"supplier_id"`
	Amount      float64   `json:"amount"`
	Currency    string    `json:"currency"`
	DueDate     time.Time `json:"due_date"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (b *Bill) FromRow(row *sql.Rows) error {
	var (
		supplierID  sql.NullInt64
		currency    sql.NullString
		description sql.NullString
	)

	err := row.Scan(
		&b.ID, &b.OwnerID, &supplierID, &b.Amount, &currency,
		&b.DueDate, &b.Status, &description, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if supplierID.Valid {
		b.SupplierID = supplierID.Int64
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
	SupplierID  int64   `json:"supplier_id"`
	Amount      float64 `json:"amount"`
	Currency    string  `json:"currency"`
	DueDate     string  `json:"due_date"`
	Status      string  `json:"status"`
	Description string  `json:"description"`
}
