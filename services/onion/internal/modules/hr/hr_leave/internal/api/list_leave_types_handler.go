package api

import (
	"context"
	"errors"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

type ListLeaveTypesHandler struct{ Service *application.Service }

type ListLeaveTypesRequest struct {
	Limit  int32 `json:"limit"`
	Offset int32 `json:"offset"`
}

type ListLeaveTypesResponse struct {
	Types  []*hrleavev1.LeaveTypePayload `json:"types"`
	Total  int64                         `json:"total"`
	Limit  int32                         `json:"limit"`
	Offset int32                         `json:"offset"`
}

// defaultLeaveTypesLimit and maxLeaveTypesLimit clamp the caller-supplied
// limit before it reaches the service, matching the same 20/100 clamp the
// service itself applies as a second line of defense. Clamping here too keeps
// the Limit/Offset the response echoes back honest about what was actually
// used, instead of echoing an unbounded value the query never honored.
const (
	defaultLeaveTypesLimit = 20
	maxLeaveTypesLimit     = 100
)

func (h *ListLeaveTypesHandler) Serve(ctx context.Context, req *ListLeaveTypesRequest) (ListLeaveTypesResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListLeaveTypesResponse{}, onekiterrors.PermissionDenied()
	}
	if req == nil {
		return ListLeaveTypesResponse{}, onekiterrors.New(http.StatusBadRequest, "request is required")
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListLeaveTypesResponse{}, onekiterrors.PermissionDenied()
	}

	offset := req.Offset
	if offset < 0 {
		offset = 0
	}
	limit := req.Limit
	if limit <= 0 {
		limit = defaultLeaveTypesLimit
	}
	if limit > maxLeaveTypesLimit {
		limit = maxLeaveTypesLimit
	}

	types, total, err := h.Service.ListLeaveTypes(ctx, businessID, domain.ListLeaveTypesParams{
		Limit: limit, Offset: offset,
	})
	if err != nil {
		if errors.Is(err, domain.ErrLeaveNotConfigured) {
			return ListLeaveTypesResponse{Types: []*hrleavev1.LeaveTypePayload{}, Total: 0, Limit: limit, Offset: offset}, nil
		}
		logger.Ctx(ctx).Error("failed to list leave types", logger.Any("business_id", businessID), logger.Error(err))
		return ListLeaveTypesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list leave types")
	}

	payloads := make([]*hrleavev1.LeaveTypePayload, 0, len(types))
	for _, t := range types {
		payloads = append(payloads, leaveTypeToProto(t))
	}

	return ListLeaveTypesResponse{Types: payloads, Total: total, Limit: limit, Offset: offset}, nil
}
