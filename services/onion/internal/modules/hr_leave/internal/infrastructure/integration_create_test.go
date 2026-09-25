package infrastructure_test

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestCreateLeaveRequest_EmployeeIDIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "EmpID Check Leave",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_requests WHERE leave_type_id = ?", lt.ID)
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "EmpID Check Leave")
	})

	req, err := repo.CreateLeaveRequest(ctx, testutil.OwnerID, domain.LeaveRequestRecord{
		EmployeeID:           4102,
		LeaveTypeID:          lt.ID,
		StartDate:            "2026-09-01",
		EndDate:              "2026-09-02",
		TotalDays:            2,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateLeaveRequest: %v", err)
	}
	if req.EmployeeID != 4102 {
		t.Errorf("expected employee_id=4102, got %d", req.EmployeeID)
	}
}

func TestCreateLeaveRequest_StatusIsPending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "Status Check Leave",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_requests WHERE leave_type_id = ?", lt.ID)
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Status Check Leave")
	})

	req, err := repo.CreateLeaveRequest(ctx, testutil.OwnerID, domain.LeaveRequestRecord{
		EmployeeID:           4101,
		LeaveTypeID:          lt.ID,
		StartDate:            "2026-10-01",
		EndDate:              "2026-10-01",
		TotalDays:            1,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateLeaveRequest: %v", err)
	}
	if req.Status != "pending" {
		t.Errorf("expected status='pending', got %q", req.Status)
	}
}

func TestListLeaveRequests_ByEmployee_4101(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// List all and verify any returned records for OwnerID have valid fields.
	requests, _, err := repo.ListLeaveRequests(ctx, testutil.OwnerID, domain.ListLeaveRequestsParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListLeaveRequests: %v", err)
	}
	for _, r := range requests {
		if r.OwnerID != testutil.OwnerID {
			t.Errorf("request %d has wrong owner_id %d", r.ID, r.OwnerID)
		}
	}
}

func TestApproveLeaveRequest_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "Approve Test Leave",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_requests WHERE leave_type_id = ?", lt.ID)
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Approve Test Leave")
	})

	req, err := repo.CreateLeaveRequest(ctx, testutil.OwnerID, domain.LeaveRequestRecord{
		EmployeeID:           4101,
		LeaveTypeID:          lt.ID,
		StartDate:            "2026-11-01",
		EndDate:              "2026-11-01",
		TotalDays:            1,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateLeaveRequest: %v", err)
	}

	if err := repo.ApproveLeaveRequest(ctx, testutil.OwnerID, req.ID); err != nil {
		t.Fatalf("ApproveLeaveRequest: %v", err)
	}

	got, err := repo.GetLeaveRequest(ctx, testutil.OwnerID, req.ID)
	if err != nil {
		t.Fatalf("GetLeaveRequest after approve: %v", err)
	}
	if got.Status != "approved" {
		t.Errorf("expected status='approved', got %q", got.Status)
	}

	if err := repo.ApproveLeaveRequest(ctx, testutil.OwnerID, req.ID); !errors.Is(err, domain.ErrLeaveRequestNotPending) {
		t.Fatalf("second approval error = %v, want %v", err, domain.ErrLeaveRequestNotPending)
	}
}

func TestApproveLeaveRequestWithBalance_IsAtomicAndIdempotent(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	const employeeID int64 = 4103
	const year int32 = 2027

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "Atomic Approve Test Leave",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_requests WHERE owner_id = ? AND employee_id = ?", testutil.OwnerID, employeeID)
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_balances WHERE owner_id = ? AND employee_id = ?", testutil.OwnerID, employeeID)
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Atomic Approve Test Leave")
	})

	if err := repo.UpsertLeaveBalance(ctx, testutil.OwnerID, domain.LeaveBalanceRecord{
		EmployeeID:   employeeID,
		LeaveTypeID:  lt.ID,
		Year:         year,
		EntitledDays: 10,
		PendingDays:  2,
		UsedDays:     1,
	}); err != nil {
		t.Fatalf("UpsertLeaveBalance: %v", err)
	}

	req, err := repo.CreateLeaveRequest(ctx, testutil.OwnerID, domain.LeaveRequestRecord{
		EmployeeID:           employeeID,
		LeaveTypeID:          lt.ID,
		StartDate:            "2027-01-10",
		EndDate:              "2027-01-11",
		TotalDays:            2,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateLeaveRequest: %v", err)
	}

	if err := repo.ApproveLeaveRequestWithBalance(ctx, testutil.OwnerID, req.ID, employeeID, lt.ID, year, 2); err != nil {
		t.Fatalf("ApproveLeaveRequestWithBalance: %v", err)
	}

	got, err := repo.GetLeaveBalance(ctx, testutil.OwnerID, employeeID, lt.ID, year)
	if err != nil {
		t.Fatalf("GetLeaveBalance after approve: %v", err)
	}
	if got.UsedDays != 3 || got.PendingDays != 0 {
		t.Fatalf("balance after approve = used %.2f pending %.2f, want used 3 pending 0", got.UsedDays, got.PendingDays)
	}

	if err := repo.ApproveLeaveRequestWithBalance(ctx, testutil.OwnerID, req.ID, employeeID, lt.ID, year, 2); !errors.Is(err, domain.ErrLeaveRequestNotPending) {
		t.Fatalf("second approval error = %v, want %v", err, domain.ErrLeaveRequestNotPending)
	}

	got, err = repo.GetLeaveBalance(ctx, testutil.OwnerID, employeeID, lt.ID, year)
	if err != nil {
		t.Fatalf("GetLeaveBalance after duplicate approve: %v", err)
	}
	if got.UsedDays != 3 || got.PendingDays != 0 {
		t.Fatalf("balance after duplicate approve = used %.2f pending %.2f, want unchanged 3/0", got.UsedDays, got.PendingDays)
	}
}
