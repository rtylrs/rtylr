package api

import (
	"context"
	"net/http"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

func GetAttendanceRoutes() []string { return []string{"/hr/attendance"} }

const (
	minimumAttendanceDate = "1000-01-01"
	maximumAttendanceDate = "9999-12-31"
)

type getAttendanceService interface {
	GetMyAttendance(ctx context.Context, businessID, employeeID int64, params domain.ListAttendanceParams) ([]domain.AttendanceRecordRecord, int64, error)
}

type GetAttendanceHandler struct {
	Service getAttendanceService
	Resolve func(ctx context.Context) (businessID, employeeID int64, err error)
}

type GetAttendanceRequest struct {
	OwnerID    int64  `json:"owner_id"`
	EmployeeID int64  `json:"employee_id"`
	DateFrom   string `json:"date_from"`
	DateTo     string `json:"date_to"`
	Limit      int32  `json:"limit"`
	Offset     int32  `json:"offset"`
}

type GetAttendanceResponse struct {
	Records []*hrattendancev1.AttendanceRecordPayload `json:"records"`
	Total   int64                                     `json:"total"`
	Limit   int32                                     `json:"limit"`
	Offset  int32                                     `json:"offset"`
}

func normalizeAttendanceDateRange(dateFrom, dateTo string) (string, string, error) {
	if dateFrom == "" {
		dateFrom = minimumAttendanceDate
	}
	if dateTo == "" {
		dateTo = maximumAttendanceDate
	}

	from, err := time.Parse(time.DateOnly, dateFrom)
	if err != nil {
		return "", "", onekiterrors.New(http.StatusBadRequest, "date_from must use YYYY-MM-DD")
	}
	to, err := time.Parse(time.DateOnly, dateTo)
	if err != nil {
		return "", "", onekiterrors.New(http.StatusBadRequest, "date_to must use YYYY-MM-DD")
	}
	if from.After(to) {
		return "", "", onekiterrors.New(http.StatusBadRequest, "date_from must not be after date_to")
	}
	return dateFrom, dateTo, nil
}

func (h *GetAttendanceHandler) Serve(ctx context.Context, req *GetAttendanceRequest) (GetAttendanceResponse, error) {
	if h.Resolve == nil || req == nil {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:employees:view") {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, employeeID, err := h.Resolve(ctx)
	if err != nil {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}
	// This is a self-service endpoint. Older clients sent these ids, while
	// the current client omits them; a supplied value may only agree with the
	// identity resolved above.
	if (req.OwnerID != 0 && req.OwnerID != businessID) ||
		(req.EmployeeID != 0 && req.EmployeeID != employeeID) {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}

	dateFrom, dateTo, err := normalizeAttendanceDateRange(req.DateFrom, req.DateTo)
	if err != nil {
		return GetAttendanceResponse{}, err
	}

	pagination, err := normalizeAttendancePagination(req.Limit, req.Offset)
	if err != nil {
		return GetAttendanceResponse{}, err
	}

	records, total, err := h.Service.GetMyAttendance(ctx, businessID, employeeID, domain.ListAttendanceParams{
		DateFrom: dateFrom,
		DateTo:   dateTo,
		Limit:    pagination.limit,
		Offset:   pagination.offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to get attendance", logger.Any("owner_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return GetAttendanceResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get attendance")
	}

	payloads := make([]*hrattendancev1.AttendanceRecordPayload, 0, len(records))
	for _, r := range records {
		payloads = append(payloads, attendanceToProto(r))
	}

	return GetAttendanceResponse{Records: payloads, Total: total, Limit: pagination.limit, Offset: pagination.offset}, nil
}
