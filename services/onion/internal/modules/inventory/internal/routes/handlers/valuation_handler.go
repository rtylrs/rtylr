package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	"github.com/voxire/onion/internal/modules/inventory/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
)

func ValuationRoutes() []string {
	return []string{"/erp/inventory/valuation"}
}

type ValuationHandler struct {
	Service *application.Service
}

type ValuationRequest struct{}

type ValuationItemPayload struct {
	CategoryName string  `json:"category_name"`
	WarehouseName string `json:"warehouse_name"`
	TotalItems   int64   `json:"total_items"`
	TotalValue   float64 `json:"total_value"`
	TotalCost    float64 `json:"total_cost"`
}

type ValuationResponse struct {
	Items      []ValuationItemPayload `json:"items"`
	TotalValue float64                `json:"total_value"`
}

func newValuationItemPayload(item domain.ValuationItem) ValuationItemPayload {
	return ValuationItemPayload{
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

	items, err := h.Service.GetValuation(ctx, authUser.UserID)
	if err != nil {
		return ValuationResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to get inventory valuation",
		}
	}

	payloads := make([]ValuationItemPayload, 0, len(items))
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
