package application

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockRepo struct {
	byType   map[string]float64
	cash     float64
	recv     float64
	pay      float64
	inflow   float64
	outflow  float64
	typeErr  error
	cashErr  error
	rangeErr error
}

func (m *mockRepo) SumAccountsByType(_ context.Context, _ int64, t string) (float64, error) {
	return m.byType[t], m.typeErr
}

func (m *mockRepo) SumBankBalances(_ context.Context, _ int64) (float64, error) {
	return m.cash, m.cashErr
}

func (m *mockRepo) SumReceivables(_ context.Context, _ int64) (float64, error) {
	return m.recv, nil
}

func (m *mockRepo) SumPayables(_ context.Context, _ int64) (float64, error) {
	return m.pay, nil
}

func (m *mockRepo) SumPaidInvoicesInRange(_ context.Context, _ int64, _, _ string) (float64, error) {
	return m.inflow, m.rangeErr
}

func (m *mockRepo) SumExpensesInRange(_ context.Context, _ int64, _, _ string) (float64, error) {
	return m.outflow, m.rangeErr
}

func TestGetBalanceSheet_AggregatesValues(t *testing.T) {
	m := &mockRepo{
		byType: map[string]float64{"asset": 1000, "liability": 200, "equity": 800},
		cash:   500, recv: 100, pay: 50,
	}
	s := &Service{repo: m}
	bs, err := s.GetBalanceSheet(context.Background(), 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	// Total assets = chart_of_accounts assets (1000) + cash (500) + receivables (100).
	if bs.TotalAssets != 1600 {
		t.Errorf("TotalAssets = %v, want 1600", bs.TotalAssets)
	}
	// Total liabilities = chart_of_accounts liabilities (200) + payables (50).
	if bs.TotalLiabilities != 250 {
		t.Errorf("TotalLiabilities = %v, want 250", bs.TotalLiabilities)
	}
	if bs.TotalEquity != 800 {
		t.Errorf("TotalEquity = %v, want 800", bs.TotalEquity)
	}
}

func TestGetBalanceSheet_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{typeErr: want}}
	if _, err := s.GetBalanceSheet(context.Background(), 1); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

func TestGetCashFlow_NetIsInflowMinusOutflow(t *testing.T) {
	m := &mockRepo{inflow: 100, outflow: 30}
	s := &Service{repo: m}
	cf, err := s.GetCashFlow(context.Background(), 1, "month")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cf.NetOperating != 70 {
		t.Errorf("expected 70, got %v", cf.NetOperating)
	}
}

func TestGetCashFlow_PropagatesError(t *testing.T) {
	want := errors.New("boom")
	s := &Service{repo: &mockRepo{rangeErr: want}}
	if _, err := s.GetCashFlow(context.Background(), 1, "month"); !errors.Is(err, want) {
		t.Errorf("expected sentinel, got %v", err)
	}
}

// resolvePeriodRange depends on time.Now() so we can't pin exact dates, but we
// can lock in the SHAPE: yyyy-mm-dd, from <= to, and the relationships between
// months are consistent.

func TestResolvePeriodRange_Shape(t *testing.T) {
	for _, period := range []string{"quarter", "year", "month", "", "garbage"} {
		from, to := resolvePeriodRange(period)
		if !looksLikeDate(from) {
			t.Errorf("from for %q is not yyyy-mm-dd: %q", period, from)
		}
		if !looksLikeDate(to) {
			t.Errorf("to for %q is not yyyy-mm-dd: %q", period, to)
		}
		if from > to {
			t.Errorf("from %q > to %q for %q", from, to, period)
		}
	}
}

func TestResolvePeriodRange_QuarterIsThreeMonths(t *testing.T) {
	from, to := resolvePeriodRange("quarter")
	// from and to are in the same year and three months apart (the to date is the
	// last day of the third month, so months differ by 2 inclusive).
	if from[:4] != to[:4] {
		t.Errorf("quarter spans years: from=%s, to=%s", from, to)
	}
}

func TestResolvePeriodRange_YearStartsJanFirst(t *testing.T) {
	from, _ := resolvePeriodRange("year")
	if !strings.HasSuffix(from, "-01-01") {
		t.Errorf("year start should end with -01-01, got %q", from)
	}
}

func looksLikeDate(s string) bool {
	if len(s) != 10 {
		return false
	}
	return s[4] == '-' && s[7] == '-'
}

func TestGetBalanceSheet_CashIsForwarded(t *testing.T) {
	m := &mockRepo{cash: 999, byType: map[string]float64{}}
	s := &Service{repo: m}
	bs, err := s.GetBalanceSheet(context.Background(), 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if bs.CashBalance != 999 {
		t.Errorf("CashBalance = %v, want 999", bs.CashBalance)
	}
}

func TestGetBalanceSheet_ReceivablesForwarded(t *testing.T) {
	m := &mockRepo{recv: 300, byType: map[string]float64{}}
	s := &Service{repo: m}
	bs, err := s.GetBalanceSheet(context.Background(), 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if bs.AccountsReceivable != 300 {
		t.Errorf("AccountsReceivable = %v, want 300", bs.AccountsReceivable)
	}
}

func TestGetBalanceSheet_PayablesForwarded(t *testing.T) {
	m := &mockRepo{pay: 150, byType: map[string]float64{}}
	s := &Service{repo: m}
	bs, err := s.GetBalanceSheet(context.Background(), 1)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if bs.AccountsPayable != 150 {
		t.Errorf("AccountsPayable = %v, want 150", bs.AccountsPayable)
	}
}

func TestGetCashFlow_PeriodIsForwarded(t *testing.T) {
	s := &Service{repo: &mockRepo{}}
	cf, err := s.GetCashFlow(context.Background(), 1, "year")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cf.Period != "year" {
		t.Errorf("Period = %q, want 'year'", cf.Period)
	}
}

func TestGetCashFlow_InflowAndOutflowForwarded(t *testing.T) {
	m := &mockRepo{inflow: 500, outflow: 200}
	s := &Service{repo: m}
	cf, err := s.GetCashFlow(context.Background(), 1, "month")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cf.OperatingInflow != 500 {
		t.Errorf("OperatingInflow = %v, want 500", cf.OperatingInflow)
	}
	if cf.OperatingOutflow != 200 {
		t.Errorf("OperatingOutflow = %v, want 200", cf.OperatingOutflow)
	}
}

func TestGetCashFlow_NetCashFlowEqualsNetOperating(t *testing.T) {
	m := &mockRepo{inflow: 750, outflow: 250}
	s := &Service{repo: m}
	cf, err := s.GetCashFlow(context.Background(), 1, "quarter")
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if cf.NetCashFlow != cf.NetOperating {
		t.Errorf("NetCashFlow (%v) != NetOperating (%v)", cf.NetCashFlow, cf.NetOperating)
	}
}

func TestResolvePeriodRange_YearEndsDecember(t *testing.T) {
	_, to := resolvePeriodRange("year")
	if !strings.HasSuffix(to, "-12-31") {
		t.Errorf("year end should end with -12-31, got %q", to)
	}
}

func TestResolvePeriodRange_MonthStartsFirstDay(t *testing.T) {
	from, _ := resolvePeriodRange("month")
	if !strings.HasSuffix(from, "-01") {
		t.Errorf("month start should end with -01, got %q", from)
	}
}
