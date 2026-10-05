package domain

import (
	"errors"
	"time"
)

var (
	ErrInvoiceNotFound      = errors.New("invoice not found")
	ErrInvoiceNotDraft      = errors.New("invoice is not in draft status")
	ErrInvalidInvoiceDate   = errors.New("due date must be on or after issue date")
	ErrDiscountExceedsTotal = errors.New("discount must not exceed the invoice subtotal")

	ErrInvalidInvoiceStatus = errors.New("invoice status is not a known status")
	// ErrInvoiceStatusNotSettable reports a status-only write aimed at a status
	// that is derived from the recorded payments.
	ErrInvoiceStatusNotSettable = errors.New("invoice status is set by recording a payment")
)

type Invoice struct {
	ID             int64
	OwnerID        int64
	BusinessID     int64
	InvoiceNumber  string
	ClientName     string
	ClientEmail    string
	ClientAddress  string
	ClientTaxID    string
	CustomerID     *int64
	Status         string
	IssueDate      time.Time
	DueDate        time.Time
	PaidDate       *time.Time
	Currency       string
	Subtotal       float64
	DiscountType   string
	DiscountValue  float64
	DiscountAmount float64
	TaxRate        float64
	TaxAmount      float64
	Total          float64
	AmountPaid     float64
	AmountDue      float64
	Notes          string
	Terms          string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Items          []InvoiceItem
	Payments       []Payment
}

type InvoiceInput struct {
	CustomerID    int64
	ClientName    string
	ClientEmail   string
	ClientAddress string
	ClientTaxID   string
	IssueDate     string
	DueDate       string
	Currency      string
	DiscountType  string
	DiscountValue float64
	TaxRate       float64
	Notes         string
	Terms         string
	Items         []InvoiceItemInput
}

type InvoiceUpdateInput struct {
	CustomerID    *int64
	ClientName    string
	ClientEmail   string
	ClientAddress string
	ClientTaxID   string
	IssueDate     string
	DueDate       string
	Currency      string
	DiscountType  string
	DiscountValue float64
	TaxRate       float64
	Notes         string
	Terms         string
	Items         []InvoiceItemInput
}
