package infrastructure_test

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/payroll/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListPayrollRecords_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("ListPayrollRecords error: %v", err)
	}
}

func TestListPayrollRecords_UnknownOwnerReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, err := repo.ListPayrollRecords(ctx, 0, "")
	if err != nil {
		t.Fatalf("ListPayrollRecords error: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected empty list for owner_id=0, got %d records", len(records))
	}
}

func TestListPayrollRecords_EachHasNonZeroID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("ListPayrollRecords error: %v", err)
	}
	for i, r := range records {
		if r.ID == 0 {
			t.Errorf("record[%d] has zero ID", i)
		}
	}
}

func TestListPayrollRecords_OwnerIDMatches(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("ListPayrollRecords error: %v", err)
	}
	for i, r := range records {
		if r.OwnerID != testutil.OwnerID {
			t.Errorf("record[%d] has OwnerID=%d, want %d", i, r.OwnerID, testutil.OwnerID)
		}
	}
}

func TestListPayrollRecords_FilterByPeriodLabel(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// An arbitrary label unlikely to match any fixture row.
	records, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "2099-99")
	if err != nil {
		t.Fatalf("ListPayrollRecords error: %v", err)
	}
	// No fixture rows with that label; result must be empty or nil.
	if len(records) != 0 {
		t.Errorf("expected 0 records for non-existent period label, got %d", len(records))
	}
}

func TestCreatePayrollRecord_InsertsRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	periodStart := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)

	id, err := repo.CreatePayrollRecord(
		ctx,
		testutil.OwnerID,
		4101, // Alice from hr_staff fixture
		periodStart, periodEnd,
		"2026-06",
		80, 0, 3200, 0, 3200,
	)
	if err != nil {
		t.Fatalf("CreatePayrollRecord error: %v", err)
	}
	if id == 0 {
		t.Fatal("expected non-zero ID")
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE id = ?", id)
	})

	records, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "2026-06")
	if err != nil {
		t.Fatalf("ListPayrollRecords after create: %v", err)
	}

	found := false
	for _, r := range records {
		if r.ID == id {
			found = true
		}
	}
	if !found {
		t.Errorf("inserted record id=%d not found in list results", id)
	}
}

func TestMarkPayrollPaid_UnknownIDErrors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.MarkPayrollPaid(ctx, testutil.OwnerID, 0)
	if err == nil {
		t.Fatal("expected error for unknown payroll ID, got nil")
	}
}

func TestGetCompletedShiftsForPeriod_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	_, err := repo.GetCompletedShiftsForPeriod(ctx, testutil.OwnerID, start, end)
	if err != nil {
		t.Fatalf("GetCompletedShiftsForPeriod error: %v", err)
	}
}

func TestGetStaffPayRates_UnknownStaffErrors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetStaffPayRates(ctx, testutil.OwnerID, 0)
	if err == nil {
		t.Fatal("expected error for unknown staff_id=0, got nil")
	}
}

func TestListPayrollRecords_TwiceReturnsSameCount(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	first, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("first ListPayrollRecords: %v", err)
	}
	second, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "")
	if err != nil {
		t.Fatalf("second ListPayrollRecords: %v", err)
	}
	if len(first) != len(second) {
		t.Errorf("count changed between calls: %d vs %d", len(first), len(second))
	}
}

func TestCreatePayrollRecord_BasePayPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantBasePay := 4500.00
	periodStart := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)

	id, err := repo.CreatePayrollRecord(
		ctx,
		testutil.OwnerID, 4102, // Bob
		periodStart, periodEnd, "2026-07",
		160, 0, wantBasePay, 0, wantBasePay,
	)
	if err != nil {
		t.Fatalf("CreatePayrollRecord error: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE id = ?", id)
	})

	records, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "2026-07")
	if err != nil {
		t.Fatalf("ListPayrollRecords: %v", err)
	}
	for _, r := range records {
		if r.ID == id {
			if r.BasePay != wantBasePay {
				t.Errorf("BasePay mismatch: got %g, want %g", r.BasePay, wantBasePay)
			}
			return
		}
	}
	t.Errorf("inserted record id=%d not found", id)
}
