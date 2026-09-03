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

type GetLeaveBalancesResponse struct {
	Balances []*hrleavev1.LeaveBalancePayload `json:"balances"`
}

func (h *GetLeaveBalancesHandler) Serve(ctx context.Context, req *GetLeaveBalancesRequest) (GetLeaveBalancesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetLeaveBalancesResponse{}, onekiterrors.PermissionDenied()
	}

	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return GetLeaveBalancesResponse{}, onekiterrors.PermissionDenied()
	}

	balances, err := h.Service.GetLeaveBalances(ctx, businessID, device.User.UserID, req.EmployeeID, req.Year)
	if err != nil {
		if errors.Is(err, domain.ErrLeaveAccessDenied) {
			return GetLeaveBalancesResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to get leave balances", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetLeaveBalancesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get leave balances")
	}

	payloads := make([]*hrleavev1.LeaveBalancePayload, 0, len(balances))
	for _, b := range balances {
		payloads = append(payloads, leaveBalanceToProto(b))
	}

	return GetLeaveBalancesResponse{Balances: payloads}, nil
}
