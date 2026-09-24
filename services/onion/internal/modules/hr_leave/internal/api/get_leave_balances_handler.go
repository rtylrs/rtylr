package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

func GetLeaveBalancesRoutes() []string { return []string{"/hr/leave/balances"} }

type GetLeaveBalancesHandler struct{ Service *application.Service }

type GetLeaveBalancesRequest struct {
	EmployeeID int64 `json:"employee_id"`
	Year       int32 `json:"year"`
}

const leaveBalancesPageLimit int32 = 500

type GetLeaveBalancesResponse struct {
	Balances []*hrleavev1.LeaveBalancePayload `json:"balances"`
	Total    int64                            `json:"total"`
	Limit    int32                            `json:"limit"`
	Offset   int32                            `json:"offset"`
}

func (h *GetLeaveBalancesHandler) Serve(ctx context.Context, req *GetLeaveBalancesRequest) (GetLeaveBalancesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetLeaveBalancesResponse{}, onekiterrors.PermissionDenied()
	}
	if req == nil {
		return GetLeaveBalancesResponse{}, onekiterrors.New(http.StatusBadRequest, "request is required")
	}

	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return GetLeaveBalancesResponse{}, onekiterrors.PermissionDenied()
	}

	balances, total, err := h.Service.GetLeaveBalancesWithTotal(ctx, businessID, device.User.UserID, req.EmployeeID, req.Year)
	if err != nil {
		if errors.Is(err, domain.ErrLeaveAccessDenied) {
			return GetLeaveBalancesResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to get leave balances", logger.Any("business_id", businessID), logger.Error(err))
		return GetLeaveBalancesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get leave balances")
	}

	return newLeaveBalancesResponse(balances, total), nil
}

func newLeaveBalancesResponse(balances []domain.LeaveBalanceRecord, total int64) GetLeaveBalancesResponse {
	payloads := make([]*hrleavev1.LeaveBalancePayload, 0, len(balances))
	for _, b := range balances {
		payloads = append(payloads, leaveBalanceToProto(b))
	}
	return GetLeaveBalancesResponse{
		Balances: payloads,
		Total:    total,
		Limit:    leaveBalancesPageLimit,
		Offset:   0,
	}
}
