package application

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/payroll/internal/domain"
)

type mockRepo struct {
	listResult    []domain.PayrollRecord
	shifts        []domain.ShiftSummary
	rates         domain.StaffPayRates
	otMultiplier  float64
	otConfigured  bool
	otErr         error
	createID      int64
	listErr       error
	shiftErr      error
	ratesErr      error
	createErr     error
	markErr       error
	multiplierAsk int64
	createGot     struct {
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

func (m *mockRepo) GetStaffPayRatesByIDs(_ context.Context, _ int64, staffIDs []int64) (map[int64]domain.StaffPayRates, error) {
	out := make(map[int64]domain.StaffPayRates, len(staffIDs))
	// Mirror the "no rate row -> skip this staff" behaviour: when the test
	// injects a rates error, return no mappings so every staff is skipped.
	if m.ratesErr != nil {
		return out, nil
	}
	for _, id := range staffIDs {
		out[id] = m.rates
	}
	return out, nil
}

func (m *mockRepo) GetDefaultOvertimeMultiplier(_ context.Context, _ int64) (float64, bool, error) {
	m.multiplierAsk++
	return m.otMultiplier, m.otConfigured, m.otErr
}

func (m *mockRepo) CreatePayrollRecords(_ context.Context, _ int64, records []domain.PayrollRecordInput) (int, error) {
	if len(records) > 0 {
		first := records[0]
		m.createGot.regular = first.RegularHours
		m.createGot.overtime = first.OvertimeHours
		m.createGot.base = first.BasePay
		m.createGot.ot = first.OvertimePay
		m.createGot.net = first.NetPay
	}
	return len(records), m.createErr
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

func TestMarkPayrollPaid_HappyPath(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	if err := s.MarkPayrollPaid(context.Background(), 1, 5); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestRunPayroll_NoShiftsReturnsZero(t *testing.T) {
	m := &mockRepo{shifts: []domain.ShiftSummary{}}
	s := &Service{repo: m}
	count, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
		PeriodStart: "2026-04-01",
		PeriodEnd:   "2026-04-30",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 records for no shifts, got %d", count)
	}
}

func TestRunPayroll_SkipsStaffWhenRatesError(t *testing.T) {
	m := &mockRepo{
		shifts:   []domain.ShiftSummary{{StaffID: 1, TotalHours: 8}},
		ratesErr: errors.New("no rates"),
	}
	s := &Service{repo: m}
	count, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
		PeriodStart: "2026-04-01",
		PeriodEnd:   "2026-04-30",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 due to rates error, got %d", count)
	}
}

func TestRunPayroll_ComputesSalaryBasedPay(t *testing.T) {
	m := &mockRepo{
		shifts: []domain.ShiftSummary{{StaffID: 2, TotalHours: 0, TotalOvertimeHours: 0}},
		rates:  domain.StaffPayRates{Salary: 36500}, // 36500/365 = 100/day
	}
	s := &Service{repo: m}
	count, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
		PeriodStart: "2026-04-01",
		PeriodEnd:   "2026-04-30", // 29 days
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 record, got %d", count)
	}
}

func TestRunPayroll_OvertimeUsesExplicitRate(t *testing.T) {
	m := &mockRepo{
		shifts: []domain.ShiftSummary{{StaffID: 3, TotalHours: 50, TotalOvertimeHours: 10}},
		rates:  domain.StaffPayRates{HourlyRate: 10, OvertimeRate: 20}, // explicit OT rate
	}
	s := &Service{repo: m}
	_, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
		PeriodStart: "2026-04-01",
		PeriodEnd:   "2026-04-30",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 40 regular × 10 = 400, 10 OT × 20 = 200, net 600
	if m.createGot.base != 400 || m.createGot.ot != 200 || m.createGot.net != 600 {
		t.Errorf("pay with explicit OT rate wrong: %+v", m.createGot)
	}
}

func TestRunPayroll_ShiftsError(t *testing.T) {
	want := errors.New("shift error")
	m := &mockRepo{shiftErr: want}
	s := &Service{repo: m}
	if _, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
		PeriodStart: "2026-04-01",
		PeriodEnd:   "2026-04-30",
	}); !errors.Is(err, want) {
		t.Errorf("expected shift error, got %v", err)
	}
}

func TestRunPayroll_RegularHoursNotNegative(t *testing.T) {
	// OT > Total would give negative regular hours; service clamps to 0.
	m := &mockRepo{
		shifts: []domain.ShiftSummary{{StaffID: 4, TotalHours: 5, TotalOvertimeHours: 10}},
		rates:  domain.StaffPayRates{HourlyRate: 10},
	}
	s := &Service{repo: m}
	_, _ = s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
		PeriodStart: "2026-04-01",
		PeriodEnd:   "2026-04-30",
	})
	if m.createGot.regular < 0 {
		t.Errorf("regular hours should be clamped to >= 0, got %f", m.createGot.regular)
	}
}

func TestListPayrollRecords_ReturnsRecords(t *testing.T) {
	records := []domain.PayrollRecord{{ID: 1}, {ID: 2}}
	s := &Service{repo: &mockRepo{listResult: records}}
	got, err := s.ListPayrollRecords(context.Background(), 1, "2026-04")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 2 {
		t.Errorf("expected 2 records, got %d", len(got))
	}
}

func TestRunPayroll_JanuarySalaryProrationIsInclusive(t *testing.T) {
	// $36,500 annual salary over all of January pays 31 days at $100, not
	// the 30 days an exclusive date subtraction produced (#758).
	m := &mockRepo{
		shifts: []domain.ShiftSummary{{StaffID: 1}},
		rates:  domain.StaffPayRates{Salary: 36500},
	}
	s := &Service{repo: m}
	if _, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
		PeriodStart: "2026-01-01",
		PeriodEnd:   "2026-01-31",
	}); err != nil {
		t.Fatalf("RunPayroll: %v", err)
	}
	if m.createGot.base != 3100.00 {
		t.Errorf("base = %f, want 3100.00", m.createGot.base)
	}
}

func TestRunPayroll_LeapYearFebruaryProration(t *testing.T) {
	m := &mockRepo{
		shifts: []domain.ShiftSummary{{StaffID: 1}},
		rates:  domain.StaffPayRates{Salary: 36600},
	}
	s := &Service{repo: m}
	if _, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
		PeriodStart: "2024-02-01",
		PeriodEnd:   "2024-02-29",
	}); err != nil {
		t.Fatalf("RunPayroll: %v", err)
	}
	want := math.Round(36600*29/366*100) / 100
	if m.createGot.base != want {
		t.Errorf("base = %f, want %f", m.createGot.base, want)
	}
}

func TestRunPayroll_OvertimeUsesConfiguredMultiplier(t *testing.T) {
	tests := []struct {
		name            string
		overtimeRate    float64
		multiplier      float64
		configured      bool
		wantOvertimePay float64
	}{
		{name: "explicit staff rate wins", overtimeRate: 30, multiplier: 2.0, configured: true, wantOvertimePay: 300},
		{name: "configured multiplier applies", overtimeRate: 0, multiplier: 2.0, configured: true, wantOvertimePay: 400},
		{name: "legacy default without config", overtimeRate: 0, multiplier: 0, configured: false, wantOvertimePay: 300},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockRepo{
				shifts:       []domain.ShiftSummary{{StaffID: 1, TotalHours: 10, TotalOvertimeHours: 10}},
				rates:        domain.StaffPayRates{HourlyRate: 20, OvertimeRate: tt.overtimeRate},
				otMultiplier: tt.multiplier,
				otConfigured: tt.configured,
			}
			s := &Service{repo: m}
			if _, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
				PeriodStart: "2026-04-01",
				PeriodEnd:   "2026-04-30",
			}); err != nil {
				t.Fatalf("RunPayroll: %v", err)
			}
			if m.createGot.ot != tt.wantOvertimePay {
				t.Errorf("overtime pay = %f, want %f", m.createGot.ot, tt.wantOvertimePay)
			}
		})
	}
}

func TestRunPayroll_RoundsMoneyToCents(t *testing.T) {
	// 7.5h x $33.333 = 249.9975 must persist as 250.00, not truncate.
	m := &mockRepo{
		shifts:       []domain.ShiftSummary{{StaffID: 1, TotalHours: 7.5}},
		rates:        domain.StaffPayRates{HourlyRate: 33.333},
		otConfigured: true,
	}
	s := &Service{repo: m}
	if _, err := s.RunPayroll(context.Background(), 1, domain.PayrollRunInput{
		PeriodStart: "2026-04-01",
		PeriodEnd:   "2026-04-30",
	}); err != nil {
		t.Fatalf("RunPayroll: %v", err)
	}
	if m.createGot.base != 250.00 {
		t.Errorf("base = %f, want 250.00", m.createGot.base)
	}
	if m.createGot.net != 250.00 {
		t.Errorf("net = %f, want 250.00", m.createGot.net)
	}
}
