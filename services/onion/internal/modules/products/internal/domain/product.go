package domain

import (
	"database/sql"
	"errors"
	"time"
)

var (
	ErrProductNotFound          = errors.New("product not found")
	ErrProductAlreadyExists     = errors.New("product SKU or barcode already exists")
	ErrProductReferenceNotFound = errors.New("product reference not found in business")
)

type Product struct {
	ID                int64     `json:"id"`
	OwnerID           int64     `json:"owner_id"`
	BusinessID        int64     `json:"business_id"`
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
	BrandID           int64     `json:"brand_id"`
	CompareAtPrice    float64   `json:"compare_at_price"`
	TaxRate           float64   `json:"tax_rate"`
	Currency          string    `json:"currency"`
	Weight            float64   `json:"weight"`
	WeightUnit        string    `json:"weight_unit"`
	Length            float64   `json:"length"`
	Width             float64   `json:"width"`
	Height            float64   `json:"height"`
	DimensionUnit     string    `json:"dimension_unit"`
	CountryOfOrigin   string    `json:"country_of_origin"`
	HSCode            string    `json:"hs_code"`
	AllowBackorder    bool      `json:"allow_backorder"`
	LeadTimeDays      int       `json:"lead_time_days"`
	ImageURL          string    `json:"image_url"`
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
		imageURL          sql.NullString
	)

	err := row.Scan(
		&p.ID,
		&p.OwnerID,
		&p.BusinessID,
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
		&imageURL,
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

	if imageURL.Valid {
		p.ImageURL = imageURL.String
	}

	return nil
}

type ProductInput struct {
	CategoryID        int64           `json:"category_id"`
	SupplierID        int64           `json:"supplier_id"`
	Name              string          `json:"name"`
	SKU               string          `json:"sku"`
	Barcode           string          `json:"barcode"`
	Description       string          `json:"description"`
	Unit              string          `json:"unit"`
	CostPrice         float64         `json:"cost_price"`
	SalePrice         float64         `json:"sale_price"`
	StockQuantity     float64         `json:"stock_quantity"`
	StockQuantitySet  bool            `json:"-"`
	ImageURLSet       bool            `json:"-"`
	PresentFields     map[string]bool `json:"-"`
	LowStockThreshold float64         `json:"low_stock_threshold"`
	ReorderQuantity   float64         `json:"reorder_quantity"`
	IsActive          bool            `json:"is_active"`
	TrackInventory    bool            `json:"track_inventory"`
	BrandID           int64           `json:"brand_id"`
	CompareAtPrice    float64         `json:"compare_at_price"`
	TaxRate           float64         `json:"tax_rate"`
	Currency          string          `json:"currency"`
	Weight            float64         `json:"weight"`
	WeightUnit        string          `json:"weight_unit"`
	Length            float64         `json:"length"`
	Width             float64         `json:"width"`
	Height            float64         `json:"height"`
	DimensionUnit     string          `json:"dimension_unit"`
	CountryOfOrigin   string          `json:"country_of_origin"`
	HSCode            string          `json:"hs_code"`
	AllowBackorder    bool            `json:"allow_backorder"`
	LeadTimeDays      int             `json:"lead_time_days"`
	ImageURL          string          `json:"image_url"`
}

// ProductAttribute is a per-business custom key/value field on a product.
type ProductAttribute struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Position int    `json:"position"`
}

// ProductAttributeInput is the create/update shape for a custom attribute.
type ProductAttributeInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
