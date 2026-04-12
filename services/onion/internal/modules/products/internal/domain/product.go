package domain

import (
	"database/sql"
	"errors"
	"time"
)

var ErrProductNotFound = errors.New("product not found")

type Product struct {
	ID                int64     `json:"id"`
	OwnerID           int64     `json:"owner_id"`
	CategoryID        int64     `json:"category_id"`
	SupplierID        int64     `json:"supplier_id"`
	Name              string    `json:"name"`
	SKU               string    `json:"sku"`
	Barcode           string    `json:"barcode"`
	Description       string    `json:"description"`
	Unit              string    `json:"unit"`
	CostPrice         float64   `json:"cost_price"`
	SalePrice         float64   `json:"sale_price"`
	StockQuantity     float64   `json:"stock_quantity"`
	LowStockThreshold float64   `json:"low_stock_threshold"`
	ReorderQuantity   float64   `json:"reorder_quantity"`
	IsActive          bool      `json:"is_active"`
	TrackInventory    bool      `json:"track_inventory"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
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
	CategoryID        int64   `json:"category_id"`
	SupplierID        int64   `json:"supplier_id"`
	Name              string  `json:"name"`
	SKU               string  `json:"sku"`
	Barcode           string  `json:"barcode"`
	Description       string  `json:"description"`
	Unit              string  `json:"unit"`
	CostPrice         float64 `json:"cost_price"`
	SalePrice         float64 `json:"sale_price"`
	LowStockThreshold float64 `json:"low_stock_threshold"`
	ReorderQuantity   float64 `json:"reorder_quantity"`
	IsActive          bool    `json:"is_active"`
	TrackInventory    bool    `json:"track_inventory"`
}
