package api

import (
	"context"
	"net/http"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

func GetAttendanceRoutes() []string { return []string{"/hr/attendance"} }

// GetAttendanceHandler serves attendance history. Identity is resolved
// server-side from the JWT — client-supplied owner/employee ids are ignored,
// since trusting them let any authenticated tenant read any other tenant's
// payroll-sensitive attendance (#817).
type GetAttendanceHandler struct {
	Service *application.Service
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

func (h *GetAttendanceHandler) Serve(ctx context.Context, req *GetAttendanceRequest) (GetAttendanceResponse, error) {
	if h.Resolve == nil {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}
	businessID, employeeID, err := h.Resolve(ctx)
	if err != nil {
		return GetAttendanceResponse{}, onekiterrors.PermissionDenied()
	}
	_ = req.OwnerID
	_ = req.EmployeeID

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	records, total, err := h.Service.GetMyAttendance(ctx, businessID, employeeID, domain.ListAttendanceParams{
		DateFrom: req.DateFrom,
		DateTo:   req.DateTo,
		Limit:    limit,
		Offset:   req.Offset,
	})
	if err != nil {
		logger.Ctx(ctx).Error("failed to get attendance", logger.Any("owner_id", businessID), logger.Any("employee_id", employeeID), logger.Error(err))
		return GetAttendanceResponse{}, onekiterrors.New(http.StatusInternalServerError, "Failed to get attendance")
	}

	payloads := make([]*hrattendancev1.AttendanceRecordPayload, 0, len(records))
	for _, r := range records {
		payloads = append(payloads, attendanceToProto(r))
	}

	return GetAttendanceResponse{Records: payloads, Total: total, Limit: limit, Offset: req.Offset}, nil
}
