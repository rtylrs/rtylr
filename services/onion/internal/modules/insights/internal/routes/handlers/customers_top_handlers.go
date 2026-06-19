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

// GET /insights/customers/top?limit=10

func GetTopCustomersRoutes() []string { return []string{"/insights/customers/top"} }

type GetTopCustomersHandler struct{ Service *application.Service }

type GetTopCustomersRequest struct {
	Limit string `json:"limit"`
}

type GetTopCustomersResponse struct {
	Customers []application.TopCustomerPayload `json:"customers"`
}

func (h *GetTopCustomersHandler) Validate(_ context.Context, req *GetTopCustomersRequest) *errorModels.ServerError {
	req.Limit = strings.TrimSpace(req.Limit)
	return nil
}

func (h *GetTopCustomersHandler) Serve(ctx context.Context, req *GetTopCustomersRequest) (GetTopCustomersResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetTopCustomersResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetTopCustomersResponse{}, errorModels.PermissionDeniedServerError()
	}
	var limit int32 = 10
	if req.Limit != "" {
		parsed, err := strconv.ParseInt(req.Limit, 10, 32)
		if err != nil || parsed <= 0 {
			return GetTopCustomersResponse{}, &errorModels.ServerError{
				HTTPCode: http.StatusBadRequest,
				Code:     http.StatusBadRequest,
				Message:  "Invalid limit",
			}
		}
		limit = int32(parsed)
	}

	rows, err := h.Service.GetTopCustomers(ctx, businessID, limit)
	if err != nil {
		return GetTopCustomersResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load top customers",
		}
	}
	return GetTopCustomersResponse{Customers: rows}, nil
}
