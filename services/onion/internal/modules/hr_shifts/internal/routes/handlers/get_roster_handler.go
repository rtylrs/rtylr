package handlers

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_shifts/internal/application"
	"github.com/voxire/onion/internal/modules/hr_shifts/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	errorModels "github.com/voxire/onion/internal/platform/errors"
	"github.com/voxire/onion/internal/platform/logger"
	hrv1 "github.com/voxire/voxie/gen/go/hr/v1"
)

func GetRosterRoutes() []string { return []string{"/hr/shifts/roster"} }

type GetRosterHandler struct{ Service *application.Service }

type GetRosterRequest struct {
	LocationID int64  `json:"location_id"`
	WeekStart  string `json:"week_start"`
}

type GetRosterResponse struct {
	Assignments []*hrv1.ShiftAssignmentPayload `json:"assignments"`
	Conflicts   []domain.ConflictInfo          `json:"conflicts"`
}

func (h *GetRosterHandler) Serve(ctx context.Context, req *GetRosterRequest) (GetRosterResponse, *errorModels.ServerError) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetRosterResponse{}, errorModels.PermissionDeniedServerError()
	}

	assignments, conflicts, err := h.Service.GetRoster(ctx, device.User.UserID, domain.RosterParams{
		LocationID: req.LocationID,
		WeekStart:  req.WeekStart,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to get roster", logger.Any("owner_id", device.User.UserID), logger.Error(err))
		return GetRosterResponse{}, &errorModels.ServerError{HTTPCode: http.StatusInternalServerError, Code: http.StatusInternalServerError, Message: "Failed to get roster"}
	}

	payloads := make([]*hrv1.ShiftAssignmentPayload, 0, len(assignments))
	for _, a := range assignments {
		payloads = append(payloads, assignmentToProto(a))
	}

	return GetRosterResponse{Assignments: payloads, Conflicts: conflicts}, nil
}
