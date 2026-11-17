package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
)

type dashboardFlowHoursFake struct {
	logged  map[dateRange]domain.FlowLoggedTime
	periods []dateRange
	err     error
}

func (f *dashboardFlowHoursFake) FlowLoggedTime(_ context.Context, _ int64, from, to time.Time) (domain.FlowLoggedTime, error) {
	f.periods = append(f.periods, dateRange{from, to})
	return f.logged[dateRange{from, to}], f.err
}

var (
	hoursNow      = time.Date(2026, time.October, 15, 12, 0, 0, 0, time.UTC)
	hoursThisMTD  = dateRange{time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC), hoursNow}
	hoursPriorMTD = dateRange{time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, time.September, 15, 12, 0, 0, 0, time.UTC)}
)

func TestReadHoursSummaryComparesWithTheSameSpanLastMonth(t *testing.T) {
	flow := &dashboardFlowHoursFake{logged: map[dateRange]domain.FlowLoggedTime{
		hoursThisMTD: {TotalMinutes: 2_400, Projects: []domain.FlowProjectTime{
			{ProjectID: 1, Name: "Website", Visibility: "public", Status: "active", Minutes: 2_070},
			{ProjectID: 2, Name: "Payroll audit", Visibility: "private", Status: "active", Minutes: 300},
			{ProjectID: 3, Name: "Legacy", Visibility: "public", Status: "archived", Minutes: 30},
		}},
		hoursPriorMTD: {TotalMinutes: 1_600},
	}}
	service := NewService(nil, DashboardReaders{Flow: flow})

	got, err := service.readHoursSummary(context.Background(), 42, hoursNow)
	if err != nil {
		t.Fatalf("readHoursSummary: %v", err)
	}
	if got.Logged.Value != 2_400 || got.Logged.DeltaPct != 50 || got.PriorMinutes != 1_600 {
		t.Errorf("logged = %+v with %d prior minutes, want 2400 up 50%% on 1600", got.Logged, got.PriorMinutes)
	}
	// Private and archived projects count in the total but are not named.
	want := []domain.ProjectHours{{ProjectID: 1, Name: "Website", Minutes: 2_070}}
	if len(got.Projects) != len(want) || got.Projects[0] != want[0] {
		t.Errorf("Projects = %+v, want %+v", got.Projects, want)
	}
	wantPeriods := []dateRange{hoursThisMTD, hoursPriorMTD}
	if len(flow.periods) != len(wantPeriods) {
		t.Fatalf("windows = %#v, want %#v", flow.periods, wantPeriods)
	}
	for i := range wantPeriods {
		if flow.periods[i] != wantPeriods[i] {
			t.Errorf("window %d = %#v, want %#v", i, flow.periods[i], wantPeriods[i])
		}
	}
}

// Time logged only last month still reports it, so the client can keep the
// card with a fall to zero.
func TestReadHoursSummaryWithTimeOnlyLastMonth(t *testing.T) {
	flow := &dashboardFlowHoursFake{logged: map[dateRange]domain.FlowLoggedTime{hoursPriorMTD: {TotalMinutes: 600}}}
	got, err := NewService(nil, DashboardReaders{Flow: flow}).readHoursSummary(context.Background(), 42, hoursNow)
	if err != nil {
		t.Fatalf("readHoursSummary: %v", err)
	}
	if got.Logged.Value != 0 || got.Logged.DeltaPct != -100 || got.PriorMinutes != 600 || len(got.Projects) != 0 {
		t.Errorf("summary = %+v, want no minutes, down 100%% on 600 and no projects", got)
	}
}

func TestReadHoursSummaryReportsReaderFailures(t *testing.T) {
	if _, err := NewService(nil, DashboardReaders{}).readHoursSummary(context.Background(), 42, hoursNow); !errors.Is(err, errDashboardReadersUnavailable) {
		t.Errorf("without a Flow reader: error = %v, want %v", err, errDashboardReadersUnavailable)
	}
	boom := errors.New("flow unavailable")
	service := NewService(nil, DashboardReaders{Flow: &dashboardFlowHoursFake{err: boom}})
	if _, err := service.readHoursSummary(context.Background(), 42, hoursNow); !errors.Is(err, boom) {
		t.Errorf("error = %v, want it to wrap %v", err, boom)
	}
}
