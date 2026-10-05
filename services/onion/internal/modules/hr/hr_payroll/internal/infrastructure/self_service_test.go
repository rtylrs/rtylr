package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestGetPayrollEntryForEmployee_ReturnsAlicesOwnEntry(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entry, err := repo.GetPayrollEntryForEmployee(ctx, testutil.BusinessID, 4301, 4101)
	if err != nil {
		t.Fatalf("GetPayrollEntryForEmployee: %v", err)
	}
	if entry.EmployeeID != 4101 {
		t.Errorf("EmployeeID = %d, want 4101", entry.EmployeeID)
	}
	if entry.NetPayCents != 110000 {
		t.Errorf("NetPayCents = %v, want 110000", entry.NetPayCents)
	}
}

func TestGetPayrollEntryForEmployee_DoesNotReturnAnotherEmployeesEntry(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	entry, err := repo.GetPayrollEntryForEmployee(ctx, testutil.BusinessID, 4301, 4102)
	if err != nil {
		t.Fatalf("GetPayrollEntryForEmployee: %v", err)
	}
	if entry.EmployeeID != 4102 {
		t.Errorf("EmployeeID = %d, want 4102", entry.EmployeeID)
	}
	if entry.NetPayCents != 60000 {
		t.Errorf("NetPayCents = %v, want 60000", entry.NetPayCents)
	}
	if entry.NetPayCents == 110000 {
		t.Error("got Alice's net pay instead of Bob's: entry lookup is not scoped to employee_id")
	}
}

func TestGetPayrollEntryForEmployee_EmployeeWithNoEntryInRun_ReturnsNotFound(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetPayrollEntryForEmployee(ctx, testutil.OwnerID, 4301, 999999999)
	if !errors.Is(err, domain.ErrPayrollRunNotFound) {
		t.Errorf("expected ErrPayrollRunNotFound, got %v", err)
	}
}

func TestListPayrollRunsForEmployee_IncludesRunWhereEmployeeHasEntry(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	runs, total, err := repo.ListPayrollRunsForEmployee(ctx, testutil.BusinessID, 4101, domain.ListPayrollRunsParams{Limit: 100})
	if err != nil {
		t.Fatalf("ListPayrollRunsForEmployee: %v", err)
	}
	if total < 1 {
		t.Fatalf("expected at least 1 run for employee 4101, got total=%d", total)
	}
	found := false
	for _, r := range runs {
		if r.ID == 4301 {
			found = true
		}
	}
	if !found {
		t.Error("expected run 4301 in employee 4101's self-service run list")
	}
}

func TestListPayrollRunsForEmployee_ExcludesRunWhereEmployeeHasNoEntry(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	run, _, err := repo.CreatePayrollRunWithEntries(ctx, testutil.OwnerID, domain.PayrollRunRecord{
		PeriodStart: "2026-12-01",
		PeriodEnd:   "2026-12-31",
		Status:      domain.PayrollStatusApproved,
	}, []domain.PayrollEntryRecord{
		{EmployeeID: 4102, NetPayCents: 50000},
	})
	if err != nil {
		t.Fatalf("CreatePayrollRunWithEntries: %v", err)
	}

	runs, _, err := repo.ListPayrollRunsForEmployee(ctx, testutil.OwnerID, 4101, domain.ListPayrollRunsParams{Limit: 200})
	if err != nil {
		t.Fatalf("ListPayrollRunsForEmployee: %v", err)
	}
	for _, r := range runs {
		if r.ID == run.ID {
			t.Errorf("employee 4101 should not see run %d — they have no entry in it", run.ID)
		}
	}
}

func TestListPayrollRunsForEmployee_ExcludesDraftRun(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	run, _, err := repo.CreatePayrollRunWithEntries(ctx, testutil.OwnerID, domain.PayrollRunRecord{
		PeriodStart: "2027-01-01",
		PeriodEnd:   "2027-01-31",
		Status:      domain.PayrollStatusDraft,
	}, []domain.PayrollEntryRecord{
		{EmployeeID: 4101, NetPayCents: 70000},
	})
	if err != nil {
		t.Fatalf("CreatePayrollRunWithEntries: %v", err)
	}

	runs, _, err := repo.ListPayrollRunsForEmployee(ctx, testutil.OwnerID, 4101, domain.ListPayrollRunsParams{Limit: 200})
	if err != nil {
		t.Fatalf("ListPayrollRunsForEmployee: %v", err)
	}
	for _, r := range runs {
		if r.ID == run.ID {
			t.Errorf("draft run %d should not be visible in self-service list", run.ID)
		}
	}
}
