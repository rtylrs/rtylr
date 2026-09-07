package api

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

func GetRosterRoutes() []string { return []string{"/hr/shifts/roster"} }

type GetRosterHandler struct{ Service *application.Service }

type GetRosterRequest struct {
	LocationID int64  `json:"location_id"`
	WeekStart  string `json:"week_start"`
}

type GetRosterResponse struct {
	Assignments []*hrshiftv1.ShiftAssignmentPayload `json:"assignments"`
	Conflicts   []domain.ConflictInfo               `json:"conflicts"`
}

func (h *GetRosterHandler) Serve(ctx context.Context, req *GetRosterRequest) (GetRosterResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetRosterResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireShiftPermission(ctx); err != nil {
		return GetRosterResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return GetRosterResponse{}, onekiterrors.PermissionDenied()
	}

	assignments, conflicts, err := h.Service.GetRoster(ctx, businessID, domain.RosterParams{
		LocationID: req.LocationID,
		WeekStart:  req.WeekStart,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to get roster", logger.Any("business_id", businessID), logger.Error(err))
		return GetRosterResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get roster")
	}

	payloads := make([]*hrshiftv1.ShiftAssignmentPayload, 0, len(assignments))
	for _, a := range assignments {
		payloads = append(payloads, assignmentToProto(a))
	}

	return GetRosterResponse{Assignments: payloads, Conflicts: conflicts}, nil
}
