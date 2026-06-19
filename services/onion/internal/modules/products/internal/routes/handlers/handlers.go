package handlers

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/products/internal/application"
	"github.com/voxire/onion/internal/modules/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─── List Products ────────────────────────────────────────────────────────────

func ListProductsRoutes() []string {
	return []string{"/erp/products"}
}

type ListProductsHandler struct {
	Service *application.Service
}

type ListProductsRequest struct {
	Search     string `json:"search"`
	CategoryID int    `json:"category_id"`
	Page       int64  `json:"page"`
	PageSize   int64  `json:"page_size"`
}

type ListProductsResponse struct {
	Products []domain.Product `json:"products"`
	HasMore  bool             `json:"has_more"`
	Page     int64            `json:"page"`
	PageSize int64            `json:"page_size"`
}

func (h *ListProductsHandler) Serve(ctx context.Context, req *ListProductsRequest) (ListProductsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListProductsResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return ListProductsResponse{}, errorModels.PermissionDeniedServerError()
	}
	products, hasMore, err := h.Service.ListProducts(ctx, ownerID, req.Search, int64(req.CategoryID), req.Page, req.PageSize)
	if err != nil {
		return ListProductsResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list products",
		}
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 50
	}
	return ListProductsResponse{Products: products, HasMore: hasMore, Page: page, PageSize: pageSize}, nil
}

// ─── Create Product ───────────────────────────────────────────────────────────

func CreateProductRoutes() []string {
	return []string{"/erp/products"}
}

type CreateProductHandler struct {
	Service *application.Service
}

type CreateProductRequest struct {
	CategoryID        int64   `json:"category_id"`
	SupplierID        int64   `json:"supplier_id"`
	Name              string  `json:"name"`
	SKU               string  `json:"sku"`
	Barcode           string  `json:"barcode"`
	Description       string  `json:"description"`
	Unit              string  `json:"unit"`
	CostPrice         float64 `json:"cost_price"`
	SalePrice         float64 `json:"sale_price"`
	StockQuantity     float64 `json:"stock_quantity"`
	BrandID           int64   `json:"brand_id"`
	CompareAtPrice    float64 `json:"compare_at_price"`
	TaxRate           float64 `json:"tax_rate"`
	Currency          string  `json:"currency"`
	Weight            float64 `json:"weight"`
	WeightUnit        string  `json:"weight_unit"`
	Length            float64 `json:"length"`
	Width             float64 `json:"width"`
	Height            float64 `json:"height"`
	DimensionUnit     string  `json:"dimension_unit"`
	CountryOfOrigin   string  `json:"country_of_origin"`
	HSCode            string  `json:"hs_code"`
	AllowBackorder    bool    `json:"allow_backorder"`
	LeadTimeDays      int     `json:"lead_time_days"`
	ImageURL          string  `json:"image_url"`
	LowStockThreshold float64 `json:"low_stock_threshold"`
	ReorderQuantity   float64 `json:"reorder_quantity"`
	IsActive          *bool   `json:"is_active"`
	TrackInventory    *bool   `json:"track_inventory"`

	Attributes []domain.ProductAttributeInput `json:"attributes"`
}

type CreateProductResponse struct {
	ProductID int64  `json:"product_id"`
	Message   string `json:"message"`
}

func (h *CreateProductHandler) Validate(_ context.Context, req *CreateProductRequest) *errorModels.ServerError {
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Product name is required",
		}
	}

	if req.CostPrice < 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Cost price must not be negative",
		}
	}

	if req.SalePrice < 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Sale price must not be negative",
		}
	}

	if req.StockQuantity < 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Stock quantity must not be negative",
		}
	}

	return nil
}

func (h *CreateProductHandler) Serve(ctx context.Context, req *CreateProductRequest) (CreateProductResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateProductResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	input := domain.ProductInput{
		CategoryID:        req.CategoryID,
		SupplierID:        req.SupplierID,
		Name:              req.Name,
		SKU:               req.SKU,
		Barcode:           req.Barcode,
		Description:       req.Description,
		Unit:              req.Unit,
		CostPrice:         req.CostPrice,
		SalePrice:         req.SalePrice,
		StockQuantity:     req.StockQuantity,
		StockQuantitySet:  true,
		BrandID:           req.BrandID,
		CompareAtPrice:    req.CompareAtPrice,
		TaxRate:           req.TaxRate,
		Currency:          req.Currency,
		Weight:            req.Weight,
		WeightUnit:        req.WeightUnit,
		Length:            req.Length,
		Width:             req.Width,
		Height:            req.Height,
		DimensionUnit:     req.DimensionUnit,
		CountryOfOrigin:   req.CountryOfOrigin,
		HSCode:            req.HSCode,
		AllowBackorder:    req.AllowBackorder,
		LeadTimeDays:      req.LeadTimeDays,
		ImageURL:          req.ImageURL,
		LowStockThreshold: req.LowStockThreshold,
		ReorderQuantity:   req.ReorderQuantity,
		IsActive:          productBoolDefault(req.IsActive, true),
		TrackInventory:    productBoolDefault(req.TrackInventory, true),
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return CreateProductResponse{}, errorModels.PermissionDeniedServerError()
	}
	id, err := h.Service.CreateProduct(ctx, ownerID, input)
	if err != nil {
		if errors.Is(err, domain.ErrProductAlreadyExists) {
			return CreateProductResponse{}, &errorModels.ServerError{
				Code:     http.StatusConflict,
				HTTPCode: http.StatusConflict,
				Message:  err.Error(),
			}
		}
		return CreateProductResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create product",
		}
	}

	// Best-effort: the product is created; attribute failure shouldn't fail it.
	_ = h.Service.ReplaceProductAttributes(ctx, ownerID, id, req.Attributes)

	return CreateProductResponse{
		ProductID: id,
		Message:   "Product created successfully",
	}, nil
}

// ─── Update Product ───────────────────────────────────────────────────────────

func UpdateProductRoutes() []string {
	return []string{"/erp/products"}
}

type UpdateProductHandler struct {
	Service *application.Service
}

type UpdateProductRequest struct {
	ID                int64    `json:"id"`
	CategoryID        int64    `json:"category_id"`
	SupplierID        int64    `json:"supplier_id"`
	Name              string   `json:"name"`
	SKU               string   `json:"sku"`
	Barcode           string   `json:"barcode"`
	Description       string   `json:"description"`
	Unit              string   `json:"unit"`
	CostPrice         float64  `json:"cost_price"`
	SalePrice         float64  `json:"sale_price"`
	StockQuantity     *float64 `json:"stock_quantity"`
	BrandID           int64    `json:"brand_id"`
	CompareAtPrice    float64  `json:"compare_at_price"`
	TaxRate           float64  `json:"tax_rate"`
	Currency          string   `json:"currency"`
	Weight            float64  `json:"weight"`
	WeightUnit        string   `json:"weight_unit"`
	Length            float64  `json:"length"`
	Width             float64  `json:"width"`
	Height            float64  `json:"height"`
	DimensionUnit     string   `json:"dimension_unit"`
	CountryOfOrigin   string   `json:"country_of_origin"`
	HSCode            string   `json:"hs_code"`
	AllowBackorder    bool     `json:"allow_backorder"`
	LeadTimeDays      int      `json:"lead_time_days"`
	ImageURL          string   `json:"image_url"`
	LowStockThreshold float64  `json:"low_stock_threshold"`
	ReorderQuantity   float64  `json:"reorder_quantity"`
	IsActive          *bool    `json:"is_active"`
	TrackInventory    *bool    `json:"track_inventory"`

	Attributes []domain.ProductAttributeInput `json:"attributes"`
}

type UpdateProductResponse struct {
	Message string `json:"message"`
}

func (h *UpdateProductHandler) Validate(_ context.Context, req *UpdateProductRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Product ID is required",
		}
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Product name is required",
		}
	}

	if req.CostPrice < 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Cost price must not be negative",
		}
	}

	if req.SalePrice < 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Sale price must not be negative",
		}
	}

	if req.StockQuantity != nil && *req.StockQuantity < 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Stock quantity must not be negative",
		}
	}

	return nil
}

func (h *UpdateProductHandler) Serve(ctx context.Context, req *UpdateProductRequest) (UpdateProductResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateProductResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return UpdateProductResponse{}, errorModels.PermissionDeniedServerError()
	}
	isActive, trackInventory, flagErr := productFlagsForUpdate(ctx, h.Service, ownerID, req.ID, req.IsActive, req.TrackInventory)
	if flagErr != nil {
		return UpdateProductResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update product",
		}
	}

	input := domain.ProductInput{
		CategoryID:        req.CategoryID,
		SupplierID:        req.SupplierID,
		Name:              req.Name,
		SKU:               req.SKU,
		Barcode:           req.Barcode,
		Description:       req.Description,
		Unit:              req.Unit,
		CostPrice:         req.CostPrice,
		SalePrice:         req.SalePrice,
		BrandID:           req.BrandID,
		CompareAtPrice:    req.CompareAtPrice,
		TaxRate:           req.TaxRate,
		Currency:          req.Currency,
		Weight:            req.Weight,
		WeightUnit:        req.WeightUnit,
		Length:            req.Length,
		Width:             req.Width,
		Height:            req.Height,
		DimensionUnit:     req.DimensionUnit,
		CountryOfOrigin:   req.CountryOfOrigin,
		HSCode:            req.HSCode,
		AllowBackorder:    req.AllowBackorder,
		LeadTimeDays:      req.LeadTimeDays,
		ImageURL:          req.ImageURL,
		LowStockThreshold: req.LowStockThreshold,
		ReorderQuantity:   req.ReorderQuantity,
		IsActive:          isActive,
		TrackInventory:    trackInventory,
	}
	if req.StockQuantity != nil {
		input.StockQuantity = *req.StockQuantity
		input.StockQuantitySet = true
	}

	if err := h.Service.UpdateProduct(ctx, ownerID, req.ID, input); err != nil {
		return UpdateProductResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update product",
		}
	}

	// Replace the product's custom attributes with the submitted set.
	_ = h.Service.ReplaceProductAttributes(ctx, ownerID, req.ID, req.Attributes)

	return UpdateProductResponse{Message: "Product updated successfully"}, nil
}

// ─── Delete Product ───────────────────────────────────────────────────────────

func DeleteProductRoutes() []string {
	return []string{"/erp/products"}
}

type DeleteProductHandler struct {
	Service *application.Service
}

type DeleteProductRequest struct {
	ID int64 `json:"id"`
}

type DeleteProductResponse struct {
	Message string `json:"message"`
}

func (h *DeleteProductHandler) Validate(_ context.Context, req *DeleteProductRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Product ID is required",
		}
	}

	return nil
}

func (h *DeleteProductHandler) Serve(ctx context.Context, req *DeleteProductRequest) (DeleteProductResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteProductResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	ownerID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return DeleteProductResponse{}, errorModels.PermissionDeniedServerError()
	}
	if err := h.Service.DeleteProduct(ctx, ownerID, req.ID); err != nil {
		return DeleteProductResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to delete product",
		}
	}

	return DeleteProductResponse{Message: "Product deleted successfully"}, nil
}
