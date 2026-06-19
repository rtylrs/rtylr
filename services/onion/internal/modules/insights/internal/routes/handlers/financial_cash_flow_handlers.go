package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/voxire/onion/internal/modules/insights/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/tenant"
)

// GET /insights/financial/cash-flow?from=&to=

func GetCashFlowRoutes() []string { return []string{"/insights/financial/cash-flow"} }

type GetCashFlowHandler struct{ Service *application.Service }

type GetCashFlowRequest struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GetCashFlowResponse struct {
	Series []application.CashFlowPointPayload `json:"series"`
}

func (h *GetCashFlowHandler) Validate(_ context.Context, req *GetCashFlowRequest) *errorModels.ServerError {
	req.From = strings.TrimSpace(req.From)
	req.To = strings.TrimSpace(req.To)
	return nil
}

func (h *GetCashFlowHandler) Serve(ctx context.Context, req *GetCashFlowRequest) (GetCashFlowResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.User == nil {
		return GetCashFlowResponse{}, errorModels.PermissionDeniedServerError()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetCashFlowResponse{}, errorModels.PermissionDeniedServerError()
	}
	from, to, err := application.ResolveDateRange(req.From, req.To)
	if err != nil {
		return GetCashFlowResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusBadRequest,
			Code:     http.StatusBadRequest,
			Message:  "Invalid date range",
		}
	}
	rows, err := h.Service.GetCashFlowSeries(ctx, businessID, from, to)
	if err != nil {
		return GetCashFlowResponse{}, &errorModels.ServerError{
			HTTPCode: http.StatusInternalServerError,
			Code:     http.StatusInternalServerError,
			Message:  "Failed to load cash-flow series",
		}
	}
	return GetCashFlowResponse{Series: rows}, nil
}
