package infrastructure_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr_leave/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

func TestListLeaveRequests_EmployeeAndStatusFilters(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	name := fmt.Sprintf("Scope filter %d", time.Now().UnixNano())

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:             name,
		RequiresApproval: false,
		IsActive:         true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	var requestIDs []int64
	t.Cleanup(func() {
		for _, requestID := range requestIDs {
			_, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_requests WHERE id = ?", requestID)
		}
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_types WHERE id = ?", lt.ID)
	})

	for _, request := range []domain.LeaveRequestRecord{
		{EmployeeID: 4101, LeaveTypeID: lt.ID, StartDate: "2035-01-01", EndDate: "2035-01-02", TotalDays: 2, Status: domain.LeaveStatusPending},
		{EmployeeID: 4102, LeaveTypeID: lt.ID, StartDate: "2035-02-01", EndDate: "2035-02-02", TotalDays: 2, Status: domain.LeaveStatusApproved},
	} {
		created, createErr := repo.CreateLeaveRequest(ctx, testutil.OwnerID, request)
		if createErr != nil {
			t.Fatalf("CreateLeaveRequest: %v", createErr)
		}
		requestIDs = append(requestIDs, created.ID)
	}

	requests, total, err := repo.ListLeaveRequests(ctx, testutil.OwnerID, domain.ListLeaveRequestsParams{
		EmployeeID: 4101,
		Status:     domain.LeaveStatusPending,
		Limit:      100,
	})
	if err != nil {
		t.Fatalf("ListLeaveRequests: %v", err)
	}
	if total != 1 || len(requests) != 1 {
		t.Fatalf("filtered requests total=%d len=%d, want 1", total, len(requests))
	}
	if requests[0].EmployeeID != 4101 || requests[0].Status != domain.LeaveStatusPending {
		t.Fatalf("filtered request = employee %d status %q, want employee 4101 pending", requests[0].EmployeeID, requests[0].Status)
	}
}

func TestListLeaveRequests_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	requests, _, err := repo.ListLeaveRequests(ctx, testutil.OwnerID, domain.ListLeaveRequestsParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListLeaveRequests: %v", err)
	}
	// Result may be empty -- fixture has no leave requests; just assert no error.
	_ = requests
}

func TestListLeaveTypes_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	types, _, err := repo.ListLeaveTypes(ctx, testutil.OwnerID, domain.ListLeaveTypesParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListLeaveTypes: %v", err)
	}
	_ = types
}

func TestListAllActiveLeaveTypes_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	types, err := repo.ListAllActiveLeaveTypesForOwner(ctx, testutil.OwnerID)
	if err != nil {
		t.Fatalf("ListAllActiveLeaveTypesForOwner: %v", err)
	}
	_ = types
}

func TestListLeaveRequests_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	requests, total, err := repo.ListLeaveRequests(ctx, 0, domain.ListLeaveRequestsParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListLeaveRequests zero owner: %v", err)
	}
	if total != 0 || len(requests) != 0 {
		t.Errorf("expected empty for owner 0, got total=%d records=%d", total, len(requests))
	}
}

func TestListLeaveTypes_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	types, total, err := repo.ListLeaveTypes(ctx, 0, domain.ListLeaveTypesParams{
		Limit:  100,
		Offset: 0,
	})
	if err != nil {
		t.Fatalf("ListLeaveTypes zero owner: %v", err)
	}
	if total != 0 || len(types) != 0 {
		t.Errorf("expected empty for owner 0, got total=%d records=%d", total, len(types))
	}
}

func TestCreateLeaveType_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:             "Test Annual Leave",
		IsPaid:           true,
		RequiresApproval: true,
		IsActive:         true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Test Annual Leave")
	})
	if lt.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
}

func TestCreateLeaveType_NameIsStored(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "Sick Leave Test",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Sick Leave Test")
	})
	if lt.Name != "Sick Leave Test" {
		t.Errorf("expected name='Sick Leave Test', got %q", lt.Name)
	}
}

func TestCreateLeaveRequest_PersistsAndReturnsID(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	// Create a leave type first.
	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "Leave For Request Test",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_requests WHERE leave_type_id = ?", lt.ID)
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Leave For Request Test")
	})

	req, err := repo.CreateLeaveRequest(ctx, testutil.OwnerID, domain.LeaveRequestRecord{
		EmployeeID:           4101,
		LeaveTypeID:          lt.ID,
		StartDate:            "2026-08-01",
		EndDate:              "2026-08-03",
		TotalDays:            3,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateLeaveRequest: %v", err)
	}
	if req.ID == 0 {
		t.Fatal("expected non-zero ID after create")
	}
}
