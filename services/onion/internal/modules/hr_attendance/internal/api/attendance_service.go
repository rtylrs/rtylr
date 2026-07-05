package api

import (
	"context"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

type AttendanceAPI struct {
	Service *application.Service
}

func NewAttendanceAPI(service *application.Service) *AttendanceAPI {
	return &AttendanceAPI{Service: service}
}

func (a *AttendanceAPI) ListAttendance(ctx context.Context, req *hrattendancev1.ListAttendanceRequest) (*hrattendancev1.ListAttendanceResponse, error) {
	date := req.GetDate()
	result, err := (&GetAttendanceHandler{Service: a.Service}).Serve(ctx, &GetAttendanceRequest{
		EmployeeID: req.GetEmployeeId(),
		DateFrom:   date,
		DateTo:     date,
		Limit:      50,
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
	device := authPkg.DeviceCtx(ctx)
	var ownerID, employeeID int64
	if device.User != nil {
		ownerID = device.User.UserID
		employeeID = device.User.UserID
	}
	result, err := (&ClockInHandler{Service: a.Service}).Serve(ctx, &ClockInRequest{
		OwnerID:       ownerID,
		EmployeeID:    employeeID,
		LocationID:    req.GetLocationId(),
		ClockInMethod: req.GetMethod(),
		ClockInLat:    req.GetLat(),
		ClockInLng:    req.GetLng(),
		QRToken:       req.GetQrToken(),
		Pin:           req.GetPin(),
	})
	if err != nil {
		return nil, err
	}
	return &hrattendancev1.AttendanceClockInResponse{Record: result.Record}, nil
}

func (a *AttendanceAPI) AttendanceClockOut(ctx context.Context, req *hrattendancev1.AttendanceClockOutRequest) (*hrattendancev1.AttendanceClockOutResponse, error) {
	result, err := (&ClockOutHandler{Service: a.Service}).Serve(ctx, &ClockOutRequest{
		EmployeeID: req.GetAttendanceRecordId(),
	})
	if err != nil {
		return nil, err
	}
	return &hrattendancev1.AttendanceClockOutResponse{Record: result.Record}, nil
}
