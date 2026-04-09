package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrExpenseNotFound = errors.New("expense not found")

type Expense struct {
	ID              int64     `json:"id"`
	OwnerID         int64     `json:"ownerId"`
	CategoryID      int64     `json:"categoryId"`
	Description     string    `json:"description"`
	Amount          float64   `json:"amount"`
	Currency        string    `json:"currency"`
	ExpenseDate     time.Time `json:"expenseDate"`
	Vendor          string    `json:"vendor"`
	PaymentMethod   string    `json:"paymentMethod"`
	ReceiptURL      string    `json:"receiptUrl"`
	IsTaxDeductible bool      `json:"isTaxDeductible"`
	TaxAmount       float64   `json:"taxAmount"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

func (e *Expense) FromRow(row *sql.Rows) error {
	var (
		categoryID    sql.NullInt64
		currency      sql.NullString
		vendor        sql.NullString
		paymentMethod sql.NullString
		receiptURL    sql.NullString
		notes         sql.NullString
	)

	err := row.Scan(
		&e.ID,
		&e.OwnerID,
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

	return nil
}

type ExpenseInput struct {
	CategoryID      int64   `json:"categoryId"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	ExpenseDate     string  `json:"expenseDate"`
	Vendor          string  `json:"vendor"`
	PaymentMethod   string  `json:"paymentMethod"`
	ReceiptURL      string  `json:"receiptUrl"`
	IsTaxDeductible bool    `json:"isTaxDeductible"`
	TaxAmount       float64 `json:"taxAmount"`
	Notes           string  `json:"notes"`
}
