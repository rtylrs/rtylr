package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/platform/authorization"
	"github.com/voxire/onion/internal/testutil"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

func TestAttendanceV1ListHonorsPaginationQuery(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		businessID = int64(917001)
		userID     = int64(917002)
		employeeID = int64(917003)
		date       = "2030-12-31"
	)

	testutil.GrantBusinessAccess(t, db, businessID, userID)
	if _, err := db.ExecContext(ctx, `DELETE FROM hr_attendance_records WHERE business_id = ? AND employee_id = ?`, businessID, employeeID); err != nil {
		t.Fatalf("clear attendance fixture: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, `DELETE FROM hr_attendance_records WHERE business_id = ? AND employee_id = ?`, businessID, employeeID)
	})

	for i := 0; i < 51; i++ {
		clockIn := time.Date(2030, 12, 31, 0, i, 0, 0, time.UTC)
		if _, err := db.ExecContext(ctx, `
			INSERT INTO hr_attendance_records
			(owner_id, business_id, employee_id, location_id, date, clock_in_at,
			 clock_out_at, duration_minutes, clock_in_method, status)
			VALUES (?, ?, ?, 0, ?, ?, ?, 1, 'app', 'present')`,
			businessID, businessID, employeeID, date, clockIn, clockIn.Add(time.Minute),
		); err != nil {
			t.Fatalf("seed attendance row %d: %v", i, err)
		}
	}

	mux := http.NewServeMux()
	if err := hrattendancev1.RegisterAttendanceServiceServer(
		NewAttendanceAPI(application.NewService(ctx, moduleinfra.NewRepository(db)), &employeeResolverStub{employeeID: employeeID}),
		hrattendancev1.WithMux(mux),
		hrattendancev1.WithMiddleware(AttendancePaginationMiddleware),
	); err != nil {
		t.Fatalf("register attendance service: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("%s?date=%s&limit=1&offset=50", attendanceListPath, date), nil)
	req = req.WithContext(authPkg.WithAuthUser(req.Context(), &authPkg.AuthUser{
		BusinessID: businessID,
		UserID:     userID,
		EffectiveAccess: &authorization.EffectiveAccess{
			Apps: []authorization.EffectiveAppAccess{{
				AppID:       authorization.AppHR,
				Licensed:    true,
				HasAccess:   true,
				Permissions: []string{"hr:employees:view"},
			}},
		},
	}))
	recorder := httptest.NewRecorder()
	mux.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	var response hrattendancev1.ListAttendanceResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Total != 51 || response.Limit != 1 || response.Offset != 50 {
		t.Fatalf("pagination metadata = total %d limit %d offset %d, want 51/1/50", response.Total, response.Limit, response.Offset)
	}
	if len(response.Records) != 1 {
		t.Fatalf("records = %d, want 1", len(response.Records))
	}
}
