package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

func GetAttendanceRoutes() []string { return []string{"/hr/attendance"} }

type EmployeeBusinessResolver interface {
	GetEmployeeBusinessID(ctx context.Context, ownerID, employeeID int64) (int64, error)
}

type GetAttendanceHandler struct {
	Service                  *application.Service
	EmployeeBusinessResolver EmployeeBusinessResolver
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

func (h *GetAttendanceHandler) Serve(ctx context.Context, req *GetAttendanceRequest) (GetAttendanceResponse, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil || req.OwnerID <= 0 || req.EmployeeID <= 0 {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}
	if !authPkg.HasAppPermission(device.User, authorization.AppHR, "hr:employees:view") {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil || businessID != req.OwnerID {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}
	if h.EmployeeBusinessResolver == nil {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}
	employeeBusinessID, err := h.EmployeeBusinessResolver.GetEmployeeBusinessID(ctx, businessID, req.EmployeeID)
	if err != nil || employeeBusinessID != businessID {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	records, total, err := h.Service.GetMyAttendance(ctx, businessID, req.EmployeeID, domain.ListAttendanceParams{
		DateFrom: req.DateFrom,
		DateTo:   req.DateTo,
		Limit:    limit,
		Offset:   req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to get attendance", logger.Any("owner_id", req.OwnerID), logger.Any("employee_id", req.EmployeeID), logger.Error(err))
		return GetAttendanceResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get attendance")
	}

	payloads := make([]*hrattendancev1.AttendanceRecordPayload, 0, len(records))
	for _, r := range records {
		payloads = append(payloads, attendanceToProto(r))
	}

	return GetAttendanceResponse{Records: payloads, Total: total, Limit: limit, Offset: req.Offset}, nil
}
