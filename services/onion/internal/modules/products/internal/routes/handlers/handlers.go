package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/products/internal/application"
	"github.com/voxire/onion/internal/modules/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
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
	CategoryID int    `json:"categoryId"`
	Limit      int    `json:"limit"`
	Offset     int    `json:"offset"`
}

type ProductPayload struct {
	ID                int64   `json:"id"`
	OwnerID           int64   `json:"ownerId"`
	CategoryID        int64   `json:"categoryId"`
	SupplierID        int64   `json:"supplierId"`
	Name              string  `json:"name"`
	SKU               string  `json:"sku"`
	Barcode           string  `json:"barcode"`
	Description       string  `json:"description"`
	Unit              string  `json:"unit"`
	CostPrice         float64 `json:"costPrice"`
	SalePrice         float64 `json:"salePrice"`
	StockQuantity     float64 `json:"stockQuantity"`
	LowStockThreshold float64 `json:"lowStockThreshold"`
	ReorderQuantity   float64 `json:"reorderQuantity"`
	IsActive          bool    `json:"isActive"`
	TrackInventory    bool    `json:"trackInventory"`
	CreatedAt         string  `json:"createdAt"`
	UpdatedAt         string  `json:"updatedAt"`
}

type ListProductsResponse struct {
	Products []ProductPayload `json:"products"`
}

func newProductPayload(p domain.Product) ProductPayload {
	return ProductPayload{
		ID:                p.ID,
		OwnerID:           p.OwnerID,
		CategoryID:        p.CategoryID,
		SupplierID:        p.SupplierID,
		Name:              p.Name,
		SKU:               p.SKU,
		Barcode:           p.Barcode,
		Description:       p.Description,
		Unit:              p.Unit,
		CostPrice:         p.CostPrice,
		SalePrice:         p.SalePrice,
		StockQuantity:     p.StockQuantity,
		LowStockThreshold: p.LowStockThreshold,
		ReorderQuantity:   p.ReorderQuantity,
		IsActive:          p.IsActive,
		TrackInventory:    p.TrackInventory,
		CreatedAt:         p.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:         p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
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

	products, err := h.Service.ListProducts(ctx, authUser.UserID, req.Search, req.CategoryID, req.Limit, req.Offset)
	if err != nil {
		return ListProductsResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list products",
		}
	}

	payloads := make([]ProductPayload, 0, len(products))
	for _, p := range products {
		payloads = append(payloads, newProductPayload(p))
	}

	return ListProductsResponse{Products: payloads}, nil
}

// ─── Create Product ───────────────────────────────────────────────────────────

func CreateProductRoutes() []string {
	return []string{"/erp/products"}
}

type CreateProductHandler struct {
	Service *application.Service
}

type CreateProductRequest struct {
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

type CreateProductResponse struct {
	ProductID int64  `json:"productId"`
	Message   string `json:"message"`
}

func (h *CreateProductHandler) Validate(_ context.Context, req *CreateProductRequest) *errorModels.ServerError {
	if req.Name == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Product name is required",
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
		LowStockThreshold: req.LowStockThreshold,
		ReorderQuantity:   req.ReorderQuantity,
		IsActive:          req.IsActive,
		TrackInventory:    req.TrackInventory,
	}

	id, err := h.Service.CreateProduct(ctx, authUser.UserID, input)
	if err != nil {
		return CreateProductResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create product",
		}
	}

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
	ID                int64   `json:"id"`
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

	if req.Name == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Product name is required",
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
		LowStockThreshold: req.LowStockThreshold,
		ReorderQuantity:   req.ReorderQuantity,
		IsActive:          req.IsActive,
		TrackInventory:    req.TrackInventory,
	}

	if err := h.Service.UpdateProduct(ctx, authUser.UserID, req.ID, input); err != nil {
		return UpdateProductResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update product",
		}
	}

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

	if err := h.Service.DeleteProduct(ctx, authUser.UserID, req.ID); err != nil {
		return DeleteProductResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to delete product",
		}
	}

	return DeleteProductResponse{Message: "Product deleted successfully"}, nil
}
