package infrastructure_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_attendance/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetDailyAttendance_TodayNoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, err := repo.GetDailyAttendance(ctx, testutil.OwnerID, 0, "2026-05-28")
	if err != nil {
		t.Fatalf("GetDailyAttendance for today: %v", err)
	}
	_ = records
}

func TestGetDailyAttendance_WithLocationID_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, err := repo.GetDailyAttendance(ctx, testutil.OwnerID, 1, "2026-05-28")
	if err != nil {
		t.Fatalf("GetDailyAttendance with locationID: %v", err)
	}
	_ = records
}

func TestListAttendance_LimitOneReturnsSingleRecord(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rec := domain.AttendanceRecordRecord{
		OwnerID:       testutil.OwnerID,
		EmployeeID:    4101,
		Date:          "2026-05-06",
		ClockInMethod: "pin",
		Status:        "clocked_in",
	}
	created, err := repo.CreateAttendanceRecord(ctx, testutil.OwnerID, rec)
	if err != nil {
		t.Fatalf("CreateAttendanceRecord: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_attendance_records WHERE id = ?", created.ID)
	})

	records, total, err := repo.ListAttendance(ctx, testutil.OwnerID, domain.ListAttendanceParams{
		DateFrom: "2026-01-01",
		DateTo:   "2026-12-31",
		Limit:    1,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListAttendance limit=1: %v", err)
	}
	if len(records) > 1 {
		t.Errorf("expected at most 1 record with limit=1, got %d", len(records))
	}
	if total < 1 {
		t.Errorf("total should be at least 1 after insert, got %d", total)
	}
}

func TestListAttendance_NarrowDateRangeFiltersCorrectly(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, _, err := repo.ListAttendance(ctx, testutil.OwnerID, domain.ListAttendanceParams{
		DateFrom: "2000-01-01",
		DateTo:   "2000-01-01",
		Limit:    50,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListAttendance narrow range: %v", err)
	}
	for _, r := range records {
		if r.Date != "2000-01-01" {
			t.Errorf("record %d has date %s outside narrow range", r.ID, r.Date)
		}
	}
}

func TestGetAttendanceConfig_UnknownOwnerReturnsNil(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetAttendanceConfig(ctx, 999999999)
	if err != nil {
		t.Fatalf("GetAttendanceConfig(unknown owner): %v", err)
	}
	if cfg != nil {
		t.Error("expected nil config for unknown owner")
	}
}

func TestCreateAttendanceRecord_DatePreserved(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantDate := "2026-05-07"
	rec := domain.AttendanceRecordRecord{
		OwnerID:       testutil.OwnerID,
		EmployeeID:    4101,
		Date:          wantDate,
		ClockInMethod: "pin",
		Status:        "clocked_in",
	}
	created, err := repo.CreateAttendanceRecord(ctx, testutil.OwnerID, rec)
	if err != nil {
		t.Fatalf("CreateAttendanceRecord: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_attendance_records WHERE id = ?", created.ID)
	})
	if created.Date != wantDate {
		t.Errorf("Date mismatch: got %q, want %q", created.Date, wantDate)
	}
}

func TestCreateAttendanceRecord_EmployeeIDPreserved(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rec := domain.AttendanceRecordRecord{
		OwnerID:       testutil.OwnerID,
		EmployeeID:    4101,
		Date:          "2026-05-08",
		ClockInMethod: "pin",
		Status:        "clocked_in",
	}
	created, err := repo.CreateAttendanceRecord(ctx, testutil.OwnerID, rec)
	if err != nil {
		t.Fatalf("CreateAttendanceRecord: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_attendance_records WHERE id = ?", created.ID)
	})
	if created.EmployeeID != 4101 {
		t.Errorf("EmployeeID mismatch: got %d, want 4101", created.EmployeeID)
	}
}

func TestUpdateAttendanceRecord_ReturnsUpdatedRecord(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rec := domain.AttendanceRecordRecord{
		OwnerID:       testutil.OwnerID,
		EmployeeID:    4101,
		Date:          "2026-05-09",
		ClockInMethod: "pin",
		Status:        "clocked_in",
	}
	created, err := repo.CreateAttendanceRecord(ctx, testutil.OwnerID, rec)
	if err != nil {
		t.Fatalf("CreateAttendanceRecord: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_attendance_records WHERE id = ?", created.ID)
	})

	created.IsLate = true
	created.LateMinutes = 15
	updated, err := repo.UpdateAttendanceRecord(ctx, testutil.OwnerID, *created)
	if err != nil {
		t.Fatalf("UpdateAttendanceRecord: %v", err)
	}
	if updated == nil {
		t.Fatal("expected non-nil updated record")
	}
	if updated.ID != created.ID {
		t.Errorf("ID mismatch after update: got %d, want %d", updated.ID, created.ID)
	}
}

// Regression: ClockIn used to check for an active record and insert in two
// separate statements, so two overlapping clock-ins (kiosk double-tap, retry,
// two devices) both inserted an open row. ClockOut then closed only the newest
// row and the orphan made every later ClockIn return ErrAlreadyClockedIn. The
// guarded create must serialize the check and the insert.
func TestCreateAttendanceRecord_ConcurrentClockInsCreateOneOpenRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	const employeeID int64 = 4101
	if _, err := db.ExecContext(ctx, "DELETE FROM hr_attendance_records WHERE business_id = ? AND employee_id = ?", testutil.BusinessID, employeeID); err != nil {
		t.Fatalf("clear attendance records for employee %d: %v", employeeID, err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_attendance_records WHERE business_id = ? AND employee_id = ?", testutil.BusinessID, employeeID)
	})

	const attempts = 4
	start := make(chan struct{})
	results := make([]error, attempts)
	var wg sync.WaitGroup
	for i := 0; i < attempts; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			_, results[i] = repo.CreateAttendanceRecord(ctx, testutil.BusinessID, domain.AttendanceRecordRecord{
				EmployeeID:    employeeID,
				Date:          "2026-06-01",
				ClockInMethod: "pin",
				Status:        "clocked_in",
			})
		}(i)
	}
	close(start)
	wg.Wait()

	created := 0
	for i, err := range results {
		switch {
		case err == nil:
			created++
		case errors.Is(err, domain.ErrAlreadyClockedIn):
			// The loser of the race must be told the employee is already
			// clocked in, not served a generic 500.
		default:
			t.Fatalf("attempt %d: unexpected error: %v", i, err)
		}
	}
	if created != 1 {
		t.Fatalf("concurrent clock-ins succeeded %d times, want exactly 1", created)
	}

	var open int
	if err := db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM hr_attendance_records
		WHERE business_id = ? AND employee_id = ? AND clock_out_at IS NULL`,
		testutil.BusinessID, employeeID).Scan(&open); err != nil {
		t.Fatalf("count open attendance records: %v", err)
	}
	if open != 1 {
		t.Fatalf("found %d open attendance records, want exactly 1", open)
	}
}
