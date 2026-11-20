package application

import (
	"context"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/hr/hr_leave/internal/domain"
)

type seedCall struct {
	BusinessID int64
	EmployeeID int64
	Year       int32
}

type mockSeedRepo struct {
	mockAccrualRepo
	seeded    int64
	seedCalls []seedCall
}

func (m *mockSeedRepo) SeedMissingLeaveBalances(_ context.Context, businessID, employeeID int64, year int32) (int64, error) {
	m.seedCalls = append(m.seedCalls, seedCall{BusinessID: businessID, EmployeeID: employeeID, Year: year})
	return m.seeded, nil
}

func (m *mockSeedRepo) CreateLeaveType(_ context.Context, _ int64, lt domain.LeaveTypeRecord) (*domain.LeaveTypeRecord, error) {
	lt.ID = 101
	return &lt, nil
}

func annualLeaveSeedRepo(seeded int64) *mockSeedRepo {
	return &mockSeedRepo{
		seeded: seeded,
		mockAccrualRepo: mockAccrualRepo{
			leaveTypes: []domain.LeaveTypeRecord{
				{ID: 101, AccrualRule: domain.AccrualRuleAnnual, AccrualAmount: 21, IsActive: true},
			},
			balances: []domain.LeaveBalanceRecord{
				{EmployeeID: 7, LeaveTypeID: 101, LastAccrualMonth: 0},
			},
		},
	}
}

// A balance created on demand must be usable at once, not sit at zero until
// the monthly job's next run.
func TestEnsureLeaveBalancesAccruesNewlySeededRows(t *testing.T) {
	repo := annualLeaveSeedRepo(1)
	s := &Service{repo: repo}

	if err := s.ensureLeaveBalances(context.Background(), 42, 7); err != nil {
		t.Fatalf("ensureLeaveBalances: %v", err)
	}

	wantYear := int32(time.Now().UTC().Year())
	if len(repo.seedCalls) != 1 || repo.seedCalls[0] != (seedCall{BusinessID: 42, EmployeeID: 7, Year: wantYear}) {
		t.Fatalf("seed calls = %+v, want one for business 42, employee 7, year %d", repo.seedCalls, wantYear)
	}
	if len(repo.addedEntitledDays) != 1 {
		t.Fatalf("accrual calls = %d, want 1", len(repo.addedEntitledDays))
	}
	if got := repo.addedEntitledDays[0]; got.EmployeeID != 7 || got.LeaveTypeID != 101 || got.Days != 21 || got.Year != wantYear {
		t.Fatalf("accrual = %+v, want 21 days for employee 7 on leave type 101", got)
	}
}

func TestEnsureLeaveBalancesSkipsAccrualWhenNothingWasMissing(t *testing.T) {
	repo := annualLeaveSeedRepo(0)
	s := &Service{repo: repo}

	if err := s.ensureLeaveBalances(context.Background(), 42, 7); err != nil {
		t.Fatalf("ensureLeaveBalances: %v", err)
	}
	if len(repo.addedEntitledDays) != 0 {
		t.Fatalf("accrual ran %d times with no new balance rows", len(repo.addedEntitledDays))
	}
}

// Leave types are usually configured after the team exists; creating one has
// to give every current employee a balance for it.
func TestCreateLeaveTypeSeedsBalancesForTheWholeBusiness(t *testing.T) {
	repo := annualLeaveSeedRepo(3)
	s := &Service{repo: repo}

	if _, err := s.CreateLeaveType(context.Background(), 42, domain.LeaveTypeRecord{
		Name: "Annual", AccrualRule: domain.AccrualRuleAnnual, AccrualAmount: 21,
	}); err != nil {
		t.Fatalf("CreateLeaveType: %v", err)
	}
	if len(repo.seedCalls) != 1 || repo.seedCalls[0].BusinessID != 42 || repo.seedCalls[0].EmployeeID != 0 {
		t.Fatalf("seed calls = %+v, want one business-wide seed for business 42", repo.seedCalls)
	}
}
