package domain

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrInvoiceNotFound    = errors.New("invoice not found")
	ErrInvoiceNotDraft    = errors.New("invoice is not in draft status")
	ErrInvalidInvoiceDate = errors.New("due date must be on or after issue date")
)

type Invoice struct {
	ID             int64         `json:"id"`
	OwnerID        int64         `json:"owner_id"`
	BusinessID     int64         `json:"business_id"`
	InvoiceNumber  string        `json:"invoice_number"`
	ClientName     string        `json:"client_name"`
	ClientEmail    string        `json:"client_email"`
	ClientAddress  string        `json:"client_address"`
	ClientTaxID    string        `json:"client_tax_id"`
	CustomerID     *int64        `json:"customer_id"`
	Status         string        `json:"status"`
	IssueDate      time.Time     `json:"issue_date"`
	DueDate        time.Time     `json:"due_date"`
	PaidDate       *time.Time    `json:"paid_date"`
	Currency       string        `json:"currency"`
	Subtotal       float64       `json:"subtotal"`
	DiscountType   string        `json:"discount_type"`
	DiscountValue  float64       `json:"discount_value"`
	DiscountAmount float64       `json:"discount_amount"`
	TaxRate        float64       `json:"tax_rate"`
	TaxAmount      float64       `json:"tax_amount"`
	Total          float64       `json:"total"`
	AmountPaid     float64       `json:"amount_paid"`
	AmountDue      float64       `json:"amount_due"`
	Notes          string        `json:"notes"`
	Terms          string        `json:"terms"`
	CreatedAt      time.Time     `json:"created_at"`
	UpdatedAt      time.Time     `json:"updated_at"`
	Items          []InvoiceItem `json:"items"`
	Payments       []Payment     `json:"payments"`
}

func (inv *Invoice) FromRow(row *sql.Rows) error {
	var (
		clientEmail   sql.NullString
		clientAddress sql.NullString
		clientTaxID   sql.NullString
		paidDate      sql.NullTime
		currency      sql.NullString
		discountType  sql.NullString
		notes         sql.NullString
		terms         sql.NullString
	)

	err := row.Scan(
		&inv.ID,
		&inv.OwnerID,
		&inv.BusinessID,
		&inv.InvoiceNumber,
		&inv.ClientName,
		&clientEmail,
		&clientAddress,
		&clientTaxID,
		&inv.Status,
		&inv.IssueDate,
		&inv.DueDate,
		&paidDate,
		&currency,
		&inv.Subtotal,
		&discountType,
		&inv.DiscountValue,
		&inv.DiscountAmount,
		&inv.TaxRate,
		&inv.TaxAmount,
		&inv.Total,
		&inv.AmountPaid,
		&inv.AmountDue,
		&notes,
		&terms,
		&inv.CreatedAt,
		&inv.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if clientEmail.Valid {
		inv.ClientEmail = clientEmail.String
	}

	if clientAddress.Valid {
		inv.ClientAddress = clientAddress.String
	}

	if clientTaxID.Valid {
		inv.ClientTaxID = clientTaxID.String
	}

	if paidDate.Valid {
		inv.PaidDate = &paidDate.Time
	}

	if currency.Valid {
		inv.Currency = currency.String
	} else {
		inv.Currency = "USD"
	}

	if discountType.Valid {
		inv.DiscountType = discountType.String
	} else {
		inv.DiscountType = "fixed"
	}

	if notes.Valid {
		inv.Notes = notes.String
	}

	if terms.Valid {
		inv.Terms = terms.String
	}

	return nil
}

type InvoiceInput struct {
	CustomerID    int64              `json:"customer_id"`
	ClientName    string             `json:"client_name"`
	ClientEmail   string             `json:"client_email"`
	ClientAddress string             `json:"client_address"`
	ClientTaxID   string             `json:"client_tax_id"`
	IssueDate     string             `json:"issue_date"`
	DueDate       string             `json:"due_date"`
	Currency      string             `json:"currency"`
	DiscountType  string             `json:"discount_type"`
	DiscountValue float64            `json:"discount_value"`
	TaxRate       float64            `json:"tax_rate"`
	Notes         string             `json:"notes"`
	Terms         string             `json:"terms"`
	Items         []InvoiceItemInput `json:"items"`
}

type InvoiceUpdateInput struct {
	CustomerID    *int64             `json:"customer_id"`
	ClientName    string             `json:"client_name"`
	ClientEmail   string             `json:"client_email"`
	ClientAddress string             `json:"client_address"`
	ClientTaxID   string             `json:"client_tax_id"`
	IssueDate     string             `json:"issue_date"`
	DueDate       string             `json:"due_date"`
	Currency      string             `json:"currency"`
	DiscountType  string             `json:"discount_type"`
	DiscountValue float64            `json:"discount_value"`
	TaxRate       float64            `json:"tax_rate"`
	Notes         string             `json:"notes"`
	Terms         string             `json:"terms"`
	Items         []InvoiceItemInput `json:"items"`
}
