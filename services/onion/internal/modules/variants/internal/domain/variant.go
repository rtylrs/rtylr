package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrVariantNotFound = errors.New("variant not found")

type Variant struct {
	ID            int64     `json:"id"`
	OwnerID       int64     `json:"ownerId"`
	ProductID     int64     `json:"productId"`
	Name          string    `json:"name"`
	SKU           string    `json:"sku"`
	Barcode       string    `json:"barcode"`
	CostPrice     float64   `json:"costPrice"`
	SalePrice     float64   `json:"salePrice"`
	StockQuantity float64   `json:"stockQuantity"`
	Attributes    string    `json:"attributes"`
	IsActive      bool      `json:"isActive"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
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
	ProductID     int64   `json:"productId"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	Barcode       string  `json:"barcode"`
	CostPrice     float64 `json:"costPrice"`
	SalePrice     float64 `json:"salePrice"`
	StockQuantity float64 `json:"stockQuantity"`
	Attributes    string  `json:"attributes"`
	IsActive      bool    `json:"isActive"`
}
