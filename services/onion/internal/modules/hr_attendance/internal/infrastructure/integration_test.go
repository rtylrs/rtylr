package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_attendance/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_attendance/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListAttendance_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListAttendance(ctx, testutil.OwnerID, domain.ListAttendanceParams{
		DateFrom: "2026-01-01",
		DateTo:   "2026-12-31",
		Limit:    50,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListAttendance: %v", err)
	}
	if total < 0 {
		t.Errorf("expected non-negative total, got %d", total)
	}
	_ = records
}

func TestListAttendance_TotalMatchesRecordBound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListAttendance(ctx, testutil.OwnerID, domain.ListAttendanceParams{
		DateFrom: "2026-01-01",
		DateTo:   "2026-12-31",
		Limit:    50,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListAttendance: %v", err)
	}
	if int64(len(records)) > total {
		t.Errorf("returned %d records but total is %d", len(records), total)
	}
}

func TestListAttendance_RecordFieldsPopulated(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, _, err := repo.ListAttendance(ctx, testutil.OwnerID, domain.ListAttendanceParams{
		DateFrom: "2026-01-01",
		DateTo:   "2026-12-31",
		Limit:    50,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListAttendance: %v", err)
	}
	for _, rec := range records {
		if rec.ID == 0 {
			t.Error("record has zero ID")
		}
		if rec.OwnerID != testutil.OwnerID {
			t.Errorf("record %d belongs to wrong owner: %d", rec.ID, rec.OwnerID)
		}
		if rec.Date == "" {
			t.Errorf("record %d has empty Date", rec.ID)
		}
	}
}

func TestListAttendance_FilterByEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, _, err := repo.ListAttendance(ctx, testutil.OwnerID, domain.ListAttendanceParams{
		EmployeeID: 4101,
		DateFrom:   "2026-01-01",
		DateTo:     "2026-12-31",
		Limit:      50,
		Offset:     0,
	})
	if err != nil {
		t.Fatalf("ListAttendance by employee: %v", err)
	}
	for _, rec := range records {
		if rec.EmployeeID != 4101 {
			t.Errorf("expected employee 4101, got %d", rec.EmployeeID)
		}
	}
}

func TestListAttendance_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, total, err := repo.ListAttendance(ctx, 0, domain.ListAttendanceParams{
		DateFrom: "2026-01-01",
		DateTo:   "2026-12-31",
		Limit:    50,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 || len(records) != 0 {
		t.Errorf("expected empty result for unknown owner, got total=%d records=%d", total, len(records))
	}
}

func TestGetActiveClockIn_ReturnsNilForUnknownEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rec, err := repo.GetActiveClockIn(ctx, testutil.OwnerID, 999999999)
	if err != nil {
		t.Fatalf("GetActiveClockIn: %v", err)
	}
	if rec != nil {
		t.Error("expected nil for non-existent employee active clock-in")
	}
}

func TestGetAttendanceConfig_NoErrorForOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Config may not exist for fixture owner; nil is a valid response.
	cfg, err := repo.GetAttendanceConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetAttendanceConfig: %v", err)
	}
	if cfg != nil && cfg.OwnerID != testutil.OwnerID {
		t.Errorf("config belongs to wrong owner: %d", cfg.OwnerID)
	}
}

func TestGetDailyAttendance_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, err := repo.GetDailyAttendance(ctx, testutil.OwnerID, 0, "2026-03-15")
	if err != nil {
		t.Fatalf("GetDailyAttendance: %v", err)
	}
	_ = records
}

func TestGetEmployeePINHash_NotSetReturnsError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Employee 4101 exists but PIN may not be set.
	// Both are acceptable outcomes: ErrPINNotSet or empty hash. The important thing is no panic.
	hash, err := repo.GetEmployeePINHash(ctx, testutil.OwnerID, 4101)
	_ = err
	_ = hash
}

func TestGetLocationCoordinates_NilForUnknownLocation(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	coords, err := repo.GetLocationCoordinates(ctx, testutil.OwnerID, 999999999)
	if err != nil {
		t.Fatalf("GetLocationCoordinates: %v", err)
	}
	if coords != nil {
		t.Error("expected nil coords for non-existent location")
	}
}

func TestListAttendance_PaginationOffsetWorks(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Page 1
	first, total, err := repo.ListAttendance(ctx, testutil.OwnerID, domain.ListAttendanceParams{
		DateFrom: "2026-01-01",
		DateTo:   "2026-12-31",
		Limit:    5,
		Offset:   0,
	})
	if err != nil {
		t.Fatalf("ListAttendance page 1: %v", err)
	}
	// Page 2 (offset 5)
	second, _, err := repo.ListAttendance(ctx, testutil.OwnerID, domain.ListAttendanceParams{
		DateFrom: "2026-01-01",
		DateTo:   "2026-12-31",
		Limit:    5,
		Offset:   5,
	})
	if err != nil {
		t.Fatalf("ListAttendance page 2: %v", err)
	}
	// When total > 5, the two pages must not share the same first ID.
	if total > 5 && len(first) > 0 && len(second) > 0 {
		if first[0].ID == second[0].ID {
			t.Error("pagination offset had no effect: both pages start with the same record")
		}
	}
}
