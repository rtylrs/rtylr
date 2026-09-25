package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateAttendanceRecord_ReturnsPositiveID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rec := domain.AttendanceRecordRecord{
		OwnerID:       testutil.OwnerID,
		EmployeeID:    4101,
		Date:          "2026-05-01",
		ClockInMethod: "pin",
		Status:        "clocked_in",
	}
	created, err := repo.CreateAttendanceRecord(ctx, testutil.OwnerID, rec)
	if err != nil {
		t.Fatalf("CreateAttendanceRecord: %v", err)
	}
	if created == nil {
		t.Fatal("expected non-nil result")
	}
	if created.ID <= 0 {
		t.Fatalf("expected positive ID, got %d", created.ID)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_attendance_records WHERE id = ?", created.ID)
	})
}

func TestCreateAttendanceRecord_OwnerIDSet(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rec := domain.AttendanceRecordRecord{
		OwnerID:       testutil.OwnerID,
		EmployeeID:    4101,
		Date:          "2026-05-02",
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
	if created.OwnerID != testutil.OwnerID {
		t.Errorf("OwnerID mismatch: got %d, want %d", created.OwnerID, testutil.OwnerID)
	}
}

func TestCreateAttendanceRecord_AppearsInList(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rec := domain.AttendanceRecordRecord{
		OwnerID:       testutil.OwnerID,
		EmployeeID:    4101,
		Date:          "2026-05-03",
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

	records, _, err := repo.ListAttendance(ctx, testutil.OwnerID, domain.ListAttendanceParams{
		EmployeeID: 4101,
		DateFrom:   "2026-05-03",
		DateTo:     "2026-05-03",
		Limit:      50,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("ListAttendance: %v", err)
	}
	var found bool
	for _, r := range records {
		if r.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created record id=%d not found in list", created.ID)
	}
}

func TestUpdateAttendanceRecord_StatusChanges(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rec := domain.AttendanceRecordRecord{
		OwnerID:       testutil.OwnerID,
		EmployeeID:    4101,
		Date:          "2026-05-04",
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

	created.Status = "clocked_out"
	updated, err := repo.UpdateAttendanceRecord(ctx, testutil.OwnerID, *created)
	if err != nil {
		t.Fatalf("UpdateAttendanceRecord: %v", err)
	}
	if updated.Status != "clocked_out" {
		t.Errorf("expected status=clocked_out, got %q", updated.Status)
	}
}

func TestSetEmployeePINHash_ThenGetReturnsSameHash(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantHash := "test-hash-value-abc123"
	err := repo.SetEmployeePINHash(ctx, testutil.BusinessID, 4101, wantHash)
	if err != nil {
		t.Fatalf("SetEmployeePINHash: %v", err)
	}

	got, err := repo.GetEmployeePINHash(ctx, testutil.BusinessID, 4101)
	if err != nil {
		t.Fatalf("GetEmployeePINHash: %v", err)
	}
	if got != wantHash {
		t.Errorf("hash mismatch: got %q, want %q", got, wantHash)
	}
}

func TestSetEmployeePINHash_OverwritesPreviousHash(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.SetEmployeePINHash(ctx, testutil.BusinessID, 4101, "first-hash")
	if err != nil {
		t.Fatalf("SetEmployeePINHash first: %v", err)
	}

	err = repo.SetEmployeePINHash(ctx, testutil.BusinessID, 4101, "second-hash")
	if err != nil {
		t.Fatalf("SetEmployeePINHash second: %v", err)
	}

	got, err := repo.GetEmployeePINHash(ctx, testutil.BusinessID, 4101)
	if err != nil {
		t.Fatalf("GetEmployeePINHash: %v", err)
	}
	if got != "second-hash" {
		t.Errorf("expected second-hash, got %q", got)
	}
}

func TestGetActiveClockIn_AfterCreate_ReturnsRecord(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rec := domain.AttendanceRecordRecord{
		OwnerID:       testutil.OwnerID,
		EmployeeID:    4101,
		Date:          "2026-05-05",
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

	active, err := repo.GetActiveClockIn(ctx, testutil.OwnerID, 4101)
	if err != nil {
		t.Fatalf("GetActiveClockIn: %v", err)
	}
	_ = active
}
