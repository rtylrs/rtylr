package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
)

type mockRepo struct {
	listResult []domain.PayrollRecord
	shifts     []domain.ShiftSummary
	rates      domain.StaffPayRates
	createID   int64
	listErr    error
	shiftErr   error
	ratesErr   error
	createErr  error
	markErr    error
	createGot  struct {
		regular, overtime, base, ot, net float64
	}
}

func (m *mockRepo) ListPayrollRecords(_ context.Context, _ int64, _ string) ([]domain.PayrollRecord, error) {
	return m.listResult, m.listErr
}
func (m *mockRepo) GetCompletedShiftsForPeriod(_ context.Context, _ int64, _, _ time.Time) ([]domain.ShiftSummary, error) {
	return m.shifts, m.shiftErr
}
func (m *mockRepo) GetStaffPayRates(_ context.Context, _, _ int64) (domain.StaffPayRates, error) {
	return m.rates, m.ratesErr
}
func (m *mockRepo) CreatePayrollRecord(_ context.Context, _, _ int64, _, _ time.Time, _ string, regular, overtime, base, ot, net float64) (int64, error) {
	m.createGot.regular = regular
	m.createGot.overtime = overtime
	m.createGot.base = base
	m.createGot.ot = ot
	m.createGot.net = net
	return m.createID, m.createErr
}
func (m *mockRepo) MarkPayrollPaid(_ context.Context, _, _ int64) error {
	return m.markErr
}

func TestRunPayroll_RejectsInvalidDates(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if _, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{PeriodStart: "bad", PeriodEnd: "2026-04-30"}); err == nil {
		t.Error("expected error for bad start")
	}
	if _, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{PeriodStart: "2026-04-01", PeriodEnd: "bad"}); err == nil {
		t.Error("expected error for bad end")
	}
}

func TestRunPayroll_ComputesHourlyPay(t *testing.T) {
	m := &mockRepo{
		shifts: []domain.ShiftSummary{{StaffID: 1, TotalHours: 50, TotalOvertimeHours: 10}},
		rates:  domain.StaffPayRates{HourlyRate: 10, OvertimeRate: 0}, // OT defaults to 1.5x
	}
	s := &Service{repo: m}
	count, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
		PeriodStart: "2026-04-01",
		PeriodEnd:   "2026-04-30",
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 record, got %d", count)
	}
	// 40 regular × 10 = 400 base, 10 OT × 15 = 150 OT, net 550.
	if m.createGot.base != 400 || m.createGot.ot != 150 || m.createGot.net != 550 {
		t.Errorf("pay computation wrong: %+v", m.createGot)
	}
}

func TestListPayrollRecords_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{listErr: want}}
	if _, err := s.ListPayrollRecords(context.Background(), 1, ""); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestMarkPayrollPaid_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{markErr: want}}
	if err := s.MarkPayrollPaid(context.Background(), 1, 5); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}
