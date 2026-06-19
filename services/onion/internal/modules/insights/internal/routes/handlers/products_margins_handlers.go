package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/products/margins

func GetMarginDistributionRoutes() []string { return []string{"/insights/products/margins"} }

type GetMarginDistributionHandler struct{ Service *application.Service }

type GetMarginDistributionRequest struct{}

type GetMarginDistributionResponse struct {
	Buckets []application.MarginBucketPayload `json:"buckets"`
}

func (h *GetMarginDistributionHandler) Serve(ctx context.Context, _ *GetMarginDistributionRequest) (GetMarginDistributionResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetMarginDistributionResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetMarginDistributionResponse{}, errorModels.PermissionDeniedServerError()
	}
	buckets, err := h.Service.GetMarginDistribution(ctx, businessID)
	if err != nil {
		return GetMarginDistributionResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load margin distribution",
		}
	}
	return GetMarginDistributionResponse{Buckets: buckets}, nil
}
