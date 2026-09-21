package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// cancelingFinancePoster cancels the request context before failing, standing
// in for a client disconnect or request deadline during the finance post.
type cancelingFinancePoster struct {
	cancel context.CancelFunc
	err    error
}

func (p cancelingFinancePoster) PostPayrollJournal(context.Context, domain.PayrollJournalEntry) error {
	p.cancel()
	return p.err
}

// A payroll run claimed for export must return to 'approved' even when the
// request context dies mid-export: the claim's only reverse transition is the
// compensating release, and a canceled context used to make that release fail
// silently, leaving the run stuck in 'exporting' forever (B62-R05).
func TestExportPayrollToFinance_ReleasesClaimWhenRequestContextIsCanceled(t *testing.T) {
	db := testutil.OpenDB(t)
	ctx := context.Background()
	const businessID int64 = 9100625
	const periodStart, periodEnd = "2026-12-01", "2026-12-31"

	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM hr_payroll_entries WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM hr_payroll_runs WHERE business_id = ?", businessID)
		_, _ = db.ExecContext(cleanupCtx, "DELETE FROM hr_employees WHERE business_id = ?", businessID)
	})

	empRes, err := db.ExecContext(ctx, `
		INSERT INTO hr_employees (owner_id, business_id, first_name, last_name, email, status)
		VALUES (?, ?, 'Export', 'Release', 'payroll-export-release@example.com', 'active')`,
		businessID, businessID)
	if err != nil {
		t.Fatalf("insert employee: %v", err)
	}
	employeeID, err := empRes.LastInsertId()
	if err != nil {
		t.Fatalf("employee id: %v", err)
	}

	runRes, err := db.ExecContext(ctx, `
		INSERT INTO hr_payroll_runs (owner_id, business_id, period_start, period_end, status, created_by)
		VALUES (?, ?, ?, ?, 'approved', ?)`,
		businessID, businessID, periodStart, periodEnd, businessID)
	if err != nil {
		t.Fatalf("insert payroll run: %v", err)
	}
	runID, err := runRes.LastInsertId()
	if err != nil {
		t.Fatalf("payroll run id: %v", err)
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO hr_payroll_entries (owner_id, business_id, payroll_run_id, employee_id, net_pay)
		VALUES (?, ?, ?, ?, 100.00)`,
		businessID, businessID, runID, employeeID); err != nil {
		t.Fatalf("insert payroll entry: %v", err)
	}

	reqCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	wantErr := errors.New("finance post interrupted")
	svc := &Service{
		repo:          infrastructure.NewRepository(db),
		financePoster: cancelingFinancePoster{cancel: cancel, err: wantErr},
		notifier:      noopNotifier{},
	}

	if err := svc.ExportPayrollToFinance(reqCtx, businessID, runID); !errors.Is(err, wantErr) {
		t.Fatalf("ExportPayrollToFinance error = %v, want %v", err, wantErr)
	}

	run, err := svc.repo.GetPayrollRun(context.Background(), businessID, runID)
	if err != nil {
		t.Fatalf("GetPayrollRun after export failure: %v", err)
	}
	if run.Status != domain.PayrollStatusApproved {
		t.Fatalf("run status = %q, want %q (export claim was never released)",
			run.Status, domain.PayrollStatusApproved)
	}
	if !run.ExportedAt.IsZero() {
		t.Fatalf("exported_at = %v, want zero after a failed export", run.ExportedAt)
	}
}
