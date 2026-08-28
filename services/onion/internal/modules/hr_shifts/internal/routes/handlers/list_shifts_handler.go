package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrshiftv1 "github.com/voxire/voxie/gen/go/hr/shift/v1"
)

func ListShiftsRoutes() []string { return []string{"/hr/shifts"} }

// defaultListShiftsLimit and maxListShiftsLimit bound the client-supplied
// limit before it ever reaches the service layer: an unset or non-positive
// value falls back to the default, and anything above the max is capped, so
// a request like ?limit=1000000 can't force an unbounded read.
const (
	defaultListShiftsLimit int32 = 50
	maxListShiftsLimit     int32 = 100
)

type ListShiftsHandler struct{ Service *application.Service }

// ListShiftsRequest filters the listing. Every filter is optional; with none
// set the response covers both template and attendance rows, because both live
// in the same table and both are read through this route.
type ListShiftsRequest struct {
	Limit      int32  `json:"limit"`
	Offset     int32  `json:"offset"`
	StaffID    int64  `json:"staff_id"`
	Date       string `json:"date"`
	LocationID int64  `json:"location_id"`
	Kind       string `json:"kind"`
}

type ListShiftsResponse struct {
	Shifts []*hrshiftv1.ShiftPayload `json:"shifts"`
	Total  int64                     `json:"total"`
	Limit  int32                     `json:"limit"`
	Offset int32                     `json:"offset"`
}

func (h *ListShiftsHandler) Serve(ctx context.Context, req *ListShiftsRequest) (ListShiftsResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ListShiftsResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireShiftPermission(ctx); err != nil {
		return ListShiftsResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ListShiftsResponse{}, onekiterrors.PermissionDenied()
	}

	kind := domain.ShiftKind(req.Kind)
	if !domain.IsValidShiftKind(kind) {
		return ListShiftsResponse{}, onekiterrors.New(http.StatusBadRequest, "kind must be either template or attendance")
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultListShiftsLimit
	}
	if limit > maxListShiftsLimit {
		limit = maxListShiftsLimit
	}

	shifts, total, err := h.Service.ListShifts(ctx, businessID, domain.ListShiftsParams{
		Limit:      limit,
		Offset:     req.Offset,
		StaffID:    req.StaffID,
		Date:       req.Date,
		LocationID: req.LocationID,
		Kind:       kind,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to list shifts", logger.Any("business_id", businessID), logger.Error(err))
		return ListShiftsResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to list shifts")
	}

	payloads := make([]*hrshiftv1.ShiftPayload, 0, len(shifts))
	for _, s := range shifts {
		payloads = append(payloads, shiftToProto(s))
	}

	return ListShiftsResponse{Shifts: payloads, Total: total, Limit: limit, Offset: req.Offset}, nil
}
