package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreatePayrollRun_RejectsDuplicateBusinessPeriod(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const businessID int64 = 9100103
	const periodStart = "2026-11-01"
	const periodEnd = "2026-11-30"

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_payroll_runs WHERE business_id = ? AND period_start = ? AND period_end = ?", businessID, periodStart, periodEnd)
	})

	input := domain.PayrollRunRecord{PeriodStart: periodStart, PeriodEnd: periodEnd, Status: domain.PayrollStatusDraft}
	if _, _, err := repo.CreatePayrollRunWithEntries(ctx, businessID, input, nil); err != nil {
		t.Fatalf("first CreatePayrollRunWithEntries: %v", err)
	}
	if _, _, err := repo.CreatePayrollRunWithEntries(ctx, businessID, input, nil); !errors.Is(err, domain.ErrPayrollRunAlreadyExists) {
		t.Fatalf("duplicate CreatePayrollRunWithEntries error = %v, want ErrPayrollRunAlreadyExists", err)
	}
}

func TestListPayrollRuns_ReturnsMarch2026Approved(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	runs, total, err := repo.ListPayrollRuns(ctx, testutil.OwnerID, domain.ListPayrollRunsParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListPayrollRuns: %v", err)
	}
	if total < 1 {
		t.Fatalf("expected at least 1 payroll run, got total=%d", total)
	}

	found := false
	for _, r := range runs {
		if r.ID == 4301 {
			found = true
			if r.Status != "approved" {
				t.Errorf("expected run 4301 status=approved, got %q", r.Status)
			}
			if r.PeriodStart != "2026-03-01" {
				t.Errorf("expected period_start=2026-03-01, got %q", r.PeriodStart)
			}
			if r.PeriodEnd != "2026-03-31" {
				t.Errorf("expected period_end=2026-03-31, got %q", r.PeriodEnd)
			}
			break
		}
	}
	if !found {
		t.Error("payroll run id=4301 not found in list")
	}
}

func TestListPayrollEntries_Run4301HasTwoEntries(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, err := repo.ListPayrollEntries(ctx, testutil.OwnerID, 4301)
	if err != nil {
		t.Fatalf("ListPayrollEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries for run 4301, got %d", len(entries))
	}

	netPays := make(map[float64]bool)
	for _, e := range entries {
		netPays[e.NetPay] = true
	}
	if !netPays[1100.0] {
		t.Errorf("expected entry with net_pay=1100, got %v", netPays)
	}
	if !netPays[600.0] {
		t.Errorf("expected entry with net_pay=600, got %v", netPays)
	}
}

func TestGetPayrollConfig_MonthlyPeriod(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetPayrollConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPayrollConfig: %v", err)
	}
	if cfg.PeriodType != "monthly" {
		t.Errorf("expected period_type=monthly, got %q", cfg.PeriodType)
	}
	if cfg.PayDayOfMonth != 28 {
		t.Errorf("expected pay_day_of_month=28, got %d", cfg.PayDayOfMonth)
	}
}

func TestListPayrollRuns_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	runs, total, err := repo.ListPayrollRuns(ctx, 0, domain.ListPayrollRunsParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListPayrollRuns zero owner: %v", err)
	}
	if total != 0 || len(runs) != 0 {
		t.Errorf("expected empty for owner 0, got total=%d records=%d", total, len(runs))
	}
}

func TestListPayrollRuns_ContainsRun4301(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	runs, _, err := repo.ListPayrollRuns(ctx, testutil.OwnerID, domain.ListPayrollRunsParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListPayrollRuns: %v", err)
	}
	for _, r := range runs {
		if r.ID == 4301 {
			return
		}
	}
	t.Error("payroll run id=4301 not found")
}

func TestListPayrollRuns_Run4301IsApproved(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	runs, _, err := repo.ListPayrollRuns(ctx, testutil.OwnerID, domain.ListPayrollRunsParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListPayrollRuns: %v", err)
	}
	for _, r := range runs {
		if r.ID == 4301 {
			if r.Status != "approved" {
				t.Errorf("expected run 4301 status=approved, got %q", r.Status)
			}
			return
		}
	}
	t.Error("run 4301 not found")
}

func TestGetPayrollRunByID_4301(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	run, err := repo.GetPayrollRun(ctx, testutil.OwnerID, 4301)
	if err != nil {
		t.Fatalf("GetPayrollRun: %v", err)
	}
	if run.ID != 4301 {
		t.Errorf("expected run id=4301, got %d", run.ID)
	}
	if run.OwnerID != testutil.OwnerID {
		t.Errorf("expected owner_id=%d, got %d", testutil.OwnerID, run.OwnerID)
	}
}

func TestGetPayrollRunByID_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetPayrollRun(ctx, testutil.OwnerID, 999999999)
	if !errors.Is(err, domain.ErrPayrollRunNotFound) {
		t.Errorf("expected ErrPayrollRunNotFound, got %v", err)
	}
}

func TestGetPayrollEntries_ForRun4301_Returns2Entries(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, err := repo.ListPayrollEntries(ctx, testutil.OwnerID, 4301)
	if err != nil {
		t.Fatalf("ListPayrollEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries for run 4301, got %d", len(entries))
	}
}

func TestGetPayrollEntries_AliceNetPay1100(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, err := repo.ListPayrollEntries(ctx, testutil.OwnerID, 4301)
	if err != nil {
		t.Fatalf("ListPayrollEntries: %v", err)
	}
	for _, e := range entries {
		if e.EmployeeID == 4101 {
			if e.NetPay != 1100.0 {
				t.Errorf("expected Alice net_pay=1100, got %v", e.NetPay)
			}
			return
		}
	}
	t.Error("Alice (employee_id=4101) not found in entries for run 4301")
}

func TestGetPayrollEntries_BobNetPay600(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entries, err := repo.ListPayrollEntries(ctx, testutil.OwnerID, 4301)
	if err != nil {
		t.Fatalf("ListPayrollEntries: %v", err)
	}
	for _, e := range entries {
		if e.EmployeeID == 4102 {
			if e.NetPay != 600.0 {
				t.Errorf("expected Bob net_pay=600, got %v", e.NetPay)
			}
			return
		}
	}
	t.Error("Bob (employee_id=4102) not found in entries for run 4301")
}

func TestCreatePayrollRun_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	run, _, err := repo.CreatePayrollRunWithEntries(ctx, testutil.OwnerID, domain.PayrollRunRecord{
		PeriodStart: "2026-06-01",
		PeriodEnd:   "2026-06-30",
		Status:      "draft",
	}, []domain.PayrollEntryRecord{})
	if err != nil {
		t.Fatalf("CreatePayrollRunWithEntries: %v", err)
	}
	if run.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
}

func TestCreatePayrollRun_StatusIsDraft(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	run, _, err := repo.CreatePayrollRunWithEntries(ctx, testutil.OwnerID, domain.PayrollRunRecord{
		PeriodStart: "2026-07-01",
		PeriodEnd:   "2026-07-31",
		Status:      "draft",
	}, []domain.PayrollEntryRecord{})
	if err != nil {
		t.Fatalf("CreatePayrollRunWithEntries: %v", err)
	}
	if run.Status != "draft" {
		t.Errorf("expected status=draft, got %q", run.Status)
	}
}

func TestUpsertPayrollConfig_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	err := repo.UpsertPayrollConfig(ctx, testutil.OwnerID, domain.PayrollConfigRecord{
		PeriodType:    "monthly",
		PayDayOfMonth: 28,
	})
	if err != nil {
		t.Fatalf("UpsertPayrollConfig: %v", err)
	}

	cfg, err := repo.GetPayrollConfig(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("GetPayrollConfig after upsert: %v", err)
	}
	if cfg.OwnerID != testutil.OwnerID {
		t.Errorf("expected owner_id=%d, got %d", testutil.OwnerID, cfg.OwnerID)
	}
}

func TestListPayrollRuns_AfterCreate_CountIncreases(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, total1, err := repo.ListPayrollRuns(ctx, testutil.OwnerID, domain.ListPayrollRunsParams{Limit: 200, Offset: 0})
	if err != nil {
		t.Fatalf("ListPayrollRuns before: %v", err)
	}

	_, _, err = repo.CreatePayrollRunWithEntries(ctx, testutil.OwnerID, domain.PayrollRunRecord{
		PeriodStart: "2026-08-01",
		PeriodEnd:   "2026-08-31",
		Status:      "draft",
	}, []domain.PayrollEntryRecord{})
	if err != nil {
		t.Fatalf("CreatePayrollRunWithEntries: %v", err)
	}

	_, total2, err := repo.ListPayrollRuns(ctx, testutil.OwnerID, domain.ListPayrollRunsParams{Limit: 200, Offset: 0})
	if err != nil {
		t.Fatalf("ListPayrollRuns after: %v", err)
	}
	if total2 <= total1 {
		t.Errorf("expected total to increase from %d, got %d", total1, total2)
	}
}

func TestListDistinctOwnerIDs_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	ids, err := repo.ListDistinctOwnerIDs(ctx)
	if err != nil {
		t.Fatalf("ListDistinctOwnerIDs: %v", err)
	}
	_ = ids
}

func TestUpdatePayrollRunStatus_Changes(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	run, _, err := repo.CreatePayrollRunWithEntries(ctx, testutil.OwnerID, domain.PayrollRunRecord{
		PeriodStart: "2026-09-01",
		PeriodEnd:   "2026-09-30",
		Status:      "draft",
	}, []domain.PayrollEntryRecord{})
	if err != nil {
		t.Fatalf("CreatePayrollRunWithEntries: %v", err)
	}

	if err := repo.UpdatePayrollRunStatus(ctx, testutil.OwnerID, run.ID, "pending_approval"); err != nil {
		t.Fatalf("UpdatePayrollRunStatus: %v", err)
	}

	got, err := repo.GetPayrollRun(ctx, testutil.OwnerID, run.ID)
	if err != nil {
		t.Fatalf("GetPayrollRun after status update: %v", err)
	}
	if got.Status != "pending_approval" {
		t.Errorf("expected status='pending_approval', got %q", got.Status)
	}
}

func TestClaimPayrollRunForExport_AllowsOnlyOneClaim(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	run, _, err := repo.CreatePayrollRunWithEntries(ctx, testutil.OwnerID, domain.PayrollRunRecord{
		PeriodStart: "2026-10-01",
		PeriodEnd:   "2026-10-31",
		Status:      domain.PayrollStatusApproved,
	}, []domain.PayrollEntryRecord{})
	if err != nil {
		t.Fatalf("CreatePayrollRunWithEntries: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_payroll_runs WHERE id = ? AND business_id = ?", run.ID, testutil.OwnerID)
	})

	if err := repo.ClaimPayrollRunForExport(ctx, testutil.OwnerID, run.ID); err != nil {
		t.Fatalf("first ClaimPayrollRunForExport: %v", err)
	}
	if err := repo.ClaimPayrollRunForExport(ctx, testutil.OwnerID, run.ID); !errors.Is(err, domain.ErrPayrollExportInProgress) {
		t.Fatalf("second ClaimPayrollRunForExport error = %v, want %v", err, domain.ErrPayrollExportInProgress)
	}

	if err := repo.ReleasePayrollRunExportClaim(ctx, testutil.OwnerID, run.ID); err != nil {
		t.Fatalf("ReleasePayrollRunExportClaim: %v", err)
	}
	if err := repo.ClaimPayrollRunForExport(ctx, testutil.OwnerID, run.ID); err != nil {
		t.Fatalf("claim after release: %v", err)
	}
}

func TestListEmployeePayInfo_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	info, err := repo.ListEmployeePayInfo(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListEmployeePayInfo: %v", err)
	}
	_ = info
}
