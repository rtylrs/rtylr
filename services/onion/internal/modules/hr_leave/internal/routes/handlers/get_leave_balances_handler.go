package handlers

import (
	"context"
	"net/http"

	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
)

func GetLeaveBalancesRoutes() []string { return []string{"/hr/leave/balances"} }

type GetLeaveBalancesHandler struct{ Service *application.Service }

type GetLeaveBalancesRequest struct {
	EmployeeID int64 `json:"employee_id"`
	Year       int32 `json:"year"`
}

type GetLeaveBalancesResponse struct {
	Balances []*hrv1.LeaveBalancePayload `json:"balances"`
}

func (h *GetLeaveBalancesHandler) Serve(ctx context.Context, req *GetLeaveBalancesRequest) (GetLeaveBalancesResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetLeaveBalancesResponse{}, errorModels.PermissionDeniedServerError()
	}

	balances, err := h.Service.GetLeaveBalances(ctx, device.User.UserID, req.EmployeeID, req.Year)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get leave balances", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetLeaveBalancesResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get leave balances"}
	}

	payloads := make([]*hrv1.LeaveBalancePayload, 0, len(balances))
	for _, b := range balances {
		payloads = append(payloads, leaveBalanceToProto(b))
	}

	return GetLeaveBalancesResponse{Balances: payloads}, nil
}
