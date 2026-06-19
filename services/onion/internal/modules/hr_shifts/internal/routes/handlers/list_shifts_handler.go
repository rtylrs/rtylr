package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/tenant"
	hrshiftv1 "github.com/voxire/voxie/gen/go/hr/shift/v1"
)

func ListShiftsRoutes() []string { return []string{"/hr/shifts"} }

type ListShiftsHandler struct{ Service *application.Service }

type ListShiftsRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListShiftsResponse struct {
	Shifts []*hrshiftv1.ShiftPayload `json:"shifts"`
	Total  int64                     `json:"total"`
	Limit  int32                     `json:"limit"`
	Offset int32                     `json:"offset"`
}

func (h *ListShiftsHandler) Serve(ctx context.Context, req *ListShiftsRequest) (ListShiftsResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListShiftsResponse{}, errorModels.PermissionDeniedServerError()
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListShiftsResponse{}, errorModels.PermissionDeniedServerError()
	}

	shifts, total, err := h.Service.ListShifts(ctx, businessID, domain.ListShiftsParams{
		Limit: req.Limit, Offset: req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list shifts", logger.Any("business_id", businessID), logger.Error(err))
		return ListShiftsResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to list shifts"}
	}

	payloads := make([]*hrshiftv1.ShiftPayload, 0, len(shifts))
	for _, s := range shifts {
		payloads = append(payloads, shiftToProto(s))
	}

	return ListShiftsResponse{Shifts: payloads, Total: total, Limit: req.Limit, Offset: req.Offset}, nil
}
