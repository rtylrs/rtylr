package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/customers/ltv

func GetLTVRoutes() []string { return []string{"/insights/customers/ltv"} }

type GetLTVHandler struct{ Service *application.Service }

type GetLTVRequest struct{}

type GetLTVResponse struct {
	Buckets []application.LTVBucketPayload `json:"buckets"`
}

func (h *GetLTVHandler) Serve(ctx context.Context, _ *GetLTVRequest) (GetLTVResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetLTVResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetLTVResponse{}, errorModels.PermissionDeniedServerError()
	}
	buckets, err := h.Service.GetLTVDistribution(ctx, businessID)
	if err != nil {
		return GetLTVResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load LTV distribution",
		}
	}
	return GetLTVResponse{Buckets: buckets}, nil
}
