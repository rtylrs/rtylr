package infrastructure

import (
	"time"
)

type productInputSnapshot struct {
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

type productSnapshot struct {
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

type productAttributeSnapshot struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Value    string `json:"value"`
	Position int    `json:"position"`
}

type productAttributeInputSnapshot struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
