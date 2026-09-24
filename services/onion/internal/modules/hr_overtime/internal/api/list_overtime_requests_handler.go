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

// Mirrors the default/max clamp applied in application.Service.ListOvertimeRequests
// so an unbounded or absent caller-provided limit never reaches the query layer.
const (
	defaultListOvertimeRequestsLimit int32 = 20
	maxListOvertimeRequestsLimit     int32 = 100
)

type ListOvertimeRequestsHandler struct{ Service *application.Service }

type ListOvertimeRequestsRequest struct {
	EmployeeID *int64  `json:"employee_id,omitempty"`
	Status     *string `json:"status,omitempty"`
	Limit      int32   `json:"limit"`
	Offset     int32   `json:"offset"`
}

type ListOvertimeRequestsResponse struct {
	Requests []*hrovertimev1.OvertimeRequestPayload `json:"requests"`
	Total    int64                                  `json:"total"`
	Limit    int32                                  `json:"limit"`
	Offset   int32                                  `json:"offset"`
}

func (h *ListOvertimeRequestsHandler) Validate(_ context.Context, req *ListOvertimeRequestsRequest) error {
	if req == nil {
		return onekiterrors.New(http.StatusBadRequest, "request is required")
	}
	if req.EmployeeID != nil && *req.EmployeeID <= 0 {
		return onekiterrors.New(http.StatusBadRequest, "employee_id must be greater than 0")
	}
	normalized := domain.NormalizeOvertimeListParams(domain.ListOvertimeRequestsParams{
		EmployeeID: req.EmployeeID,
		Status:     req.Status,
		Limit:      req.Limit,
		Offset:     req.Offset,
	})
	req.EmployeeID = normalized.EmployeeID
	req.Status = normalized.Status
	req.Offset = normalized.Offset
	return nil
}

func (h *ListOvertimeRequestsHandler) Serve(ctx context.Context, req *ListOvertimeRequestsRequest) (ListOvertimeRequestsResponse, error) {
	if err := h.Validate(ctx, req); err != nil {
		return ListOvertimeRequestsResponse{}, err
	}
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListOvertimeRequestsResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListOvertimeRequestsResponse{}, onekiterrors.PermissionDenied()
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultListOvertimeRequestsLimit
	}
	if limit > maxListOvertimeRequestsLimit {
		limit = maxListOvertimeRequestsLimit
	}

	requests, total, err := h.Service.ListOvertimeRequests(ctx, businessID, domain.ListOvertimeRequestsParams{
		EmployeeID: req.EmployeeID,
		Status:     req.Status,
		Limit:      limit,
		Offset:     req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list overtime requests", logger.Any("business_id", businessID), logger.Error(err))
		return ListOvertimeRequestsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list overtime requests")
	}

	payloads := make([]*hrovertimev1.OvertimeRequestPayload, 0, len(requests))
	for _, r := range requests {
		payloads = append(payloads, overtimeRequestToProto(r))
	}

	return ListOvertimeRequestsResponse{Requests: payloads, Total: total, Limit: limit, Offset: req.Offset}, nil
}
