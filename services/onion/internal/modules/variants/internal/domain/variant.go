package domain

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrVariantNotFound        = errors.New("variant not found")
	ErrVariantProductNotFound = errors.New("product not found in business")
)

type Variant struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"owner_id"`
	BusinessID    int64     `json:"business_id"`
	ProductID     int64     `json:"product_id"`
	Name          string    `json:"name"`
	SKU           string    `json:"sku"`
	Barcode       string    `json:"barcode"`
	CostPrice     float64   `json:"cost_price"`
	SalePrice     float64   `json:"sale_price"`
	StockQuantity float64   `json:"stock_quantity"`
	Attributes    string    `json:"attributes"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (v *Variant) FromRow(row *sql.Rows) error {
	var (
		sku        sql.NullString
		barcode    sql.NullString
		attributes sql.NullString
	)

	err := row.Scan(
		&v.ID,
		&v.OwnerID,
		&v.BusinessID,
		&v.ProductID,
		&v.Name,
		&sku,
		&barcode,
		&v.CostPrice,
		&v.SalePrice,
		&v.StockQuantity,
		&attributes,
		&v.IsActive,
		&v.CreatedAt,
		&v.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if sku.Valid {
		v.SKU = sku.String
	}

	if barcode.Valid {
		v.Barcode = barcode.String
	}

	if attributes.Valid {
		v.Attributes = attributes.String
	}

	return nil
}

type VariantInput struct {
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	Barcode       string  `json:"barcode"`
	CostPrice     float64 `json:"cost_price"`
	SalePrice     float64 `json:"sale_price"`
	StockQuantity float64 `json:"stock_quantity"`
	Attributes    string  `json:"attributes"`
	IsActive      bool    `json:"is_active"`
}
