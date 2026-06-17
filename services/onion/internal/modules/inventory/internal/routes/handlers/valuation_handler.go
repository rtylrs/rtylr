package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpinventoryv1 "github.com/voxire/voxie/gen/go/erp/inventory/v1"
)

func ValuationRoutes() []string {
	return []string{"/erp/inventory/valuation"}
}

type ValuationHandler struct {
	Service *application.Service
}

type ValuationRequest struct{}

type ValuationResponse struct {
	Items      []*erpinventoryv1.ValuationItem `json:"items"`
	TotalValue float64                         `json:"total_value"`
}

func newValuationItemPayload(item domain.ValuationItem) *erpinventoryv1.ValuationItem {
	return &erpinventoryv1.ValuationItem{
		CategoryName:  item.CategoryName,
		WarehouseName: item.WarehouseName,
		TotalItems:    item.TotalItems,
		TotalValue:    item.TotalValue,
		TotalCost:     item.TotalCost,
	}
}

func (h *ValuationHandler) Serve(ctx context.Context, _ *ValuationRequest) (ValuationResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return ValuationResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	ownerID := tenant.ResolveOwnerID(ctx, authUser.BusinessID, authUser.UserID)
	items, err := h.Service.GetValuation(ctx, ownerID)
	if err != nil {
		return ValuationResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to get inventory valuation",
		}
	}

	payloads := make([]*erpinventoryv1.ValuationItem, 0, len(items))
	var totalValue float64

	for _, item := range items {
		payloads = append(payloads, newValuationItemPayload(item))
		totalValue += item.TotalValue
	}

	return ValuationResponse{
		Items:      payloads,
		TotalValue: totalValue,
	}, nil
}
