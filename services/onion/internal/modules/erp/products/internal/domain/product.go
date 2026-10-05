package domain

import (
	"errors"
	"time"
)

var (
	ErrProductNotFound          = errors.New("product not found")
	ErrProductAlreadyExists     = errors.New("product SKU or barcode already exists")
	ErrProductReferenceNotFound = errors.New("product reference not found in business")
)

type Product struct {
	ID                int64
	OwnerID           int64
	BusinessID        int64
	CategoryID        int64
	SupplierID        int64
	Name              string
	SKU               string
	Barcode           string
	Description       string
	Unit              string
	CostPrice         float64
	SalePrice         float64
	StockQuantity     float64
	LowStockThreshold float64
	ReorderQuantity   float64
	IsActive          bool
	TrackInventory    bool
	BrandID           int64
	CompareAtPrice    float64
	TaxRate           float64
	Currency          string
	Weight            float64
	WeightUnit        string
	Length            float64
	Width             float64
	Height            float64
	DimensionUnit     string
	CountryOfOrigin   string
	HSCode            string
	AllowBackorder    bool
	LeadTimeDays      int
	ImageURL          string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

type ProductInput struct {
	CategoryID        int64
	SupplierID        int64
	Name              string
	SKU               string
	Barcode           string
	Description       string
	Unit              string
	CostPrice         float64
	SalePrice         float64
	StockQuantity     float64
	StockQuantitySet  bool
	ImageURLSet       bool
	PresentFields     map[string]bool
	LowStockThreshold float64
	ReorderQuantity   float64
	IsActive          bool
	TrackInventory    bool
	BrandID           int64
	CompareAtPrice    float64
	TaxRate           float64
	Currency          string
	Weight            float64
	WeightUnit        string
	Length            float64
	Width             float64
	Height            float64
	DimensionUnit     string
	CountryOfOrigin   string
	HSCode            string
	AllowBackorder    bool
	LeadTimeDays      int
	ImageURL          string
}

// ProductAttribute is a per-business custom key/value field on a product.
type ProductAttribute struct {
	ID       int64
	Name     string
	Value    string
	Position int
}

// ProductAttributeInput is the create/update shape for a custom attribute.
type ProductAttributeInput struct {
	Name  string
	Value string
}
