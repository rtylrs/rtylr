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
	ID             int64          `json:"id"`
	OwnerID        int64          `json:"ownerId"`
	InvoiceNumber  string         `json:"invoiceNumber"`
	ClientName     string         `json:"clientName"`
	ClientEmail    string         `json:"clientEmail"`
	ClientAddress  string         `json:"clientAddress"`
	ClientTaxID    string         `json:"clientTaxId"`
	Status         string         `json:"status"`
	IssueDate      time.Time      `json:"issueDate"`
	DueDate        time.Time      `json:"dueDate"`
	PaidDate       *time.Time     `json:"paidDate"`
	Currency       string         `json:"currency"`
	Subtotal       float64        `json:"subtotal"`
	DiscountType   string         `json:"discountType"`
	DiscountValue  float64        `json:"discountValue"`
	DiscountAmount float64        `json:"discountAmount"`
	TaxRate        float64        `json:"taxRate"`
	TaxAmount      float64        `json:"taxAmount"`
	Total          float64        `json:"total"`
	AmountPaid     float64        `json:"amountPaid"`
	AmountDue      float64        `json:"amountDue"`
	Notes          string         `json:"notes"`
	Terms          string         `json:"terms"`
	CreatedAt      time.Time      `json:"createdAt"`
	UpdatedAt      time.Time      `json:"updatedAt"`
	Items          []InvoiceItem  `json:"items"`
	Payments       []Payment      `json:"payments"`
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
	ClientName    string             `json:"clientName"`
	ClientEmail   string             `json:"clientEmail"`
	ClientAddress string             `json:"clientAddress"`
	ClientTaxID   string             `json:"clientTaxId"`
	IssueDate     string             `json:"issueDate"`
	DueDate       string             `json:"dueDate"`
	Currency      string             `json:"currency"`
	DiscountType  string             `json:"discountType"`
	DiscountValue float64            `json:"discountValue"`
	TaxRate       float64            `json:"taxRate"`
	Notes         string             `json:"notes"`
	Terms         string             `json:"terms"`
	Items         []InvoiceItemInput `json:"items"`
}

type InvoiceUpdateInput struct {
	ClientName    string             `json:"clientName"`
	ClientEmail   string             `json:"clientEmail"`
	ClientAddress string             `json:"clientAddress"`
	ClientTaxID   string             `json:"clientTaxId"`
	IssueDate     string             `json:"issueDate"`
	DueDate       string             `json:"dueDate"`
	Currency      string             `json:"currency"`
	DiscountType  string             `json:"discountType"`
	DiscountValue float64            `json:"discountValue"`
	TaxRate       float64            `json:"taxRate"`
	Notes         string             `json:"notes"`
	Terms         string             `json:"terms"`
	Items         []InvoiceItemInput `json:"items"`
}
