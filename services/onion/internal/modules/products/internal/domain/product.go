package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrProductNotFound = errors.New("product not found")

type Product struct {
	ID                int64     `json:"id"`
	OwnerID           int64     `json:"ownerId"`
	CategoryID        int64     `json:"categoryId"`
	SupplierID        int64     `json:"supplierId"`
	Name              string    `json:"name"`
	SKU               string    `json:"sku"`
	Barcode           string    `json:"barcode"`
	Description       string    `json:"description"`
	Unit              string    `json:"unit"`
	CostPrice         float64   `json:"costPrice"`
	SalePrice         float64   `json:"salePrice"`
	StockQuantity     float64   `json:"stockQuantity"`
	LowStockThreshold float64   `json:"lowStockThreshold"`
	ReorderQuantity   float64   `json:"reorderQuantity"`
	IsActive          bool      `json:"isActive"`
	TrackInventory    bool      `json:"trackInventory"`
	CreatedAt         time.Time `json:"createdAt"`
	UpdatedAt         time.Time `json:"updatedAt"`
}

func (p *Product) FromRow(row *sql.Rows) error {
	var (
		categoryID        sql.NullInt64
		supplierID        sql.NullInt64
		sku               sql.NullString
		barcode           sql.NullString
		description       sql.NullString
		unit              sql.NullString
		lowStockThreshold sql.NullFloat64
		reorderQuantity   sql.NullFloat64
	)

	err := row.Scan(
		&p.ID,
		&p.OwnerID,
		&categoryID,
		&supplierID,
		&p.Name,
		&sku,
		&barcode,
		&description,
		&unit,
		&p.CostPrice,
		&p.SalePrice,
		&p.StockQuantity,
		&lowStockThreshold,
		&reorderQuantity,
		&p.IsActive,
		&p.TrackInventory,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if categoryID.Valid {
		p.CategoryID = categoryID.Int64
	}

	if supplierID.Valid {
		p.SupplierID = supplierID.Int64
	}

	if sku.Valid {
		p.SKU = sku.String
	}

	if barcode.Valid {
		p.Barcode = barcode.String
	}

	if description.Valid {
		p.Description = description.String
	}

	if unit.Valid {
		p.Unit = unit.String
	}

	if lowStockThreshold.Valid {
		p.LowStockThreshold = lowStockThreshold.Float64
	}

	if reorderQuantity.Valid {
		p.ReorderQuantity = reorderQuantity.Float64
	}

	return nil
}

type ProductInput struct {
	CategoryID        int64   `json:"categoryId"`
	SupplierID        int64   `json:"supplierId"`
	Name              string  `json:"name"`
	SKU               string  `json:"sku"`
	Barcode           string  `json:"barcode"`
	Description       string  `json:"description"`
	Unit              string  `json:"unit"`
	CostPrice         float64 `json:"costPrice"`
	SalePrice         float64 `json:"salePrice"`
	LowStockThreshold float64 `json:"lowStockThreshold"`
	ReorderQuantity   float64 `json:"reorderQuantity"`
	IsActive          bool    `json:"isActive"`
	TrackInventory    bool    `json:"trackInventory"`
}
