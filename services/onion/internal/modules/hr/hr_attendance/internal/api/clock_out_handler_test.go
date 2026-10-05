package api

import (
	"context"
	"errors"
	"net/http"
	"testing"

	onekithttp "github.com/voxire/onion/internal/platform/onekiterrors"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
)

type recordingClockOutService struct {
	calls      int
	businessID int64
	employeeID int64
}

func (s *recordingClockOutService) ClockOut(_ context.Context, businessID, employeeID int64) (*domain.AttendanceRecordRecord, error) {
	s.calls++
	s.businessID = businessID
	s.employeeID = employeeID
	return &domain.AttendanceRecordRecord{ID: 1, EmployeeID: employeeID}, nil
}

func clockOutTestContext(businessID, userID int64) context.Context {
	device := &authPkg.Device{User: &authPkg.AuthUser{BusinessID: businessID, UserID: userID}}
	return device.Ctx(context.Background())
}

func clockOutStatus(err error) int {
	var httpErr *onekithttp.Error
	if errors.As(err, &httpErr) {
		return int(httpErr.GetStatusCode())
	}
	return 0
}

func TestClockOutRejectsCallerSuppliedEmployeeID(t *testing.T) {
	service := &recordingClockOutService{}
	handler := &ClockOutHandler{
		Service: service,
		Resolve: func(context.Context) (int64, int64, error) {
			return 101, 202, nil
		},
	}

	_, err := handler.Serve(clockOutTestContext(101, 7), &ClockOutRequest{
		EmployeeID: 999,
	})
	if got := clockOutStatus(err); got != http.StatusForbidden {
		t.Fatalf("cross-employee clock-out status = %d, want %d (err: %v)", got, http.StatusForbidden, err)
	}
	if service.calls != 0 {
		t.Fatalf("clock-out service was called %d times for a rejected identity", service.calls)
	}
}

func TestClockOutUsesAuthenticatedEmployee(t *testing.T) {
	service := &recordingClockOutService{}
	handler := &ClockOutHandler{
		Service: service,
		Resolve: func(context.Context) (int64, int64, error) {
			return 101, 202, nil
		},
	}

	result, err := handler.Serve(clockOutTestContext(101, 7), &ClockOutRequest{
		EmployeeID: 202,
	})
	if err != nil {
		t.Fatalf("clock-out: %v", err)
	}
	if service.businessID != 101 || service.employeeID != 202 {
		t.Fatalf("service scope = (%d, %d), want (101, 202)", service.businessID, service.employeeID)
	}
	if result.Record == nil || result.Record.EmployeeId != 202 {
		t.Fatalf("clock-out response employee = %+v, want 202", result.Record)
	}
}
