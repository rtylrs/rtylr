package api

import (
	"context"
	"testing"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/application"
	authPkg "github.com/voxire/onion/internal/platform/auth"
	"github.com/voxire/onion/internal/testutil"
	hrattendancev1 "github.com/voxire/voxie/gen/go/hr/attendance/v1"
)

type clockOutEmployeeResolver struct {
	employeeID int64
}

func (r clockOutEmployeeResolver) GetEmployeeIDByUserID(context.Context, int64, int64) (int64, error) {
	return r.employeeID, nil
}

func TestAttendanceClockOutUsesAuthenticatedEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()

	const (
		businessID       int64 = 9830101
		userID           int64 = 9830102
		callerEmployeeID int64 = 9830103
		targetEmployeeID int64 = 9830104
	)

	testutil.GrantBusinessAccess(t, db, businessID, userID)

	insertOpenRecord := func(employeeID int64) int64 {
		t.Helper()
		result, err := db.ExecContext(ctx, `
			INSERT INTO hr_attendance_records
				(owner_id, business_id, employee_id, date, clock_in_at, clock_in_method, status)
			VALUES (?, ?, ?, CURDATE(), UTC_TIMESTAMP(), 'app', 'present')`,
			businessID, businessID, employeeID)
		if err != nil {
			t.Fatalf("seed attendance record for employee %d: %v", employeeID, err)
		}
		recordID, err := result.LastInsertId()
		if err != nil {
			t.Fatalf("attendance record id for employee %d: %v", employeeID, err)
		}
		return recordID
	}

	callerRecordID := insertOpenRecord(callerEmployeeID)
	targetRecordID := insertOpenRecord(targetEmployeeID)
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), `
			DELETE FROM hr_attendance_records
			WHERE business_id = ? AND id IN (?, ?)`,
			businessID, callerRecordID, targetRecordID)
	})

	service := application.NewService(ctx, moduleinfra.NewRepository(db))
	attendanceAPI := NewAttendanceAPI(service, clockOutEmployeeResolver{employeeID: callerEmployeeID})
	device := &authPkg.Device{User: &authPkg.AuthUser{BusinessID: businessID, UserID: userID}}
	authCtx := device.Ctx(ctx)

	// The old handler treated this caller-controlled value as an employee id.
	// It must not be allowed to close the other employee's open record.
	attackerValue := targetEmployeeID
	response, err := attendanceAPI.AttendanceClockOut(authCtx, &hrattendancev1.AttendanceClockOutRequest{
		AttendanceRecordId: &attackerValue,
	})
	if err != nil {
		t.Fatalf("AttendanceClockOut: %v", err)
	}
	if response.GetRecord() == nil || response.GetRecord().GetEmployeeId() != callerEmployeeID {
		t.Fatalf("clock-out employee = %+v, want authenticated employee %d", response.GetRecord(), callerEmployeeID)
	}

	var callerClosed, targetClosed bool
	if err := db.QueryRowContext(ctx, `
		SELECT clock_out_at IS NOT NULL
		FROM hr_attendance_records
		WHERE business_id = ? AND id = ?`, businessID, callerRecordID).Scan(&callerClosed); err != nil {
		t.Fatalf("read caller record: %v", err)
	}
	if err := db.QueryRowContext(ctx, `
		SELECT clock_out_at IS NOT NULL
		FROM hr_attendance_records
		WHERE business_id = ? AND id = ?`, businessID, targetRecordID).Scan(&targetClosed); err != nil {
		t.Fatalf("read target record: %v", err)
	}
	if !callerClosed {
		t.Fatal("authenticated employee's open record was not clocked out")
	}
	if targetClosed {
		t.Fatal("caller-supplied employee id closed another employee's record")
	}
}
