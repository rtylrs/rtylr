package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/time_off/internal/application"
	"github.com/voxire/onion/internal/modules/time_off/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
)

func ApproveTimeOffRoutes() []string { return []string{"/hr/time-off/approve"} }

// ─── Approve ─────────────────────────────────────────────────────────────────

type ApproveTimeOffHandler struct{ Service *application.Service }

type ApproveTimeOffRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
	Notes  string `json:"notes"`
}

type ApproveTimeOffResponse struct {
	Success bool `json:"success"`
}

func (h *ApproveTimeOffHandler) Validate(_ context.Context, req *ApproveTimeOffRequest) error {
	if req.ID == 0 {
		return onekiterrors.New(http.StatusBadRequest, "id is required")
	}
	if req.Status != "approved" && req.Status != "rejected" {
		return onekiterrors.New(http.StatusBadRequest, "status must be 'approved' or 'rejected'")
	}
	return nil
}

func (h *ApproveTimeOffHandler) Serve(ctx context.Context, req *ApproveTimeOffRequest) (ApproveTimeOffResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return ApproveTimeOffResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, bizErr := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if bizErr != nil {
		return ApproveTimeOffResponse{}, onekiterrors.PermissionDenied()
	}
	input := domain.TimeOffApprovalInput{
		ID: req.ID, Status: req.Status, ApprovedBy: device.User.UserID, Notes: req.Notes,
	}
	if err := h.Service.ApproveTimeOffRequest(ctx, businessID, input); err != nil {
		return ApproveTimeOffResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to update time off request")
	}
	return ApproveTimeOffResponse{Success: true}, nil
}
