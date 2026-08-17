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
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

func ListLeaveTypesRoutes() []string { return []string{"/hr/leave/types"} }

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

	limit := req.Limit
	if limit <= 0 {
		limit = defaultLeaveTypesLimit
	}
	if limit > maxLeaveTypesLimit {
		limit = maxLeaveTypesLimit
	}

	types, total, err := h.Service.ListLeaveTypes(ctx, device.User.UserID, domain.ListLeaveTypesParams{
		Limit: limit, Offset: req.Offset,
	})
	if err != nil {
		if errors.Is(err, domain.ErrLeaveNotConfigured) {
			return ListLeaveTypesResponse{Types: []*hrleavev1.LeaveTypePayload{}, Total: 0, Limit: limit, Offset: req.Offset}, nil
		}
		logger.Ctx(ctx).Error("failed to list leave types", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListLeaveTypesResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list leave types")
	}

	payloads := make([]*hrleavev1.LeaveTypePayload, 0, len(types))
	for _, t := range types {
		payloads = append(payloads, leaveTypeToProto(t))
	}

	return ListLeaveTypesResponse{Types: payloads, Total: total, Limit: limit, Offset: req.Offset}, nil
}
