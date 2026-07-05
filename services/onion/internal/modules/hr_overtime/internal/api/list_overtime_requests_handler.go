package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_overtime/internal/application"
	"github.com/voxire/onion/internal/modules/hr_overtime/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrovertimev1 "github.com/voxire/voxie/gen/go/hr/overtime/v1"
)

func ListOvertimeRequestsRoutes() []string { return []string{"/hr/overtime/requests"} }

type ListOvertimeRequestsHandler struct{ Service *application.Service }

type ListOvertimeRequestsRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListOvertimeRequestsResponse struct {
	Requests []*hrovertimev1.OvertimeRequestPayload `json:"requests"`
	Total    int64                                  `json:"total"`
	Limit    int32                                  `json:"limit"`
	Offset   int32                                  `json:"offset"`
}

func (h *ListOvertimeRequestsHandler) Serve(ctx context.Context, req *ListOvertimeRequestsRequest) (ListOvertimeRequestsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListOvertimeRequestsResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListOvertimeRequestsResponse{}, onekiterrors.PermissionDenied()
	}

	requests, total, err := h.Service.ListOvertimeRequests(ctx, businessID, domain.ListOvertimeRequestsParams{
		Limit: req.Limit, Offset: req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list overtime requests", logger.Any("business_id", businessID), logger.Error(err))
		return ListOvertimeRequestsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list overtime requests")
	}

	payloads := make([]*hrovertimev1.OvertimeRequestPayload, 0, len(requests))
	for _, r := range requests {
		payloads = append(payloads, overtimeRequestToProto(r))
	}

	return ListOvertimeRequestsResponse{Requests: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
