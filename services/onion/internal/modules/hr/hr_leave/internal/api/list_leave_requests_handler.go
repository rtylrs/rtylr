package api

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/application"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrleavev1 "github.com/voxire/voxie/gen/go/hr/leave/v1"
)

type ListLeaveRequestsHandler struct{ Service *application.Service }

type ListLeaveRequestsRequest struct {
	EmployeeID int64  `json:"employee_id"`
	Status     string `json:"status"`
	// FromDate/ToDate bound the overlap window (ISO-8601); nil is open-ended.
	FromDate *string `json:"from_date"`
	ToDate   *string `json:"to_date"`
	Limit    int32   `json:"limit"`
	Offset   int32   `json:"offset"`
}

type ListLeaveRequestsResponse struct {
	Requests []*hrleavev1.LeaveRequestPayload `json:"requests"`
	Total    int64                            `json:"total"`
	Limit    int32                            `json:"limit"`
	Offset   int32                            `json:"offset"`
}

// defaultLeaveRequestsLimit and maxLeaveRequestsLimit clamp the caller-supplied
// limit before it reaches the service, matching the same 20/100 clamp the
// service itself applies as a second line of defense. Clamping here too keeps
// the Limit/Offset the response echoes back honest about what was actually
// used, instead of echoing an unbounded value the query never honored.
const (
	defaultLeaveRequestsLimit = 20
	maxLeaveRequestsLimit     = 100
)

func (h *ListLeaveRequestsHandler) Serve(ctx context.Context, req *ListLeaveRequestsRequest) (ListLeaveRequestsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListLeaveRequestsResponse{}, onekiterrors.PermissionDenied()
	}
	if req == nil {
		return ListLeaveRequestsResponse{}, onekiterrors.New(http.StatusBadRequest, "Leave list request is required")
	}

	var fromDate, toDate time.Time
	if req.FromDate != nil && *req.FromDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.FromDate)
		if err != nil {
			return ListLeaveRequestsResponse{}, onekiterrors.New(http.StatusBadRequest, "from_date must be in YYYY-MM-DD format")
		}
		fromDate = parsed
	}
	if req.ToDate != nil && *req.ToDate != "" {
		parsed, err := time.Parse("2006-01-02", *req.ToDate)
		if err != nil {
			return ListLeaveRequestsResponse{}, onekiterrors.New(http.StatusBadRequest, "to_date must be in YYYY-MM-DD format")
		}
		toDate = parsed
	}
	if !fromDate.IsZero() && !toDate.IsZero() && fromDate.After(toDate) {
		return ListLeaveRequestsResponse{}, onekiterrors.New(http.StatusBadRequest, "from_date must not be after to_date")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultLeaveRequestsLimit
	}
	if limit > maxLeaveRequestsLimit {
		limit = maxLeaveRequestsLimit
	}

	businessID, businessErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if businessErr != nil {
		return ListLeaveRequestsResponse{}, onekiterrors.PermissionDenied()
	}

	requests, total, err := h.Service.ListLeaveRequests(ctx, businessID, device.User.UserID, domain.ListLeaveRequestsParams{
		EmployeeID: req.EmployeeID,
		Status:     req.Status,
		FromDate:   req.FromDate,
		ToDate:     req.ToDate,
		Limit:      limit,
		Offset:     req.Offset,
	})
	if err != nil {
		if errors.Is(err, domain.ErrLeaveAccessDenied) {
			return ListLeaveRequestsResponse{}, onekiterrors.PermissionDenied()
		}
		logger.Ctx(ctx).Error("failed to list leave requests", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return ListLeaveRequestsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list leave requests")
	}

	payloads := make([]*hrleavev1.LeaveRequestPayload, 0, len(requests))
	for _, r := range requests {
		payloads = append(payloads, leaveRequestToProto(r))
	}

	return ListLeaveRequestsResponse{Requests: payloads, Total: total, Limit: limit, Offset: req.Offset}, nil
}
