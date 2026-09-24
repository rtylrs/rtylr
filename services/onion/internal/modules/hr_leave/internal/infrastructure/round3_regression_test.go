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

func TestListLeaveCalendar_FiltersByLocationAndBusiness(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	stamp := time.Now().UnixNano()
	locationOne := int64(880001 + stamp%1000)
	locationTwo := locationOne + 1

	ownType, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name: fmt.Sprintf("calendar own %d", stamp), IsPaid: true, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create own leave type: %v", err)
	}
	otherType, err := repo.CreateLeaveType(ctx, testutil.OtherBusinessID, domain.LeaveTypeRecord{
		Name: fmt.Sprintf("calendar other %d", stamp), IsPaid: true, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create other leave type: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_requests WHERE leave_type_id IN (?, ?)", ownType.ID, otherType.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_types WHERE id IN (?, ?)", ownType.ID, otherType.ID)
	})

	insertEmployee := func(businessID, locationID int64, label string) int64 {
		result, insertErr := db.ExecContext(ctx, `
			INSERT INTO hr_employees
			(owner_id, business_id, location_id, first_name, last_name, email, status)
			VALUES (?, ?, ?, ?, ?, ?, 'active')`, businessID, businessID, locationID, "Calendar", label, fmt.Sprintf("calendar-%d-%s@example.test", stamp, label))
		if insertErr != nil {
			t.Fatalf("insert employee %s: %v", label, insertErr)
		}
		id, idErr := result.LastInsertId()
		if idErr != nil {
			t.Fatalf("employee %s id: %v", label, idErr)
		}
		t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM hr_employees WHERE id = ?", id) })
		return id
	}

	ownLocationOne := insertEmployee(testutil.OwnerID, locationOne, fmt.Sprintf("own-one-%d", stamp))
	ownLocationTwo := insertEmployee(testutil.OwnerID, locationTwo, fmt.Sprintf("own-two-%d", stamp))
	otherLocationOne := insertEmployee(testutil.OtherBusinessID, locationOne, fmt.Sprintf("other-one-%d", stamp))

	requests := []struct {
		employeeID int64
		typeID     int64
	}{
		{ownLocationOne, ownType.ID},
		{ownLocationTwo, ownType.ID},
		{otherLocationOne, otherType.ID},
	}
	var requestIDs []int64
	for _, item := range requests {
		created, createErr := repo.CreateLeaveRequest(ctx, map[int64]int64{
			ownLocationOne: testutil.OwnerID, ownLocationTwo: testutil.OwnerID, otherLocationOne: testutil.OtherBusinessID,
		}[item.employeeID], domain.LeaveRequestRecord{
			EmployeeID: item.employeeID, LeaveTypeID: item.typeID,
			StartDate: "2042-02-01", EndDate: "2042-02-01", TotalDays: 1,
			Status: domain.LeaveStatusApproved,
		})
		if createErr != nil {
			t.Fatalf("create leave request for employee %d: %v", item.employeeID, createErr)
		}
		requestIDs = append(requestIDs, created.ID)
	}
	t.Cleanup(func() {
		for _, requestID := range requestIDs {
			_, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_requests WHERE id = ?", requestID)
		}
	})

	one, err := repo.ListLeaveCalendar(ctx, testutil.OwnerID, locationOne, "2042-01-01", "2042-12-31")
	if err != nil {
		t.Fatalf("location one calendar: %v", err)
	}
	if len(one) != 1 || one[0].EmployeeID != ownLocationOne {
		t.Fatalf("location one returned %+v, want only employee %d", one, ownLocationOne)
	}

	two, err := repo.ListLeaveCalendar(ctx, testutil.OwnerID, locationTwo, "2042-01-01", "2042-12-31")
	if err != nil {
		t.Fatalf("location two calendar: %v", err)
	}
	if len(two) != 1 || two[0].EmployeeID != ownLocationTwo {
		t.Fatalf("location two returned %+v, want only employee %d", two, ownLocationTwo)
	}

	unfiltered, err := repo.ListLeaveCalendar(ctx, testutil.OwnerID, 0, "2042-01-01", "2042-12-31")
	if err != nil {
		t.Fatalf("unfiltered calendar: %v", err)
	}
	if len(unfiltered) != 2 {
		t.Fatalf("unfiltered own-business calendar returned %d records, want 2", len(unfiltered))
	}
}

func TestUpdateLeaveTypePatch_PreservesOmittedFieldsAndRejectsMissingRow(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	name := fmt.Sprintf("patch leave type %d", time.Now().UnixNano())
	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{
		Name: name, IsPaid: true, RequiresApproval: true, AccrualRule: domain.AccrualRuleMonthly,
		AccrualAmount: 2.5, MaxCarryoverDays: 10, MinNoticeDays: 3,
		RequiresDocumentationAfterDays: 7, IsActive: true,
	})
	if err != nil {
		t.Fatalf("create leave type: %v", err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_types WHERE id = ?", lt.ID) })

	renamed := "Renamed patch leave type"
	if err := repo.UpdateLeaveTypePatch(ctx, testutil.OwnerID, lt.ID, domain.LeaveTypePatch{Name: &renamed}); err != nil {
		t.Fatalf("patch name: %v", err)
	}
	got, err := repo.GetLeaveType(ctx, testutil.OwnerID, lt.ID)
	if err != nil {
		t.Fatalf("get patched type: %v", err)
	}
	if got.Name != renamed || !got.IsPaid || !got.RequiresApproval || got.AccrualRule != domain.AccrualRuleMonthly ||
		got.AccrualAmount != 2.5 || got.MaxCarryoverDays != 10 || got.MinNoticeDays != 3 ||
		got.RequiresDocumentationAfterDays != 7 || !got.IsActive {
		t.Fatalf("omitted fields changed after name patch: %+v", got)
	}

	disabled := false
	zero := int32(0)
	rule := domain.AccrualRuleNone
	if err := repo.UpdateLeaveTypePatch(ctx, testutil.OwnerID, lt.ID, domain.LeaveTypePatch{
		IsPaid: &disabled, MaxCarryoverDays: &zero, AccrualRule: &rule,
	}); err != nil {
		t.Fatalf("patch explicit false/zero values: %v", err)
	}
	got, err = repo.GetLeaveType(ctx, testutil.OwnerID, lt.ID)
	if err != nil {
		t.Fatalf("get explicit-value type: %v", err)
	}
	if got.IsPaid || got.MaxCarryoverDays != 0 || got.AccrualRule != domain.AccrualRuleNone {
		t.Fatalf("explicit false/zero patch was not applied: %+v", got)
	}

	if err := repo.UpdateLeaveTypePatch(ctx, testutil.OwnerID, 999999999, domain.LeaveTypePatch{Name: &renamed}); !errors.Is(err, domain.ErrLeaveTypeNotFound) {
		t.Fatalf("missing leave type error = %v, want ErrLeaveTypeNotFound", err)
	}
}

func TestCountLeaveBalances_ReturnsTotalForOptionalEmployeeFilter(t *testing.T) {
	db := testutil.OpenDB(t)
	repo := infrastructure.NewRepository(db)
	ctx := context.Background()
	year := int32(2043)
	name := fmt.Sprintf("count balances %d", time.Now().UnixNano())
	lt, err := repo.CreateLeaveType(ctx, testutil.OwnerID, domain.LeaveTypeRecord{Name: name, IsActive: true})
	if err != nil {
		t.Fatalf("create leave type: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_balances WHERE leave_type_id = ?", lt.ID)
		_, _ = db.ExecContext(ctx, "DELETE FROM hr_leave_types WHERE id = ?", lt.ID)
	})
	for _, employeeID := range []int64{880101, 880102} {
		if err := repo.UpsertLeaveBalance(ctx, testutil.OwnerID, domain.LeaveBalanceRecord{
			EmployeeID: employeeID, LeaveTypeID: lt.ID, Year: year, EntitledDays: 1,
		}); err != nil {
			t.Fatalf("upsert balance for %d: %v", employeeID, err)
		}
	}
	total, err := repo.CountLeaveBalances(ctx, testutil.OwnerID, 0, year)
	if err != nil {
		t.Fatalf("count all balances: %v", err)
	}
	if total != 2 {
		t.Fatalf("all-employee balance total = %d, want 2", total)
	}
	total, err = repo.CountLeaveBalances(ctx, testutil.OwnerID, 880101, year)
	if err != nil {
		t.Fatalf("count one balance: %v", err)
	}
	if total != 1 {
		t.Fatalf("single-employee balance total = %d, want 1", total)
	}
}
