package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr/hr_shifts/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func PublishRosterRoutes() []string { return []string{"/hr/shifts/roster/publish"} }

type PublishRosterHandler struct{ Service *application.Service }

type PublishRosterRequest struct {
	LocationID int64  `json:"location_id"`
	WeekStart  string `json:"week_start"`
}

type PublishRosterResponse struct {
	Success bool `json:"success"`
}

// Validate requires only week_start. location_id is accepted for forward
// compatibility but is not required, because publishing is not scoped by it:
// PublishRosterAssignments matches on business and date range only, so a publish
// covers every location's assignments for that week. Rejecting a request for
// omitting a field that has no effect was a bug. Scoping publish per location is
// a separate behaviour change.
func (h *PublishRosterHandler) Validate(_ context.Context, req *PublishRosterRequest) error {
	if req.WeekStart == "" {
		return onekiterrors.New(http.StatusBadRequest, "week_start is required")
	}
	return nil
}

func (h *PublishRosterHandler) Serve(ctx context.Context, req *PublishRosterRequest) (PublishRosterResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return PublishRosterResponse{}, onekiterrors.PermissionDenied()
	}
	if err := requireShiftPermission(ctx); err != nil {
		return PublishRosterResponse{}, err
	}

	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return PublishRosterResponse{}, onekiterrors.PermissionDenied()
	}

	err := h.Service.PublishRoster(ctx, businessID, req.LocationID, req.WeekStart)
	if err != nil {
		logger.Ctx(ctx).Error("failed to publish roster", logger.Any("business_id", businessID), logger.Error(err))
		return PublishRosterResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to publish roster")
	}

	return PublishRosterResponse{Success: true}, nil
}
