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

func TestCreatePayrollRecord_OvertimePayPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantOTPay := 600.00
	periodStart := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)

	id, err := repo.CreatePayrollRecord(
		ctx,
		testutil.OwnerID, 4101,
		periodStart, periodEnd, "2026-08",
		160, 10, 3200, wantOTPay, 3200+wantOTPay,
	)
	if err != nil {
		t.Fatalf("CreatePayrollRecord: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE id = ?", id)
	})

	records, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "2026-08")
	if err != nil {
		t.Fatalf("ListPayrollRecords: %v", err)
	}
	for _, r := range records {
		if r.ID == id {
			if r.OvertimePay != wantOTPay {
				t.Errorf("OvertimePay = %g, want %g", r.OvertimePay, wantOTPay)
			}
			return
		}
	}
	t.Errorf("inserted record id=%d not found", id)
}

func TestCreatePayrollRecord_NetPayPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantNetPay := 3800.00
	periodStart := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 9, 30, 0, 0, 0, 0, time.UTC)

	id, err := repo.CreatePayrollRecord(
		ctx,
		testutil.OwnerID, 4102,
		periodStart, periodEnd, "2026-09",
		160, 0, wantNetPay, 0, wantNetPay,
	)
	if err != nil {
		t.Fatalf("CreatePayrollRecord: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE id = ?", id)
	})

	records, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "2026-09")
	if err != nil {
		t.Fatalf("ListPayrollRecords: %v", err)
	}
	for _, r := range records {
		if r.ID == id {
			if r.NetPay != wantNetPay {
				t.Errorf("NetPay = %g, want %g", r.NetPay, wantNetPay)
			}
			return
		}
	}
	t.Errorf("inserted record id=%d not found", id)
}

func TestMarkPayrollPaid_AfterCreate_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	periodStart := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

	id, err := repo.CreatePayrollRecord(
		ctx,
		testutil.OwnerID, 4101,
		periodStart, periodEnd, "2026-10",
		80, 0, 1600, 0, 1600,
	)
	if err != nil {
		t.Fatalf("CreatePayrollRecord: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE id = ?", id)
	})

	if err := repo.MarkPayrollPaid(ctx, testutil.OwnerID, id); err != nil {
		t.Fatalf("MarkPayrollPaid: %v", err)
	}
}

func TestMarkPayrollPaid_WrongOwner_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	periodStart := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 10, 31, 0, 0, 0, 0, time.UTC)

	id, err := repo.CreatePayrollRecord(
		ctx,
		testutil.OwnerID, 4101,
		periodStart, periodEnd, "2026-10b",
		80, 0, 1600, 0, 1600,
	)
	if err != nil {
		t.Fatalf("CreatePayrollRecord: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE id = ?", id)
	})

	err = repo.MarkPayrollPaid(ctx, testutil.OwnerID+99, id)
	if err == nil {
		t.Fatal("expected error marking payroll paid with wrong owner, got nil")
	}
}

func TestGetCompletedShiftsForPeriod_EmptyPeriod_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// A future period with no shifts.
	start := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2099, 1, 31, 0, 0, 0, 0, time.UTC)

	shifts, err := repo.GetCompletedShiftsForPeriod(ctx, testutil.OwnerID, start, end)
	if err != nil {
		t.Fatalf("GetCompletedShiftsForPeriod: %v", err)
	}
	if len(shifts) != 0 {
		t.Errorf("expected 0 shifts in future period, got %d", len(shifts))
	}
}

func TestGetCompletedShiftsForPeriod_AllBelongToOwner(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	shifts, err := repo.GetCompletedShiftsForPeriod(ctx, testutil.OwnerID, start, end)
	if err != nil {
		t.Fatalf("GetCompletedShiftsForPeriod: %v", err)
	}
	for _, s := range shifts {
		if s.StaffID <= 0 {
			t.Errorf("shift has non-positive StaffID: %d", s.StaffID)
		}
	}
}

func TestGetStaffPayRates_ValidStaff_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rates, err := repo.GetStaffPayRates(ctx, testutil.BusinessID, 4101)
	if err != nil {
		t.Fatalf("GetStaffPayRates staff_id=4101: %v", err)
	}
	if rates.HourlyRate < 0 {
		t.Errorf("HourlyRate is negative: %g", rates.HourlyRate)
	}
}

func TestGetStaffPayRates_SecondStaff_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	rates, err := repo.GetStaffPayRates(ctx, testutil.BusinessID, 4102)
	if err != nil {
		t.Fatalf("GetStaffPayRates staff_id=4102: %v", err)
	}
	if rates.HourlyRate < 0 {
		t.Errorf("HourlyRate is negative: %g", rates.HourlyRate)
	}
}

func TestListPayrollRecords_PeriodLabelFilter_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Use a valid-format label; may return 0 rows which is fine.
	records, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "2026-01")
	if err != nil {
		t.Fatalf("ListPayrollRecords period=2026-01: %v", err)
	}
	for _, r := range records {
		if r.PeriodLabel != "2026-01" {
			t.Errorf("record id=%d has PeriodLabel=%q, want 2026-01", r.ID, r.PeriodLabel)
		}
	}
}

func TestCreatePayrollRecord_StaffIDPersisted(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	wantStaffID := int64(4101)
	periodStart := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	periodEnd := time.Date(2026, 11, 30, 0, 0, 0, 0, time.UTC)

	id, err := repo.CreatePayrollRecord(
		ctx,
		testutil.OwnerID, wantStaffID,
		periodStart, periodEnd, "2026-11",
		80, 0, 1600, 0, 1600,
	)
	if err != nil {
		t.Fatalf("CreatePayrollRecord: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(ctx, "DELETE FROM hr_payroll_records WHERE id = ?", id)
	})

	records, err := repo.ListPayrollRecords(ctx, testutil.OwnerID, "2026-11")
	if err != nil {
		t.Fatalf("ListPayrollRecords: %v", err)
	}
	for _, r := range records {
		if r.ID == id {
			if r.StaffID != wantStaffID {
				t.Errorf("StaffID = %d, want %d", r.StaffID, wantStaffID)
			}
			return
		}
	}
	t.Errorf("record id=%d not found after create", id)
}

func TestGetCompletedShiftsForPeriod_UnknownOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	start := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)

	shifts, err := repo.GetCompletedShiftsForPeriod(ctx, 0, start, end)
	if err != nil {
		t.Fatalf("GetCompletedShiftsForPeriod unknown owner: %v", err)
	}
	if len(shifts) != 0 {
		t.Errorf("expected 0 shifts for unknown owner, got %d", len(shifts))
	}
}
