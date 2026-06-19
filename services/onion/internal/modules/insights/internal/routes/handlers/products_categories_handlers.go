package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/products/categories

func GetCategorySalesRoutes() []string { return []string{"/insights/products/categories"} }

type GetCategorySalesHandler struct{ Service *application.Service }

type GetCategorySalesRequest struct{}

type GetCategorySalesResponse struct {
	Categories []application.CategorySalesPayload `json:"categories"`
}

func (h *GetCategorySalesHandler) Serve(ctx context.Context, _ *GetCategorySalesRequest) (GetCategorySalesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCategorySalesResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetCategorySalesResponse{}, errorModels.PermissionDeniedServerError()
	}
	rows, err := h.Service.GetSalesByCategory(ctx, businessID)
	if err != nil {
		return GetCategorySalesResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load sales by category",
		}
	}
	return GetCategorySalesResponse{Categories: rows}, nil
}
