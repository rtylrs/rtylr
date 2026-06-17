package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/variants/internal/application"
	"github.com/voxire/onion/internal/modules/variants/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// --- List Variants -----------------------------------------------------------

func ListVariantsRoutes() []string {
	return []string{"/erp/variants"}
}

type ListVariantsHandler struct {
	Service *application.Service
}

type ListVariantsRequest struct {
	ProductID int64 `json:"product_id"`
}

type VariantPayload struct {
	ID            int64   `json:"id"`
	OwnerID       int64   `json:"owner_id"`
	ProductID     int64   `json:"product_id"`
	Name          string  `json:"name"`
	SKU           string  `json:"sku"`
	Barcode       string  `json:"barcode"`
	CostPrice     float64 `json:"cost_price"`
	SalePrice     float64 `json:"sale_price"`
	StockQuantity float64 `json:"stock_quantity"`
	Attributes    string  `json:"attributes"`
	IsActive      bool    `json:"is_active"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type ListVariantsResponse struct {
	Variants []VariantPayload `json:"variants"`
}

func newVariantPayload(v domain.Variant) VariantPayload {
	return VariantPayload{
		ID:            v.ID,
		OwnerID:       v.OwnerID,
		ProductID:     v.ProductID,
		Name:          v.Name,
		SKU:           v.SKU,
		Barcode:       v.Barcode,
		CostPrice:     v.CostPrice,
		SalePrice:     v.SalePrice,
		StockQuantity: v.StockQuantity,
		Attributes:    v.Attributes,
		IsActive:      v.IsActive,
		CreatedAt:     v.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:     v.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ListVariantsHandler) Serve(ctx context.Context, req *ListVariantsRequest) (ListVariantsResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ListVariantsResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	variants, err := h.Service.ListVariants(ctx, ownerID, req.ProductID)
	if err != nil {
		return ListVariantsResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to list variants",
		}
	}

	payloads := make([]VariantPayload, 0, len(variants))
	for _, v := range variants {
		payloads = append(payloads, newVariantPayload(v))
	}

	return ListVariantsResponse{Variants: payloads}, nil
}

// --- Create Variant ----------------------------------------------------------

func CreateVariantRoutes() []string {
	return []string{"/erp/variants"}
}

type CreateVariantHandler struct {
	Service *application.Service
}

type CreateVariantRequest struct {
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

type CreateVariantResponse struct {
	VariantID int64  `json:"variant_id"`
	Message   string `json:"message"`
}

func (h *CreateVariantHandler) Validate(_ context.Context, req *CreateVariantRequest) *errorModels.ServerError {
	if req.ProductID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Product ID is required",
		}
	}

	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Variant name is required",
		}
	}

	return nil
}

func (h *CreateVariantHandler) Serve(ctx context.Context, req *CreateVariantRequest) (CreateVariantResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return CreateVariantResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	input := domain.VariantInput{
		ProductID:     req.ProductID,
		Name:          req.Name,
		SKU:           req.SKU,
		Barcode:       req.Barcode,
		CostPrice:     req.CostPrice,
		SalePrice:     req.SalePrice,
		StockQuantity: req.StockQuantity,
		Attributes:    req.Attributes,
		IsActive:      req.IsActive,
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	id, err := h.Service.CreateVariant(ctx, ownerID, input)
	if err != nil {
		return CreateVariantResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to create variant",
		}
	}

	return CreateVariantResponse{VariantID: id, Message: "Variant created successfully"}, nil
}

// --- Update Variant ----------------------------------------------------------

func UpdateVariantRoutes() []string {
	return []string{"/erp/variants"}
}

type UpdateVariantHandler struct {
	Service *application.Service
}

type UpdateVariantRequest struct {
	ID            int64   `json:"id"`
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

type UpdateVariantResponse struct {
	Message string `json:"message"`
}

func (h *UpdateVariantHandler) Validate(_ context.Context, req *UpdateVariantRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Variant ID is required",
		}
	}

	if strings.TrimSpace(req.Name) == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Variant name is required",
		}
	}

	return nil
}

func (h *UpdateVariantHandler) Serve(ctx context.Context, req *UpdateVariantRequest) (UpdateVariantResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return UpdateVariantResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	input := domain.VariantInput{
		ProductID:     req.ProductID,
		Name:          req.Name,
		SKU:           req.SKU,
		Barcode:       req.Barcode,
		CostPrice:     req.CostPrice,
		SalePrice:     req.SalePrice,
		StockQuantity: req.StockQuantity,
		Attributes:    req.Attributes,
		IsActive:      req.IsActive,
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	if err := h.Service.UpdateVariant(ctx, ownerID, req.ID, input); err != nil {
		return UpdateVariantResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to update variant",
		}
	}

	return UpdateVariantResponse{Message: "Variant updated successfully"}, nil
}

// --- Delete Variant ----------------------------------------------------------

func DeleteVariantRoutes() []string {
	return []string{"/erp/variants"}
}

type DeleteVariantHandler struct {
	Service *application.Service
}

type DeleteVariantRequest struct {
	ID int64 `json:"id"`
}

type DeleteVariantResponse struct {
	Message string `json:"message"`
}

func (h *DeleteVariantHandler) Validate(_ context.Context, req *DeleteVariantRequest) *errorModels.ServerError {
	if req.ID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Variant ID is required",
		}
	}

	return nil
}

func (h *DeleteVariantHandler) Serve(ctx context.Context, req *DeleteVariantRequest) (DeleteVariantResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return DeleteVariantResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	if err := h.Service.DeleteVariant(ctx, ownerID, req.ID); err != nil {
		return DeleteVariantResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to delete variant",
		}
	}

	return DeleteVariantResponse{Message: "Variant deleted successfully"}, nil
}
