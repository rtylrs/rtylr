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

func TestGetProjectLabourCostReportsUncostedBillableMinutes(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const businessID int64 = 9100686
	const projectID int64 = 9900686

	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_time_entries WHERE business_id = ? AND project_id = ?", businessID, projectID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_employees WHERE business_id = ?", businessID)
	})

	res, err := db.ExecContext(ctx, `
		INSERT INTO hr_employees (owner_id, business_id, first_name, last_name, email, status, hourly_rate)
		VALUES (?, ?, ?, ?, ?, 'active', ?), (?, ?, ?, ?, ?, 'active', NULL)`,
		businessID, businessID, "Costed", "Employee", "labour-cost-costed@example.com", "30.00",
		businessID, businessID, "Uncosted", "Employee", "labour-cost-uncosted@example.com")
	if err != nil {
		t.Fatalf("insert labour cost employees: %v", err)
	}
	firstEmployeeID, err := res.LastInsertId()
	if err != nil {
		t.Fatalf("first employee id: %v", err)
	}
	secondEmployeeID := firstEmployeeID + 1

	_, err = db.ExecContext(ctx, `
		INSERT INTO hr_time_entries
		(owner_id, business_id, employee_id, project_id, date, duration_minutes, is_billable, status)
		VALUES (?, ?, ?, ?, '2026-08-01', 60, TRUE, 'approved'),
		       (?, ?, ?, ?, '2026-08-01', 30, TRUE, 'approved'),
		       (?, ?, ?, ?, '2026-08-01', 90, FALSE, 'approved')`,
		businessID, businessID, firstEmployeeID, projectID,
		businessID, businessID, secondEmployeeID, projectID,
		businessID, businessID, firstEmployeeID, projectID)
	if err != nil {
		t.Fatalf("insert labour cost entries: %v", err)
	}

	cost, err := repo.GetProjectLabourCost(ctx, businessID, projectID)
	if err != nil {
		t.Fatalf("GetProjectLabourCost: %v", err)
	}
	if cost.BusinessID != businessID || cost.ProjectID != projectID {
		t.Fatalf("scope = business %d, project %d; want %d, %d", cost.BusinessID, cost.ProjectID, businessID, projectID)
	}
	if cost.CostCents != 3000 || cost.CostedMinutes != 60 || cost.UncostedMinutes != 30 || cost.UncostedEmployeeCount != 1 {
		t.Fatalf("aggregate = %+v; want 3000 cents, 60 costed minutes, 30 uncosted minutes, 1 uncosted employee", cost)
	}
}

func TestGetProjectLabourCostReturnsZeroForProjectWithoutBillableTime(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const businessID int64 = 9100687
	const projectID int64 = 9900687

	cost, err := repo.GetProjectLabourCost(ctx, businessID, projectID)
	if err != nil {
		t.Fatalf("GetProjectLabourCost on a project with no time entries: %v", err)
	}
	if cost.BusinessID != businessID || cost.ProjectID != projectID {
		t.Fatalf("scope = business %d, project %d; want %d, %d", cost.BusinessID, cost.ProjectID, businessID, projectID)
	}
	if cost.CostCents != 0 || cost.CostedMinutes != 0 || cost.UncostedMinutes != 0 || cost.UncostedEmployeeCount != 0 {
		t.Fatalf("aggregate = %+v; want all zero", cost)
	}
}

func TestListPayrollRuns_ReturnsMarch2026Approved(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	runs, total, err := repo.ListPayrollRuns(ctx, testutil.BusinessID, domain.ListPayrollRunsParams{
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

	entries, err := repo.ListPayrollEntries(ctx, testutil.BusinessID, 4301)
	if err != nil {
		t.Fatalf("ListPayrollEntries: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries for run 4301, got %d", len(entries))
	}

	netPays := make(map[int64]bool)
	for _, e := range entries {
		netPays[e.NetPayCents] = true
	}
	if !netPays[110000] {
		t.Errorf("expected entry with net_pay_cents=110000, got %v", netPays)
	}
	if !netPays[60000] {
		t.Errorf("expected entry with net_pay_cents=60000, got %v", netPays)
	}
}

func TestGetPayrollConfig_MonthlyPeriod(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetPayrollConfig(ctx, testutil.BusinessID)
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

	runs, _, err := repo.ListPayrollRuns(ctx, testutil.BusinessID, domain.ListPayrollRunsParams{
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

	runs, _, err := repo.ListPayrollRuns(ctx, testutil.BusinessID, domain.ListPayrollRunsParams{
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

	run, err := repo.GetPayrollRun(ctx, testutil.BusinessID, 4301)
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

// ExportPayrollToFinance posts the re-read run's BusinessID as the finance
// journal's tenant scope, and the journal rejects business_id 0. The read-back
// must carry the column for a run created outside the fixture tenant too.
func TestGetPayrollRun_CarriesBusinessIDForFinanceExport(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const businessID int64 = 9100107
	const periodStart = "2027-03-01"
	const periodEnd = "2027-03-31"

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(),
			"DELETE FROM hr_payroll_runs WHERE business_id = ? AND period_start = ? AND period_end = ?",
			businessID, periodStart, periodEnd)
	})

	created, _, err := repo.CreatePayrollRunWithEntries(ctx, businessID, domain.PayrollRunRecord{
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Status:      domain.PayrollStatusApproved,
	}, []domain.PayrollEntryRecord{})
	if err != nil {
		t.Fatalf("CreatePayrollRunWithEntries: %v", err)
	}

	run, err := repo.GetPayrollRun(ctx, businessID, created.ID)
	if err != nil {
		t.Fatalf("GetPayrollRun: %v", err)
	}
	if run.BusinessID != businessID {
		t.Errorf("BusinessID = %d, want %d; the finance journal poster rejects business_id 0", run.BusinessID, businessID)
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

	entries, err := repo.ListPayrollEntries(ctx, testutil.BusinessID, 4301)
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

	entries, err := repo.ListPayrollEntries(ctx, testutil.BusinessID, 4301)
	if err != nil {
		t.Fatalf("ListPayrollEntries: %v", err)
	}
	for _, e := range entries {
		if e.EmployeeID == 4101 {
			if e.NetPayCents != 110000 {
				t.Errorf("expected Alice net_pay_cents=110000, got %v", e.NetPayCents)
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

	entries, err := repo.ListPayrollEntries(ctx, testutil.BusinessID, 4301)
	if err != nil {
		t.Fatalf("ListPayrollEntries: %v", err)
	}
	for _, e := range entries {
		if e.EmployeeID == 4102 {
			if e.NetPayCents != 60000 {
				t.Errorf("expected Bob net_pay_cents=60000, got %v", e.NetPayCents)
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

	// hr_payroll_config has a UNIQUE KEY on owner_id, so this upsert matches the
	// shared 30_hr.sql fixture row (owner_id=testutil.OwnerID) and updates its
	// period fields in place — but ON DUPLICATE KEY UPDATE never touches
	// business_id, so the row's business_id stays testutil.BusinessID (the
	// fixture's original value) and the read-back must filter on that.
	cfg, err := repo.GetPayrollConfig(ctx, testutil.BusinessID)
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

func TestSetPayrollRunApproved_ChangesInReview(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const (
		businessID int64 = 9100203
		approverID int64 = 9300203
	)

	run, _, err := repo.CreatePayrollRunWithEntries(ctx, businessID, domain.PayrollRunRecord{
		PeriodStart: "2031-05-01",
		PeriodEnd:   "2031-05-31",
		Status:      domain.PayrollStatusInReview,
	}, []domain.PayrollEntryRecord{})
	if err != nil {
		t.Fatalf("CreatePayrollRunWithEntries: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "DELETE FROM hr_payroll_runs WHERE id = ? AND business_id = ?", run.ID, businessID)
	})

	if err := repo.SetPayrollRunApproved(ctx, businessID, run.ID, approverID); err != nil {
		t.Fatalf("SetPayrollRunApproved: %v", err)
	}

	got, err := repo.GetPayrollRun(ctx, businessID, run.ID)
	if err != nil {
		t.Fatalf("GetPayrollRun after approval: %v", err)
	}
	if got.Status != domain.PayrollStatusApproved {
		t.Errorf("approved status = %q, want %q", got.Status, domain.PayrollStatusApproved)
	}
	if got.ApprovedBy != approverID {
		t.Errorf("approved_by = %d, want %d", got.ApprovedBy, approverID)
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
