package application

import (
	"context"
	"testing"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
)

type mockPolicyRepo struct {
	mockSeedRepo
	stored  domain.LeaveTypeRecord
	patched []domain.LeaveTypePatch
}

func (m *mockPolicyRepo) GetLeaveType(_ context.Context, _, _ int64) (*domain.LeaveTypeRecord, error) {
	lt := m.stored
	return &lt, nil
}

func (m *mockPolicyRepo) UpdateLeaveTypePatch(_ context.Context, _, _ int64, patch domain.LeaveTypePatch) error {
	m.patched = append(m.patched, patch)
	return nil
}

func newPolicyRepo() *mockPolicyRepo {
	repo := &mockPolicyRepo{stored: domain.LeaveTypeRecord{ID: 101, Name: "Annual", IsActive: true}}
	repo.leaveTypes = []domain.LeaveTypeRecord{
		{ID: 101, AccrualRule: domain.AccrualRuleAnnual, AccrualAmount: 21, IsActive: true},
	}
	// Balances already exist, so nothing is seeded; the allowance must still land.
	repo.balances = []domain.LeaveBalanceRecord{{EmployeeID: 7, LeaveTypeID: 101, LastAccrualMonth: 0}}
	return repo
}

// Giving an existing leave type an allowance has to reach the balances that
// already exist for it, not wait for the monthly job.
func TestUpdateLeaveTypePatchAppliesANewAllowanceToExistingBalances(t *testing.T) {
	repo := newPolicyRepo()
	s := &Service{repo: repo}
	rule, amount := domain.AccrualRuleAnnual, 21.0

	if _, err := s.UpdateLeaveTypePatch(context.Background(), 42, 101, domain.LeaveTypePatch{
		AccrualRule: &rule, AccrualAmount: &amount,
	}); err != nil {
		t.Fatalf("UpdateLeaveTypePatch: %v", err)
	}

	if len(repo.patched) != 1 || repo.patched[0].AccrualAmount == nil || *repo.patched[0].AccrualAmount != 21 {
		t.Fatalf("stored patch = %+v, want the 21 day allowance", repo.patched)
	}
	if len(repo.addedEntitledDays) != 1 || repo.addedEntitledDays[0].Days != 21 || repo.addedEntitledDays[0].EmployeeID != 7 {
		t.Fatalf("accrual = %+v, want 21 days granted to employee 7", repo.addedEntitledDays)
	}
}

func TestUpdateLeaveTypePatchLeavesBalancesAloneForOtherFields(t *testing.T) {
	repo := newPolicyRepo()
	s := &Service{repo: repo}
	name := "Vacation"

	if _, err := s.UpdateLeaveTypePatch(context.Background(), 42, 101, domain.LeaveTypePatch{Name: &name}); err != nil {
		t.Fatalf("UpdateLeaveTypePatch: %v", err)
	}
	if len(repo.seedCalls) != 0 || len(repo.addedEntitledDays) != 0 {
		t.Fatalf("a rename touched balances: seeds=%d accruals=%d", len(repo.seedCalls), len(repo.addedEntitledDays))
	}
}
