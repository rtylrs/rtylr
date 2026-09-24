package infrastructure_test

import (
	"context"
	"errors"
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

func TestRejectLeaveRequest_ChangesStatus(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:     "Reject Test Leave",
		IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() {
		db.ExecContext(context.Background(), "DELETE FROM hr_leave_types WHERE owner_id = ? AND name = ?", testutil.OwnerID, "Reject Test Leave")
	})

	req, err := repo.CreateLeaveRequest(ctx, testutil.OwnerID, domain.LeaveRequestRecord{
		EmployeeID:           4102,
		LeaveTypeID:          lt.ID,
		StartDate:            "2026-12-01",
		EndDate:              "2026-12-01",
		TotalDays:            1,
		Status:               "pending",
		CurrentApprovalLevel: 1,
	})
	if err != nil {
		t.Fatalf("CreateLeaveRequest: %v", err)
	}

	stillPending, err := repo.RejectLeaveRequest(ctx, testutil.OwnerID, req.ID, "not approved")
	if err != nil {
		t.Fatalf("RejectLeaveRequest: %v", err)
	}
	if !stillPending {
		t.Fatal("expected request to still be pending at reject time")
	}

	got, err := repo.GetLeaveRequest(ctx, testutil.OwnerID, req.ID)
	if err != nil {
		t.Fatalf("GetLeaveRequest after reject: %v", err)
	}
	if got.Status != "rejected" {
		t.Errorf("expected status='rejected', got %q", got.Status)
	}
}

func TestGetLeaveRequest_NotFound_Errors(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	_, err := repo.GetLeaveRequest(ctx, testutil.OwnerID, 999999999)
	if err == nil {
		t.Error("expected error for non-existent leave request, got nil")
	}
}

func TestListLeaveBalances_ZeroOwner_ReturnsEmpty(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	balances, err := repo.ListLeaveBalances(ctx, 0, 4101, 2026)
	if err != nil {
		t.Fatalf("ListLeaveBalances zero owner: %v", err)
	}
	if len(balances) != 0 {
		t.Errorf("expected 0 balances for owner 0, got %d", len(balances))
	}
}

func TestListLeaveCalendar_NoError(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()

	records, err := repo.ListLeaveCalendar(ctx, testutil.OwnerID, 0, "2026-01-01", "2026-12-31")
	if err != nil {
		t.Fatalf("ListLeaveCalendar: %v", err)
	}
	_ = records
}

func TestCancelLeaveRequest_OnlyWhenPending(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	name := fmt.Sprintf("Cancel test %d", time.Now().UnixNano())

	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name:             name,
		RequiresApproval: false,
		IsActive:         true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_types WHERE id = ?", lt.ID) })

	pending, err := repo.CreateLeaveRequest(ctx, testutil.OwnerID, domain.LeaveRequestRecord{
		EmployeeID: 4101, LeaveTypeID: lt.ID, StartDate: "2035-03-01", EndDate: "2035-03-02", TotalDays: 2, Status: domain.LeaveStatusPending,
	})
	if err != nil {
		t.Fatalf("CreateLeaveRequest (pending): %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_requests WHERE id = ?", pending.ID) })

	canceled, err := repo.CancelLeaveRequest(ctx, testutil.OwnerID, pending.ID)
	if err != nil {
		t.Fatalf("CancelLeaveRequest: %v", err)
	}
	if !canceled {
		t.Fatal("want a pending request to be cancelable")
	}

	got, err := repo.GetLeaveRequest(ctx, testutil.OwnerID, pending.ID)
	if err != nil {
		t.Fatalf("GetLeaveRequest after cancel: %v", err)
	}
	if got.Status != domain.LeaveStatusCanceled {
		t.Fatalf("status after cancel = %q, want %q", got.Status, domain.LeaveStatusCanceled)
	}

	// Already canceled - a second attempt must report false, not re-cancel
	// or error, since the row no longer matches the query's pending guard.
	canceledAgain, err := repo.CancelLeaveRequest(ctx, testutil.OwnerID, pending.ID)
	if err != nil {
		t.Fatalf("CancelLeaveRequest (already canceled): %v", err)
	}
	if canceledAgain {
		t.Fatal("want an already-canceled request to report false, not cancel again")
	}

	approved, err := repo.CreateLeaveRequest(ctx, testutil.OwnerID, domain.LeaveRequestRecord{
		EmployeeID: 4101, LeaveTypeID: lt.ID, StartDate: "2035-04-01", EndDate: "2035-04-02", TotalDays: 2, Status: domain.LeaveStatusApproved,
	})
	if err != nil {
		t.Fatalf("CreateLeaveRequest (approved): %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_requests WHERE id = ?", approved.ID) })

	canceledApproved, err := repo.CancelLeaveRequest(ctx, testutil.OwnerID, approved.ID)
	if err != nil {
		t.Fatalf("CancelLeaveRequest (approved): %v", err)
	}
	if canceledApproved {
		t.Fatal("want an approved request to not be cancelable through this path")
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

// Range semantics per voxie#47: a request is included when it starts on/before
// ToDate and ends on/after FromDate, so multi-day leave crossing either
// boundary is never missed (rtylr#838's warning depends on that).
func TestListLeaveRequests_DateRangeOverlap(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	name := fmt.Sprintf("Range overlap %d", time.Now().UnixNano())

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
		{EmployeeID: 4101, LeaveTypeID: lt.ID, StartDate: "2035-04-01", EndDate: "2035-04-10", TotalDays: 10, Status: domain.LeaveStatusApproved}, // spans the window
		{EmployeeID: 4101, LeaveTypeID: lt.ID, StartDate: "2035-05-01", EndDate: "2035-05-02", TotalDays: 2, Status: domain.LeaveStatusApproved},  // after it
		{EmployeeID: 4101, LeaveTypeID: lt.ID, StartDate: "2035-03-01", EndDate: "2035-03-02", TotalDays: 2, Status: domain.LeaveStatusApproved},  // before it
	} {
		created, createErr := repo.CreateLeaveRequest(ctx, testutil.OwnerID, request)
		if createErr != nil {
			t.Fatalf("CreateLeaveRequest: %v", createErr)
		}
		requestIDs = append(requestIDs, created.ID)
	}

	from, to := "2035-04-05", "2035-04-06"
	requests, total, err := repo.ListLeaveRequests(ctx, testutil.OwnerID, domain.ListLeaveRequestsParams{
		EmployeeID: 4101,
		Status:     domain.LeaveStatusApproved,
		FromDate:   &from,
		ToDate:     &to,
		Limit:      100,
	})
	if err != nil {
		t.Fatalf("ListLeaveRequests with range: %v", err)
	}
	if total != 1 || len(requests) != 1 || requests[0].StartDate != "2035-04-01" {
		t.Fatalf("expected exactly the spanning request, got total=%d len=%d", total, len(requests))
	}

	all, totalAll, err := repo.ListLeaveRequests(ctx, testutil.OwnerID, domain.ListLeaveRequestsParams{
		EmployeeID: 4101,
		Status:     domain.LeaveStatusApproved,
		Limit:      100,
	})
	if err != nil {
		t.Fatalf("ListLeaveRequests without range: %v", err)
	}
	if totalAll != 3 || len(all) != 3 {
		t.Fatalf("open-ended range should return all three, got total=%d len=%d", totalAll, len(all))
	}
}
