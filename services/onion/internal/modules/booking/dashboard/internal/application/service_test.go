package application

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/booking/dashboard/internal/domain"
	"github.com/voxire/onion/internal/platform/money"
)

type mockDashboardRepo struct {
	todayCount     int
	upcomingCount  int
	totalMonth     int
	cancelledMonth int
	revenue        money.Amount
	err            error
	// per-method override
	errToday    error
	errUpcoming error
	errMonth    error
	errRevenue  error
	delay       time.Duration
	activeCalls atomic.Int64
	maxActive   atomic.Int64
}

func (m *mockDashboardRepo) GetTodayCount(_ context.Context, _ int64, _, _ time.Time) (int, error) {
	done := m.trackCall()
	defer done()

	if m.errToday != nil {
		return 0, m.errToday
	}
	return m.todayCount, m.err
}

func (m *mockDashboardRepo) GetUpcomingCount(_ context.Context, _ int64, _ time.Time) (int, error) {
	done := m.trackCall()
	defer done()

	if m.errUpcoming != nil {
		return 0, m.errUpcoming
	}
	return m.upcomingCount, m.err
}

func (m *mockDashboardRepo) GetMonthCounts(_ context.Context, _ int64, _ time.Time) (int, int, error) {
	done := m.trackCall()
	defer done()

	if m.errMonth != nil {
		return 0, 0, m.errMonth
	}
	return m.totalMonth, m.cancelledMonth, m.err
}

func (m *mockDashboardRepo) GetMonthRevenue(_ context.Context, _ int64, _ time.Time) (money.Amount, error) {
	done := m.trackCall()
	defer done()

	if m.errRevenue != nil {
		return 0, m.errRevenue
	}
	return m.revenue, m.err
}

func (m *mockDashboardRepo) trackCall() func() {
	if m.delay <= 0 {
		return func() {}
	}

	active := m.activeCalls.Add(1)
	for {
		maxActive := m.maxActive.Load()
		if active <= maxActive || m.maxActive.CompareAndSwap(maxActive, active) {
			break
		}
	}

	time.Sleep(m.delay)

	return func() {
		m.activeCalls.Add(-1)
	}
}

func newDashboardService(r domain.Repository) *Service {
	return &Service{repo: r}
}

func TestFetchDashboard_HappyPath(t *testing.T) {
	repo := &mockDashboardRepo{todayCount: 5, upcomingCount: 10, totalMonth: 20, cancelledMonth: 4, revenue: money.Amount(150000)}
	svc := newDashboardService(repo)
	data, err := svc.fetchDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.TodayCount != 5 {
		t.Errorf("expected TodayCount=5, got %d", data.TodayCount)
	}
	if data.UpcomingCount != 10 {
		t.Errorf("expected UpcomingCount=10, got %d", data.UpcomingCount)
	}
}

func TestFetchDashboardRunsIndependentQueriesConcurrently(t *testing.T) {
	repo := &mockDashboardRepo{
		todayCount:     5,
		upcomingCount:  10,
		totalMonth:     20,
		cancelledMonth: 4,
		revenue:        money.Amount(150000),
		delay:          20 * time.Millisecond,
	}
	svc := newDashboardService(repo)

	if _, err := svc.fetchDashboard(context.Background(), 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if maxActive := repo.maxActive.Load(); maxActive < 2 {
		t.Fatalf("expected concurrent repo calls, max active calls = %d", maxActive)
	}
}

func TestFetchDashboard_CancellationRateCalculated(t *testing.T) {
	repo := &mockDashboardRepo{totalMonth: 10, cancelledMonth: 2, revenue: money.Amount(50000)}
	svc := newDashboardService(repo)
	data, err := svc.fetchDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := 20.0
	if data.CancellationRate != expected {
		t.Errorf("expected CancellationRate=%.1f, got %.1f", expected, data.CancellationRate)
	}
}

func TestFetchDashboard_ZeroTotalMonthNoCancellationRate(t *testing.T) {
	repo := &mockDashboardRepo{totalMonth: 0, cancelledMonth: 0, revenue: money.Amount(0)}
	svc := newDashboardService(repo)
	data, err := svc.fetchDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.CancellationRate != 0 {
		t.Errorf("expected CancellationRate=0 when no bookings, got %.2f", data.CancellationRate)
	}
}

func TestFetchDashboard_RevenuePopulated(t *testing.T) {
	repo := &mockDashboardRepo{revenue: money.Amount(999999)}
	svc := newDashboardService(repo)
	data, err := svc.fetchDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.Revenue != money.Amount(999999) {
		t.Errorf("expected Revenue=9999.99, got %s", data.Revenue.String())
	}
}

func TestFetchDashboard_PropagatesErrorOnTodayCount(t *testing.T) {
	repo := &mockDashboardRepo{errToday: errors.New("db error")}
	svc := newDashboardService(repo)
	_, err := svc.fetchDashboard(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from GetTodayCount")
	}
}

func TestFetchDashboard_PropagatesErrorOnUpcomingCount(t *testing.T) {
	repo := &mockDashboardRepo{errUpcoming: errors.New("upstream failure")}
	svc := newDashboardService(repo)
	_, err := svc.fetchDashboard(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from GetUpcomingCount")
	}
}

func TestFetchDashboard_PropagatesErrorOnMonthCounts(t *testing.T) {
	repo := &mockDashboardRepo{errMonth: errors.New("month query failed")}
	svc := newDashboardService(repo)
	_, err := svc.fetchDashboard(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from GetMonthCounts")
	}
}

func TestFetchDashboard_PropagatesErrorOnRevenue(t *testing.T) {
	repo := &mockDashboardRepo{errRevenue: errors.New("revenue query failed")}
	svc := newDashboardService(repo)
	_, err := svc.fetchDashboard(context.Background(), 1)
	if err == nil {
		t.Fatal("expected error from GetMonthRevenue")
	}
}

func TestFetchDashboard_AllZeros(t *testing.T) {
	repo := &mockDashboardRepo{}
	svc := newDashboardService(repo)
	data, err := svc.fetchDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.TodayCount != 0 || data.UpcomingCount != 0 || data.Revenue != money.Amount(0) || data.CancellationRate != 0 {
		t.Errorf("expected all zeros, got %+v", data)
	}
}

func TestFetchDashboard_FullCancellationRate(t *testing.T) {
	repo := &mockDashboardRepo{totalMonth: 5, cancelledMonth: 5, revenue: money.Amount(0)}
	svc := newDashboardService(repo)
	data, err := svc.fetchDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.CancellationRate != 100.0 {
		t.Errorf("expected 100%% cancellation rate, got %.1f", data.CancellationRate)
	}
}

func TestFetchDashboard_PartialCancellationRate(t *testing.T) {
	repo := &mockDashboardRepo{totalMonth: 4, cancelledMonth: 1, revenue: money.Amount(20000)}
	svc := newDashboardService(repo)
	data, err := svc.fetchDashboard(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if data.CancellationRate != 25.0 {
		t.Errorf("expected 25%% cancellation rate, got %.1f", data.CancellationRate)
	}
}
