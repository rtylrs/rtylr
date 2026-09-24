package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/onekiterrors"
	"github.com/voxire/onion/internal/platform/tenant"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

type EmployeeResolver interface {
	GetEmployeeIDByUserID(ctx context.Context, ownerID, userID int64) (int64, error)
}

type AttendanceAPI struct {
	Service          *application.Service
	EmployeeResolver EmployeeResolver
}

func NewAttendanceAPI(service *application.Service, employeeResolver EmployeeResolver) *AttendanceAPI {
	return &AttendanceAPI{Service: service, EmployeeResolver: employeeResolver}
}

func (a *AttendanceAPI) resolveIdentity(ctx context.Context) (int64, int64, error) {
	device := authPkg.DeviceCtx(ctx)
	if device.IsAnonymous || device.User == nil || a.EmployeeResolver == nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}

	businessID, err := tenant.ResolveBusinessID(ctx, device.User.BusinessID, device.User.UserID)
	if err != nil {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	employeeID, err := a.EmployeeResolver.GetEmployeeIDByUserID(ctx, businessID, device.User.UserID)
	if err != nil || employeeID <= 0 {
		return 0, 0, onekiterrors.PermissionDenied()
	}
	return businessID, employeeID, nil
}

// ResolveIdentity exposes the same self-service scope used by the generated
// attendance service to the compatibility routes registered by this module.
func (a *AttendanceAPI) ResolveIdentity(ctx context.Context) (int64, int64, error) {
	return a.resolveIdentity(ctx)
}

func (a *AttendanceAPI) ListAttendance(ctx context.Context, req *hrattendancev1.ListAttendanceRequest) (*hrattendancev1.ListAttendanceResponse, error) {
	ownerID, employeeID, err := a.resolveIdentity(ctx)
	if err != nil {
		return nil, err
	}
	date := req.GetDate()
	pagination := attendancePaginationFromContext(ctx)
	result, err := (&GetAttendanceHandler{Service: a.Service, Resolve: a.resolveIdentity}).Serve(ctx, &GetAttendanceRequest{
		OwnerID:    ownerID,
		EmployeeID: employeeID,
		DateFrom:   date,
		DateTo:     date,
		Limit:      pagination.limit,
		Offset:     pagination.offset,
	})
	if err != nil {
		return nil, err
	}
	return &hrattendancev1.ListAttendanceResponse{
		Records: result.Records,
		Total:   result.Total,
		Limit:   result.Limit,
		Offset:  result.Offset,
	}, nil
}

func (a *AttendanceAPI) AttendanceClockIn(ctx context.Context, req *hrattendancev1.AttendanceClockInRequest) (*hrattendancev1.AttendanceClockInResponse, error) {
	ownerID, employeeID, err := a.resolveIdentity(ctx)
	if err != nil {
		return nil, err
	}
	handler := &ClockInHandler{Service: a.Service}
	clockInReq := &ClockInRequest{
		OwnerID:       ownerID,
		EmployeeID:    employeeID,
		LocationID:    req.GetLocationId(),
		ClockInMethod: req.GetMethod(),
		ClockInLat:    req.GetLat(),
		ClockInLng:    req.GetLng(),
		QRToken:       req.GetQrToken(),
		Pin:           req.GetPin(),
	}
	if err := handler.Validate(ctx, clockInReq); err != nil {
		return nil, err
	}
	result, err := handler.Serve(ctx, clockInReq)
	if err != nil {
		return nil, err
	}
	return &hrattendancev1.AttendanceClockInResponse{Record: result.Record}, nil
}

func (a *AttendanceAPI) AttendanceClockOut(ctx context.Context, req *hrattendancev1.AttendanceClockOutRequest) (*hrattendancev1.AttendanceClockOutResponse, error) {
	_, employeeID, err := a.resolveIdentity(ctx)
	if err != nil {
		return nil, err
	}

	// The generated request carries an attendance record id, but it is not an
	// authorization input. Clock-out is self-scoped to the authenticated
	// employee, just like clock-in.
	handler := &ClockOutHandler{Service: a.Service, Resolve: a.resolveIdentity}
	clockOutReq := &ClockOutRequest{EmployeeID: employeeID}
	if err := handler.Validate(ctx, clockOutReq); err != nil {
		return nil, err
	}
	result, err := handler.Serve(ctx, clockOutReq)
	if err != nil {
		return nil, err
	}
	return &hrattendancev1.AttendanceClockOutResponse{Record: result.Record}, nil
}
