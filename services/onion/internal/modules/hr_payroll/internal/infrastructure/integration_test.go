package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_payroll/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

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
