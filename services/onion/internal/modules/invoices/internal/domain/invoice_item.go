package domain

import (
	"database/sql"
)

type InvoiceItem struct {
	ID          int64   `json:"id"`
	InvoiceID   int64   `json:"invoiceId"`
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unitPrice"`
	TaxRate     float64 `json:"taxRate"`
	DiscountPct float64 `json:"discountPct"`
	Total       float64 `json:"total"`
	SortOrder   int     `json:"sortOrder"`
}

func (item *InvoiceItem) FromRow(row *sql.Rows) error {
	var (
		unit sql.NullString
	)

	err := row.Scan(
		&item.ID,
		&item.InvoiceID,
		&item.Description,
		&item.Quantity,
		&unit,
		&item.UnitPrice,
		&item.TaxRate,
		&item.DiscountPct,
		&item.Total,
		&item.SortOrder,
	)
	if err != nil {
		return err
	}

	if unit.Valid {
		item.Unit = unit.String
	} else {
		item.Unit = "unit"
	}

	return nil
}

type InvoiceItemInput struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unitPrice"`
	TaxRate     float64 `json:"taxRate"`
	DiscountPct float64 `json:"discountPct"`
	SortOrder   int     `json:"sortOrder"`
}

// CalculateTotal computes the line total after per-item discount and tax.
func (i InvoiceItemInput) CalculateTotal() float64 {
	lineBase := i.Quantity * i.UnitPrice
	discountAmt := lineBase * (i.DiscountPct / 100)
	afterDiscount := lineBase - discountAmt
	taxAmt := afterDiscount * (i.TaxRate / 100)

	return afterDiscount + taxAmt
}
