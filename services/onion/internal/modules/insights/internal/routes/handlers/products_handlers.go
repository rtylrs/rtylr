package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/products

func GetProductKPIsRoutes() []string { return []string{"/insights/products"} }

type GetProductKPIsHandler struct{ Service *application.Service }

type GetProductKPIsRequest struct{}

type GetProductKPIsResponse = application.ProductKPIPayload

func (h *GetProductKPIsHandler) Serve(ctx context.Context, _ *GetProductKPIsRequest) (GetProductKPIsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetProductKPIsResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetProductKPIsResponse{}, errorModels.PermissionDeniedServerError()
	}
	payload, err := h.Service.GetProductKPIs(ctx, businessID)
	if err != nil {
		return GetProductKPIsResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load product KPIs",
		}
	}
	return payload, nil
}
