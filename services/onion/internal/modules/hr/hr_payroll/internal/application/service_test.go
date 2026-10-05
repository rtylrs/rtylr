package application

import (
	"context"
	"errors"
	"testing"
	"time"

	moduleinfra "github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/infrastructure"

	"github.com/voxire/onion/internal/modules/hr/hr_payroll/internal/domain"
	"github.com/voxire/onion/internal/testutil"
)

func TestNoopFinancePoster_ReportsUnavailable(t *testing.T) {
	err := (noopFinancePoster{}).PostPayrollJournal(context.Background(), domain.PayrollJournalEntry{})
	if !errors.Is(err, domain.ErrFinanceIntegrationUnavailable) {
		t.Fatalf("noop finance poster error = %v, want %v", err, domain.ErrFinanceIntegrationUnavailable)
	}
}

func TestNoopNotifier_DoesNotError(t *testing.T) {
	if err := (noopNotifier{}).Dispatch(context.Background(), domain.NotificationEvent{}); err != nil {
		t.Fatalf("noop notifier returned error: %v", err)
	}
}

func TestTriggerPayrollRun_RejectsInvalidDates(t *testing.T) {
	s := &Service{}
	cases := []struct {
		name, start, end string
	}{
		{"unparseable start", "not-a-date", "2026-04-30"},
		{"unparseable end", "2026-04-01", "31-04-2026"},
		{"end before start", "2026-04-30", "2026-04-01"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := s.TriggerPayrollRun(context.Background(), 1, 1, tc.start, tc.end)
			if !errors.Is(err, domain.ErrInvalidDateRange) {
				t.Fatalf("expected ErrInvalidDateRange for %s/%s, got %v", tc.start, tc.end, err)
			}
		})
	}
}

type spyNotifier struct {
	events []domain.NotificationEvent
}

func (s *spyNotifier) Dispatch(_ context.Context, event domain.NotificationEvent) error {
	s.events = append(s.events, event)
	return nil
}

func TestSetNotifier_NilIsIgnored(t *testing.T) {
	s := &Service{notifier: noopNotifier{}}
	s.SetNotifier(nil)
	if _, ok := s.notifier.(noopNotifier); !ok {
		t.Fatalf("expected notifier to remain noopNotifier after SetNotifier(nil)")
	}
}

func TestSetNotifier_RealOverridesNoop(t *testing.T) {
	s := &Service{notifier: noopNotifier{}}
	spy := &spyNotifier{}
	s.SetNotifier(spy)
	if s.notifier != spy {
		t.Fatalf("expected notifier replaced with spy")
	}
}

func TestSetFinancePoster_NilIsIgnored(t *testing.T) {
	s := &Service{financePoster: noopFinancePoster{}}
	s.SetFinancePoster(nil)
	if _, ok := s.financePoster.(noopFinancePoster); !ok {
		t.Fatalf("expected financePoster to remain noopFinancePoster after SetFinancePoster(nil)")
	}
}

type spyFinancePoster struct {
	entries []domain.PayrollJournalEntry
}

func (s *spyFinancePoster) PostPayrollJournal(_ context.Context, entry domain.PayrollJournalEntry) error {
	s.entries = append(s.entries, entry)
	return nil
}

func TestSetFinancePoster_RealOverridesNoop(t *testing.T) {
	s := &Service{financePoster: noopFinancePoster{}}
	spy := &spyFinancePoster{}
	s.SetFinancePoster(spy)
	if s.financePoster != spy {
		t.Fatalf("expected financePoster replaced with spy")
	}
}

func TestCalculateMonthlyBasePay_UsesInclusiveCalendarDays(t *testing.T) {
	start := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.April, 30, 0, 0, 0, 0, time.UTC)

	if got := calculateMonthlyBasePay(3000, start, end); got != 3000 {
		t.Errorf("full April salary = %.2f, want 3000.00", got)
	}
}

func TestCalculateMonthlyBasePay_HandlesMonthLengths(t *testing.T) {
	cases := []struct {
		name  string
		start time.Time
		end   time.Time
		want  float64
	}{
		{
			name:  "February in a non-leap year",
			start: time.Date(2025, time.February, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2025, time.February, 28, 0, 0, 0, 0, time.UTC),
			want:  3000,
		},
		{
			name:  "February in a leap year",
			start: time.Date(2024, time.February, 1, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2024, time.February, 29, 0, 0, 0, 0, time.UTC),
			want:  3000,
		},
		{
			name:  "single inclusive day",
			start: time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC),
			end:   time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC),
			want:  3000.0 / 31.0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := calculateMonthlyBasePay(3000, tc.start, tc.end); got != tc.want {
				t.Errorf("base pay = %.8f, want %.8f", got, tc.want)
			}
		})
	}
}

func TestCalculateMonthlyBasePay_SplitsCrossMonthPeriods(t *testing.T) {
	start := time.Date(2026, time.April, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.May, 14, 0, 0, 0, 0, time.UTC)
	want := 3000.0*16.0/30.0 + 3000.0*14.0/31.0

	if got := calculateMonthlyBasePay(3000, start, end); got != want {
		t.Errorf("cross-month base pay = %.8f, want %.8f", got, want)
	}
}

func TestTriggerPayrollRun_StartAfterEndRejected(t *testing.T) {
	s := &Service{}
	_, _, err := s.TriggerPayrollRun(context.Background(), 1, 1, "2026-06-30", "2026-06-01")
	if !errors.Is(err, domain.ErrInvalidDateRange) {
		t.Errorf("start after end should return ErrInvalidDateRange, got %v", err)
	}
}

func TestTriggerPayrollRun_MalformedStartDate(t *testing.T) {
	s := &Service{}
	_, _, err := s.TriggerPayrollRun(context.Background(), 1, 1, "01/04/2026", "2026-04-30")
	if !errors.Is(err, domain.ErrInvalidDateRange) {
		t.Errorf("malformed start date should return ErrInvalidDateRange, got %v", err)
	}
}

func TestTriggerPayrollRun_MalformedEndDate(t *testing.T) {
	s := &Service{}
	_, _, err := s.TriggerPayrollRun(context.Background(), 1, 1, "2026-04-01", "30-04-2026")
	if !errors.Is(err, domain.ErrInvalidDateRange) {
		t.Errorf("malformed end date should return ErrInvalidDateRange, got %v", err)
	}
}

func TestTriggerPayrollRun_BothDatesBlank(t *testing.T) {
	s := &Service{}
	_, _, err := s.TriggerPayrollRun(context.Background(), 1, 1, "", "")
	if !errors.Is(err, domain.ErrInvalidDateRange) {
		t.Errorf("blank dates should return ErrInvalidDateRange, got %v", err)
	}
}

type fakePermissionProvider struct {
	employeeID  int64
	employeeErr error
	allowed     bool
	permErr     error
}

func (f fakePermissionProvider) GetEmployeeIDByUserID(_ context.Context, _, _ int64) (int64, error) {
	return f.employeeID, f.employeeErr
}

func (f fakePermissionProvider) HasPermission(_ context.Context, _, _ int64, _ string) (bool, error) {
	return f.allowed, f.permErr
}

func (f fakePermissionProvider) HasRole(_ context.Context, _, _, _ int64) (bool, error) {
	return f.allowed, f.permErr
}

func TestService_GetEmployeeIDByUserID_DelegatesToPermissionProvider(t *testing.T) {
	s := &Service{permissions: fakePermissionProvider{employeeID: 4101}}
	id, err := s.GetEmployeeIDByUserID(context.Background(), 1001, 5555)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != 4101 {
		t.Errorf("employeeID = %d, want 4101", id)
	}
}

func TestService_HasPermission_DelegatesToPermissionProvider(t *testing.T) {
	s := &Service{permissions: fakePermissionProvider{allowed: true}}
	allowed, err := s.HasPermission(context.Background(), 1001, 4101, "manage_payroll")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected allowed=true")
	}
}

func TestService_HasPermission_FalseWhenNotGranted(t *testing.T) {
	s := &Service{permissions: fakePermissionProvider{allowed: false}}
	allowed, err := s.HasPermission(context.Background(), 1001, 4102, "manage_payroll")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected allowed=false")
	}
}

func TestValidatePayrollApprover_RejectsRunCreator(t *testing.T) {
	err := validatePayrollApprover(&domain.PayrollRunRecord{CreatedBy: 4101}, 4101)
	if !errors.Is(err, domain.ErrPayrollSelfApproval) {
		t.Fatalf("expected ErrPayrollSelfApproval, got %v", err)
	}
}

func TestValidatePayrollApprover_AllowsDifferentApprover(t *testing.T) {
	if err := validatePayrollApprover(&domain.PayrollRunRecord{CreatedBy: 4101}, 4102); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestApproverMatchesChain_ExplicitEmployee(t *testing.T) {
	s := &Service{permissions: fakePermissionProvider{}}
	allowed, err := s.approverMatchesChain(context.Background(), 1001, 4101, []domain.ApprovalLevelInfo{
		{ApproverEmployeeID: 4101},
	})
	if err != nil || !allowed {
		t.Fatalf("expected explicit employee approver to match, allowed=%v err=%v", allowed, err)
	}
}

func TestApproverMatchesChain_Role(t *testing.T) {
	s := &Service{permissions: fakePermissionProvider{allowed: true}}
	allowed, err := s.approverMatchesChain(context.Background(), 1001, 4101, []domain.ApprovalLevelInfo{
		{ApproverRoleID: 7},
	})
	if err != nil || !allowed {
		t.Fatalf("expected role approver to match, allowed=%v err=%v", allowed, err)
	}
}

func TestApproverMatchesChain_RejectsUnmatchedApprover(t *testing.T) {
	s := &Service{permissions: fakePermissionProvider{}}
	allowed, err := s.approverMatchesChain(context.Background(), 1001, 4101, []domain.ApprovalLevelInfo{
		{ApproverEmployeeID: 4102},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected unmatched approver to be rejected")
	}
}

func TestApproverMatchesChain_RejectsMultiLevelChainWithoutPersistedCursor(t *testing.T) {
	s := &Service{permissions: fakePermissionProvider{allowed: true}}
	allowed, err := s.approverMatchesChain(context.Background(), 1001, 4101, []domain.ApprovalLevelInfo{
		{LevelOrder: 1, ApproverRoleID: 7},
		{LevelOrder: 2, ApproverRoleID: 8},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected multi-level chain to fail closed")
	}
}

func TestApproverMatchesChain_RejectsNonFirstLevel(t *testing.T) {
	s := &Service{permissions: fakePermissionProvider{allowed: true}}
	allowed, err := s.approverMatchesChain(context.Background(), 1001, 4101, []domain.ApprovalLevelInfo{
		{LevelOrder: 2, ApproverEmployeeID: 4101},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Fatal("expected non-first approval level to be rejected")
	}
}

func TestCalculateOvertimePay_UsesWeightedMinutes(t *testing.T) {
	// 60 minutes at 2x plus 60 minutes at 1.5x equals 210 weighted minutes,
	// not 120 minutes multiplied by an unweighted average in a later step.
	got := calculateOvertimePay(60, 210, 0, 1.5)
	if got != 210 {
		t.Fatalf("overtime pay = %.2f, want 210.00", got)
	}
}

func TestPayrollPayCents_ComposesRoundedComponents(t *testing.T) {
	basePayCents, overtimePayCents, netPayCents := calculatePayCents(10.13, 15.19)

	if basePayCents != 1013 || overtimePayCents != 1519 {
		t.Fatalf("component cents = %d + %d, want 1013 + 1519", basePayCents, overtimePayCents)
	}
	if netPayCents != 2532 {
		t.Fatalf("net pay cents = %d, want 2532", netPayCents)
	}
}

func TestNoopFinancePoster_IsFinanceJournalPoster(_ *testing.T) {
	var _ domain.FinanceJournalPoster = noopFinancePoster{}
}

func TestNoopNotifier_IsNotificationDispatcher(_ *testing.T) {
	var _ domain.NotificationDispatcher = noopNotifier{}
}

func TestSpyNotifier_CollectsEvents(t *testing.T) {
	spy := &spyNotifier{}
	err := spy.Dispatch(context.Background(), domain.NotificationEvent{OwnerID: 1, Type: "payroll.payslip"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(spy.events) != 1 {
		t.Errorf("expected 1 event, got %d", len(spy.events))
	}
	if spy.events[0].Type != "payroll.payslip" {
		t.Errorf("unexpected event type: %s", spy.events[0].Type)
	}
}

type fakeApprovalChainProvider struct {
	chain *domain.ApprovalChainInfo
}

func (f fakeApprovalChainProvider) GetApprovalChain(_ context.Context, _ int64, _ string) (*domain.ApprovalChainInfo, error) {
	return f.chain, nil
}

func TestTriggerAndApprovePayrollRun_PersistsContractStates(t *testing.T) {
	ctx := context.Background()
	db := testutil.OpenDB(t)

	const (
		businessID     int64 = 9100202
		creatorUserID  int64 = 9200202
		approverUserID int64 = 9300202
		approverEmpID  int64 = 9400202
		periodStart          = "2031-04-01"
		periodEnd            = "2031-04-30"
	)

	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(),
			"DELETE FROM hr_payroll_runs WHERE business_id = ? AND period_start = ? AND period_end = ?",
			businessID, periodStart, periodEnd)
		_, _ = db.ExecContext(context.Background(),
			"DELETE FROM hr_payroll_config WHERE owner_id = ? AND business_id = ?",
			businessID, businessID)
	})

	service := NewService(ctx, moduleinfra.NewRepository(db), fakeApprovalChainProvider{
		chain: &domain.ApprovalChainInfo{
			TotalLevels: 1,
			Levels: []domain.ApprovalLevelInfo{{
				LevelOrder:         1,
				ApproverEmployeeID: approverEmpID,
			}},
		},
	}, fakePermissionProvider{employeeID: approverEmpID})

	if err := service.UpsertPayrollConfig(ctx, businessID, domain.PayrollConfigRecord{
		PeriodType:    domain.PayrollPeriodMonthly,
		PayDayOfMonth: 28,
	}); err != nil {
		t.Fatalf("UpsertPayrollConfig: %v", err)
	}

	run, _, err := service.TriggerPayrollRun(ctx, businessID, creatorUserID, periodStart, periodEnd)
	if err != nil {
		t.Fatalf("TriggerPayrollRun: %v", err)
	}
	if run.Status != "in_review" {
		t.Fatalf("triggered payroll status = %q, want in_review", run.Status)
	}

	persisted, err := service.repo.GetPayrollRun(ctx, businessID, run.ID)
	if err != nil {
		t.Fatalf("GetPayrollRun before approval: %v", err)
	}
	if persisted.Status != "in_review" {
		t.Fatalf("persisted payroll status = %q, want in_review", persisted.Status)
	}

	if err := service.ApprovePayrollRun(ctx, businessID, run.ID, approverUserID); err != nil {
		t.Fatalf("ApprovePayrollRun: %v", err)
	}

	approved, err := service.repo.GetPayrollRun(ctx, businessID, run.ID)
	if err != nil {
		t.Fatalf("GetPayrollRun after approval: %v", err)
	}
	if approved.Status != domain.PayrollStatusApproved {
		t.Errorf("approved payroll status = %q, want %q", approved.Status, domain.PayrollStatusApproved)
	}
	if approved.ApprovedBy != approverUserID {
		t.Errorf("approved_by = %d, want %d", approved.ApprovedBy, approverUserID)
	}
}
