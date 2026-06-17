package handlers

import (
	"context"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/products/internal/application"
	"github.com/voxire/onion/internal/modules/products/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// ─── Update Product by Path ID ────────────────────────────────────────────────
//
// PUT /erp/products/{id} — REST-style variant. Coexists with the existing
// PUT /erp/products (body-ID) route. Path id wins over body id when both are set.
// Reuses the underlying Service.UpdateProduct business logic.

func UpdateProductByIDRoutes() []string {
	return []string{"/erp/products/{id}"}
}

type UpdateProductByIDHandler struct {
	Service *application.Service
}

type UpdateProductByIDRequest struct {
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

type UpdateProductByIDResponse struct {
	Message string `json:"message"`
}

func (h *UpdateProductByIDHandler) Validate(_ context.Context, req *UpdateProductByIDRequest) *errorModels.ServerError {
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

	return nil
}

func (h *UpdateProductByIDHandler) Serve(ctx context.Context, req *UpdateProductByIDRequest) (UpdateProductByIDResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateProductByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	idRaw := chi.URLParamFromCtx(ctx, "id")
	id, err := strconv.ParseInt(idRaw, 10, 64)
	if err != nil || id <= 0 {
		return UpdateProductByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Invalid product id",
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

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	if err := h.Service.UpdateProduct(ctx, ownerID, id, input); err != nil {
		return UpdateProductByIDResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update product",
		}
	}

	return UpdateProductByIDResponse{Message: "Product updated successfully"}, nil
}
