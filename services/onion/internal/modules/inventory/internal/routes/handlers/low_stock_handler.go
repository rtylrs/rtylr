package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/inventory/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
	erpinventoryv1 "github.com/voxire/voxie/gen/go/erp/inventory/v1"
)

func LowStockRoutes() []string {
	return []string{"/erp/inventory/low-stock"}
}

type LowStockHandler struct {
	Service *application.Service
}

type LowStockRequest struct{}

type LowStockResponse struct {
	Items []*erpinventoryv1.InventoryReportItem `json:"items"`
	Count int                                   `json:"count"`
}

func (h *LowStockHandler) Serve(ctx context.Context, _ *LowStockRequest) (LowStockResponse, *errorModels.ServerError) {
	authUser, ok := authPkg.GetAuthUser(ctx)
	if !ok {
		return LowStockResponse{}, &errorModels.ServerError{
			Code:     http.StatusUnauthorized,
			HTTPCode: http.StatusUnauthorized,
			Message:  "unauthorized",
		}
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, authUser.BusinessID, authUser.UserID)
	if bizErr != nil {
		return LowStockResponse{}, errorModels.PermissionDeniedServerError()
	}

	items, err := h.Service.GetLowStockProducts(ctx, businessID)
	if err != nil {
		return LowStockResponse{}, &errorModels.ServerError{
			Code:     http.StatusInternalServerError,
			HTTPCode: http.StatusInternalServerError,
			Message:  "Failed to get low stock products",
		}
	}

	payloads := make([]*erpinventoryv1.InventoryReportItem, 0, len(items))
	for _, item := range items {
		payloads = append(payloads, newInventoryReportItemPayload(item))
	}

	return LowStockResponse{
		Items: payloads,
		Count: len(payloads),
	}, nil
}
