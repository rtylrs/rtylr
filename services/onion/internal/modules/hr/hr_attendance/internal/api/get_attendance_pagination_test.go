package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
)

type recordingAttendanceService struct {
	params domain.ListAttendanceParams
}

func (s *recordingAttendanceService) GetMyAttendance(_ context.Context, _, _ int64, params domain.ListAttendanceParams) ([]domain.AttendanceRecordRecord, int64, error) {
	s.params = params
	return nil, 0, nil
}

func attendanceViewContext() context.Context {
	return authPkg.WithAuthUser(context.Background(), &authPkg.AuthUser{
		BusinessID: 9,
		UserID:     7,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHR,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"hr:employees:view"},
			}},
		},
	})
}

func TestGetAttendanceUsesSelfResolvedIdentityAndPagination(t *testing.T) {
	service := &recordingAttendanceService{}
	handler := &GetAttendanceHandler{
		Service: service,
		Resolve: func(context.Context) (int64, int64, error) {
			return 9, 42, nil
		},
	}

	response, err := handler.Serve(attendanceViewContext(), &GetAttendanceRequest{
		DateFrom: "2026-01-01",
		DateTo:   "2026-01-31",
		Limit:    7,
		Offset:   14,
	})
	if err != nil {
		t.Fatalf("Serve: %v", err)
	}

	if service.params.Limit != 7 || service.params.Offset != 14 {
		t.Fatalf("service pagination = limit %d offset %d, want 7/14", service.params.Limit, service.params.Offset)
	}
	if response.Limit != 7 || response.Offset != 14 {
		t.Fatalf("response pagination = limit %d offset %d, want 7/14", response.Limit, response.Offset)
	}
}

func TestGetAttendanceRejectsNegativeOffset(t *testing.T) {
	service := &recordingAttendanceService{}
	handler := &GetAttendanceHandler{
		Service: service,
		Resolve: func(context.Context) (int64, int64, error) {
			return 9, 42, nil
		},
	}

	_, err := handler.Serve(attendanceViewContext(), &GetAttendanceRequest{
		DateFrom: "2026-01-01",
		DateTo:   "2026-01-31",
		Offset:   -1,
	})
	if oneKitStatus(err) != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", oneKitStatus(err), http.StatusBadRequest)
	}
	if service.params != (domain.ListAttendanceParams{}) {
		t.Fatal("service should not be called for invalid pagination")
	}
}
