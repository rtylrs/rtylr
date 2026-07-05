package api

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

func GetLocationAttendanceRoutes() []string { return []string{"/hr/attendance/location/{id}/today"} }

type GetLocationAttendanceHandler struct{ Service *application.Service }

type GetLocationAttendanceRequest struct{}

type GetLocationAttendanceResponse struct {
	Records []*hrattendancev1.AttendanceRecordPayload `json:"records"`
}

func (h *GetLocationAttendanceHandler) Serve(ctx context.Context, _ *GetLocationAttendanceRequest) (GetLocationAttendanceResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetLocationAttendanceResponse{}, onekiterrors.PermissionDenied()
	}

	idStr := chi.URLParamFromCtx(ctx, "id")
	locationID, parseErr := strconv.ParseInt(idStr, 10, 64)
	if parseErr != nil {
		return GetLocationAttendanceResponse{}, onekiterrors.New(http.StatusBadRequest, "Invalid location id")
	}

	today := time.Now().Format("2006-01-02")

	records, err := h.Service.GetLocationAttendance(ctx, device.User.UserID, locationID, today)
	if err != nil {
		logger.Ctx(ctx).Error("failed to get location attendance", logger.Any("owner_id", device.User.UserID), logger.Any("location_id", locationID), logger.Error(err))
		return GetLocationAttendanceResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get location attendance")
	}

	payloads := make([]*hrattendancev1.AttendanceRecordPayload, 0, len(records))
	for _, r := range records {
		payloads = append(payloads, attendanceToProto(r))
	}

	return GetLocationAttendanceResponse{Records: payloads}, nil
}
