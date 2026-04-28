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
