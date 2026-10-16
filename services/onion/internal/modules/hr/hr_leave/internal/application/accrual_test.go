package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
)

type mockAccrualRepo struct {
	repository
	leaveTypes        []domain.LeaveTypeRecord
	balances          []domain.LeaveBalanceRecord
	addedEntitledDays []addedDaysCall
}

type addedDaysCall struct {
	OwnerID     int64
	EmployeeID  int64
	LeaveTypeID int64
	Year        int32
	Days        float64
	Month       int32
}

func (m *mockAccrualRepo) ListAllActiveLeaveTypesForOwner(_ context.Context, _ int64) ([]domain.LeaveTypeRecord, error) {
	return m.leaveTypes, nil
}

func (m *mockAccrualRepo) ListBalancesForAccrual(_ context.Context, _ int64, _, _ int32) ([]domain.LeaveBalanceRecord, error) {
	return m.balances, nil
}

func (m *mockAccrualRepo) AddToEntitledDays(_ context.Context, ownerID, employeeID, leaveTypeID int64, year int32, days float64, month int32) error {
	m.addedEntitledDays = append(m.addedEntitledDays, addedDaysCall{
		OwnerID:     ownerID,
		EmployeeID:  employeeID,
		LeaveTypeID: leaveTypeID,
		Year:        year,
		Days:        days,
		Month:       month,
	})
	return nil
}

func (m *mockAccrualRepo) ApplyYearEndCarryover(_ context.Context, _ int64, _, _ int32) error {
	return nil
}

func TestAccrueForOwner_MonthlyCatchup(t *testing.T) {
	ctx := context.Background()
	repo := &mockAccrualRepo{
		leaveTypes: []domain.LeaveTypeRecord{
			{
				ID:            101,
				AccrualRule:   domain.AccrualRuleMonthly,
				AccrualAmount: 1.5,
				IsActive:      true,
			},
		},
		balances: []domain.LeaveBalanceRecord{
			{
				EmployeeID:       1,
				LeaveTypeID:      101,
				LastAccrualMonth: 0, // Never accrued yet this year
			},
		},
	}
	s := &Service{repo: repo}

	// Month 3 (March)
	if err := s.accrueForOwner(ctx, 42, 2026, 3); err != nil {
		t.Fatalf("accrueForOwner failed: %v", err)
	}

	if len(repo.addedEntitledDays) != 1 {
		t.Fatalf("expected 1 call to AddToEntitledDays, got %d", len(repo.addedEntitledDays))
	}

	call := repo.addedEntitledDays[0]
	// 3 months owed (1, 2, 3) * 1.5 = 4.5 days
	if call.Days != 4.5 {
		t.Errorf("expected 4.5 days accrued, got %v", call.Days)
	}
	if call.Month != 3 {
		t.Errorf("expected target month 3, got %d", call.Month)
	}
}

func TestAccrueForOwner_AnnualAccrual(t *testing.T) {
	ctx := context.Background()
	repo := &mockAccrualRepo{
		leaveTypes: []domain.LeaveTypeRecord{
			{
				ID:            202,
				AccrualRule:   domain.AccrualRuleAnnual,
				AccrualAmount: 20.0,
				IsActive:      true,
			},
		},
		balances: []domain.LeaveBalanceRecord{
			{
				EmployeeID:       2,
				LeaveTypeID:      202,
				LastAccrualMonth: 0,
			},
			{
				EmployeeID:       3,
				LeaveTypeID:      202,
				LastAccrualMonth: 12, // already accrued
			},
		},
	}
	s := &Service{repo: repo}

	if err := s.accrueForOwner(ctx, 42, 2026, 1); err != nil {
		t.Fatalf("accrueForOwner failed: %v", err)
	}

	if len(repo.addedEntitledDays) != 1 {
		t.Fatalf("expected 1 call to AddToEntitledDays (only for employee 2), got %d", len(repo.addedEntitledDays))
	}

	call := repo.addedEntitledDays[0]
	if call.EmployeeID != 2 {
		t.Errorf("expected employee 2, got %d", call.EmployeeID)
	}
	if call.Days != 20.0 {
		t.Errorf("expected 20.0 days, got %v", call.Days)
	}
	if call.Month != 12 {
		t.Errorf("expected month 12 to mark annual accrual complete, got %d", call.Month)
	}
}
