package application

import (
	"context"
	"errors"
	"testing"

	"github.com/voxire/onion/internal/modules/hr_payroll/internal/domain"
)

func TestNoopFinancePoster_DoesNotError(t *testing.T) {
	if err := (noopFinancePoster{}).PostPayrollJournal(context.Background(), domain.PayrollJournalEntry{}); err != nil {
		t.Fatalf("noop finance poster returned error: %v", err)
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
		{"end equal to start", "2026-04-15", "2026-04-15"},
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

func TestTriggerPayrollRun_EqualDatesRejected(t *testing.T) {
	s := &Service{}
	_, _, err := s.TriggerPayrollRun(context.Background(), 1, 1, "2026-05-01", "2026-05-01")
	if !errors.Is(err, domain.ErrInvalidDateRange) {
		t.Errorf("equal dates should return ErrInvalidDateRange, got %v", err)
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
