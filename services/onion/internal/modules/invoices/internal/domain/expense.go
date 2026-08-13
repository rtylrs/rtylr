package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrExpenseNotFound = errors.New("expense not found")

type Expense struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"owner_id"`
	BusinessID      int64     `json:"business_id"`
	CategoryID      int64     `json:"category_id"`
	Description     string    `json:"description"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	ExpenseDate     time.Time `json:"expense_date"`
	Vendor          string    `json:"vendor"`
	PaymentMethod   string    `json:"payment_method"`
	ReceiptURL      string    `json:"receipt_url"`
	IsTaxDeductible bool      `json:"is_tax_deductible"`
	TaxAmount       float64   `json:"tax_amount"`
	Notes           string    `json:"notes"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

func (e *Expense) FromRow(row *sql.Rows) error {
	var (
		categoryID    sql.NullInt64
		currency      sql.NullString
		vendor        sql.NullString
		paymentMethod sql.NullString
		receiptURL    sql.NullString
		notes         sql.NullString
		status        sql.NullString
	)

	err := row.Scan(
		&e.ID,
		&e.OwnerID,
		&e.BusinessID,
		&categoryID,
		&e.Description,
		&e.Amount,
		&currency,
		&e.ExpenseDate,
		&vendor,
		&paymentMethod,
		&receiptURL,
		&e.IsTaxDeductible,
		&e.TaxAmount,
		&notes,
		&status,
		&e.CreatedAt,
		&e.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if categoryID.Valid {
		e.CategoryID = categoryID.Int64
	}

	if currency.Valid {
		e.Currency = currency.String
	} else {
		e.Currency = "USD"
	}

	if vendor.Valid {
		e.Vendor = vendor.String
	}

	if paymentMethod.Valid {
		e.PaymentMethod = paymentMethod.String
	} else {
		e.PaymentMethod = "cash"
	}

	if receiptURL.Valid {
		e.ReceiptURL = receiptURL.String
	}

	if notes.Valid {
		e.Notes = notes.String
	}
	if status.Valid {
		e.Status = status.String
	}

	return nil
}

type ExpenseInput struct {
	CategoryID      int64   `json:"category_id"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	ExpenseDate     string  `json:"expense_date"`
	Vendor          string  `json:"vendor"`
	PaymentMethod   string  `json:"payment_method"`
	ReceiptURL      string  `json:"receipt_url"`
	IsTaxDeductible bool    `json:"is_tax_deductible"`
	TaxAmount       float64 `json:"tax_amount"`
	Notes           string  `json:"notes"`
}
