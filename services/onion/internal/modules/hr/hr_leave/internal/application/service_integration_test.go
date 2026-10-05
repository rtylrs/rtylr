package application

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/infrastructure"
	"github.com/voxire/onion/internal/testutil"
)

// TestGetLeaveBalances_ScopesToCallerAgainstARealRepository exercises the new
// authorization check through a real database rather than a fake, since
// scopeLeaveRequests being correct in isolation (authorization_test.go)
// doesn't prove the resolved employee id actually reaches the repository call.
func TestGetLeaveBalances_ScopesToCallerAgainstARealRepository(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	ownerID := testutil.OwnerID
	year := int32(2031)
	name := fmt.Sprintf("Balance scoping test %d", time.Now().UnixNano())

	lt, err := repo.CreateLeaveType(ctx, ownerID, domain.LeaveTypeRecord{
		Name: name, RequiresApproval: false, IsActive: true,
	})
	if err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_types WHERE id = ?", lt.ID) })

	const callerEmployeeID = 4201
	const otherEmployeeID = 4202

	for _, employeeID := range []int64{callerEmployeeID, otherEmployeeID} {
		if err := repo.UpsertLeaveBalance(ctx, ownerID, domain.LeaveBalanceRecord{
			EmployeeID: employeeID, LeaveTypeID: lt.ID, Year: year, EntitledDays: 20,
		}); err != nil {
			t.Fatalf("UpsertLeaveBalance(%d): %v", employeeID, err)
		}
		t.Cleanup(func() {
			_, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_balances WHERE employee_id = ? AND year = ?", employeeID, year)
		})
	}

	t.Run("an ordinary caller is narrowed to their own balances regardless of what they asked for", func(t *testing.T) {
		svc := &Service{repo: repo, permissions: fakeLeavePermissionProvider{employeeID: callerEmployeeID}}

		balances, err := svc.GetLeaveBalances(ctx, ownerID, 9001, otherEmployeeID, year)
		if err != nil {
			t.Fatalf("GetLeaveBalances: %v", err)
		}
		for _, b := range balances {
			if b.EmployeeID != callerEmployeeID {
				t.Fatalf("balance for employee %d leaked to a caller scoped to %d", b.EmployeeID, callerEmployeeID)
			}
		}
	})

	t.Run("an approver sees the requested employee's balances", func(t *testing.T) {
		svc := &Service{
			repo:        repo,
			permissions: fakeLeavePermissionProvider{employeeID: callerEmployeeID, canApprove: true},
		}

		balances, err := svc.GetLeaveBalances(ctx, ownerID, 9001, otherEmployeeID, year)
		if err != nil {
			t.Fatalf("GetLeaveBalances: %v", err)
		}
		found := false
		for _, b := range balances {
			if b.EmployeeID == otherEmployeeID {
				found = true
			}
		}
		if !found {
			t.Fatal("approver did not receive the requested employee's balance")
		}
	})
}

func TestGetLeaveBalances_DeniesUnresolvableCallerAgainstARealRepository(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)

	svc := &Service{
		repo:        repo,
		permissions: fakeLeavePermissionProvider{employeeErr: errors.New("employee lookup failed")},
	}

	_, err := svc.GetLeaveBalances(context.Background(), testutil.OwnerID, 9001, 4202, 2031)
	if !errors.Is(err, domain.ErrLeaveAccessDenied) {
		t.Fatalf("error = %v, want ErrLeaveAccessDenied", err)
	}
}
