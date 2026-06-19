package handlers

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/products/best-sellers?limit=10

func GetBestSellersRoutes() []string { return []string{"/insights/products/best-sellers"} }

type GetBestSellersHandler struct{ Service *application.Service }

type GetBestSellersRequest struct {
	Limit string `json:"limit"`
}

type GetBestSellersResponse struct {
	Products []application.BestSellerPayload `json:"products"`
}

func (h *GetBestSellersHandler) Validate(_ context.Context, req *GetBestSellersRequest) *errorModels.ServerError {
	req.Limit = strings.TrimSpace(req.Limit)
	return nil
}

func (h *GetBestSellersHandler) Serve(ctx context.Context, req *GetBestSellersRequest) (GetBestSellersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetBestSellersResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetBestSellersResponse{}, errorModels.PermissionDeniedServerError()
	}
	var limit int32 = 10
	if req.Limit != "" {
		parsed, err := strconv.ParseInt(req.Limit, 10, 32)
		if err != nil || parsed <= 0 {
			return GetBestSellersResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Invalid limit",
			}
		}
		limit = int32(parsed)
	}
	rows, err := h.Service.GetBestSellers(ctx, businessID, limit)
	if err != nil {
		return GetBestSellersResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load best sellers",
		}
	}
	return GetBestSellersResponse{Products: rows}, nil
}
