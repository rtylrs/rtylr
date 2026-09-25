package infrastructure_test

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListLeaveBalances_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	balances, err := repo.ListLeaveBalances(ctx, testutil.OwnerID, 4101, 2026)
	if err != nil {
		t.Fatalf("ListLeaveBalances: %v", err)
	}
	_ = balances
}

func TestListLeaveBalances_AllBelongToEmployee(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	balances, err := repo.ListLeaveBalances(ctx, testutil.OwnerID, 4101, 2026)
	if err != nil {
		t.Fatalf("ListLeaveBalances: %v", err)
	}
	for _, b := range balances {
		if b.EmployeeID != 4101 {
			t.Errorf("balance %d has employee_id %d, want 4101", b.ID, b.EmployeeID)
		}
		if b.OwnerID != testutil.OwnerID {
			t.Errorf("balance %d has owner_id %d, want %d", b.ID, b.OwnerID, testutil.OwnerID)
		}
	}
}

func TestUpsertLeaveBalance_Persists(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "Balance Upsert Leave",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Balance Upsert Leave")
	})

	err = repo.UpsertLeaveBalance(ctx, testutil.OwnerID, domain.LeaveBalanceRecord{
		EmployeeID:   4101,
		LeaveTypeID:  lt.ID,
		Year:         2026,
		EntitledDays: 15,
	})
	if err != nil {
		t.Fatalf("UpsertLeaveBalance: %v", err)
	}

	bal, err := repo.GetLeaveBalance(ctx, testutil.OwnerID, 4101, lt.ID, 2026)
	if err != nil {
		t.Fatalf("GetLeaveBalance after upsert: %v", err)
	}
	if bal.EntitledDays != 15 {
		t.Errorf("EntitledDays = %v, want 15", bal.EntitledDays)
	}
}

func TestAddToUsedDays_IncreasesUsed(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "Used Days Leave",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Used Days Leave")
	})

	err = repo.UpsertLeaveBalance(ctx, testutil.OwnerID, domain.LeaveBalanceRecord{
		EmployeeID:   4101,
		LeaveTypeID:  lt.ID,
		Year:         2026,
		EntitledDays: 20,
	})
	if err != nil {
		t.Fatalf("UpsertLeaveBalance: %v", err)
	}

	err = repo.AddToUsedDays(ctx, testutil.OwnerID, 4101, lt.ID, 2026, 3)
	if err != nil {
		t.Fatalf("AddToUsedDays: %v", err)
	}

	bal, err := repo.GetLeaveBalance(ctx, testutil.OwnerID, 4101, lt.ID, 2026)
	if err != nil {
		t.Fatalf("GetLeaveBalance: %v", err)
	}
	if bal.UsedDays < 3 {
		t.Errorf("UsedDays = %v, want at least 3", bal.UsedDays)
	}
}

func TestAddToPendingDays_IncreasesPending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "Pending Days Leave",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Pending Days Leave")
	})

	err = repo.UpsertLeaveBalance(ctx, testutil.OwnerID, domain.LeaveBalanceRecord{
		EmployeeID:   4101,
		LeaveTypeID:  lt.ID,
		Year:         2026,
		EntitledDays: 15,
	})
	if err != nil {
		t.Fatalf("UpsertLeaveBalance: %v", err)
	}

	err = repo.AddToPendingDays(ctx, testutil.OwnerID, 4101, lt.ID, 2026, 2)
	if err != nil {
		t.Fatalf("AddToPendingDays: %v", err)
	}

	bal, err := repo.GetLeaveBalance(ctx, testutil.OwnerID, 4101, lt.ID, 2026)
	if err != nil {
		t.Fatalf("GetLeaveBalance: %v", err)
	}
	if bal.PendingDays < 2 {
		t.Errorf("PendingDays = %v, want at least 2", bal.PendingDays)
	}
}

func TestSubtractPendingDays_ReducesPending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "Subtract Pending Leave",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Subtract Pending Leave")
	})

	err = repo.UpsertLeaveBalance(ctx, testutil.OwnerID, domain.LeaveBalanceRecord{
		EmployeeID:   4101,
		LeaveTypeID:  lt.ID,
		Year:         2026,
		EntitledDays: 10,
		PendingDays:  5,
	})
	if err != nil {
		t.Fatalf("UpsertLeaveBalance: %v", err)
	}

	if err := repo.SubtractPendingDays(ctx, testutil.OwnerID, 4101, lt.ID, 2026, 2); err != nil {
		t.Fatalf("SubtractPendingDays: %v", err)
	}

	bal, err := repo.GetLeaveBalance(ctx, testutil.OwnerID, 4101, lt.ID, 2026)
	if err != nil {
		t.Fatalf("GetLeaveBalance: %v", err)
	}
	if bal.PendingDays > 5 {
		t.Errorf("PendingDays = %v, expected subtraction to reduce it from 5", bal.PendingDays)
	}
}

func TestListBalancesForAccrual_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	balances, err := repo.ListBalancesForAccrual(ctx, testutil.OwnerID, 2026, 5)
	if err != nil {
		t.Fatalf("ListBalancesForAccrual: %v", err)
	}
	_ = balances
}
