package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func AdjustInventoryRoutes() []string {
	return []string{"/erp/inventory/adjust"}
}

type AdjustInventoryHandler struct {
	Service *application.Service
}

type AdjustInventoryRequest struct {
	ProductID int64   `json:"product_id"`
	Change    float64 `json:"change"`
	Type      string  `json:"type"`
	Notes     string  `json:"notes"`
}

type AdjustInventoryResponse struct {
	Message string `json:"message"`
}

var validInventoryTypes = map[string]struct{}{
	string(domain.InventoryTypePurchase):   {},
	string(domain.InventoryTypeSale):       {},
	string(domain.InventoryTypeAdjustment): {},
	string(domain.InventoryTypeReturn):     {},
	string(domain.InventoryTypeWaste):      {},
}

func (h *AdjustInventoryHandler) Validate(_ context.Context, req *AdjustInventoryRequest) *errorModels.ServerError {
	if req.ProductID == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Product ID is required",
		}
	}

	if req.Change == 0 {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Change quantity cannot be zero",
		}
	}

	if req.Type == "" {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Transaction type is required",
		}
	}

	if _, valid := validInventoryTypes[req.Type]; !valid {
		return &errorModels.ServerError{
			Code:     http.StatusBadRequest,
			HTTPCode: http.StatusBadRequest,
			Message:  "Invalid type. Must be one of: purchase, sale, adjustment, return, waste",
		}
	}

	return nil
}

func (h *AdjustInventoryHandler) Serve(ctx context.Context, req *AdjustInventoryRequest) (AdjustInventoryResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return AdjustInventoryResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	input := domain.InventoryAdjustInput{
		ProductID: req.ProductID,
		Change:    req.Change,
		Type:      domain.InventoryTransactionType(req.Type),
		Notes:     req.Notes,
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	if err := h.Service.AdjustInventory(ctx, ownerID, input); err != nil {
		return AdjustInventoryResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to adjust inventory",
		}
	}

	return AdjustInventoryResponse{Message: "Inventory adjusted successfully"}, nil
}
