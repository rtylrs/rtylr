package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

type wonDeals struct{ deals, value int64 }

type dashboardCRMSalesFake struct {
	openDeals, openValue int64
	won                  map[dateRange]wonDeals
	periods              []dateRange
	err                  error
}

func (f *dashboardCRMSalesFake) OpenDealPipeline(context.Context, int64) (int64, int64, error) {
	return f.openDeals, f.openValue, f.err
}

func (f *dashboardCRMSalesFake) WonDealsBetween(_ context.Context, _ int64, from, to time.Time) (int64, int64, error) {
	f.periods = append(f.periods, dateRange{from, to})
	w := f.won[dateRange{from, to}]
	return w.deals, w.value, f.err
}

var (
	salesNow      = time.Date(2026, time.October, 15, 12, 0, 0, 0, time.UTC)
	salesThisMTD  = dateRange{time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC), salesNow}
	salesPriorMTD = dateRange{time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, time.September, 15, 12, 0, 0, 0, time.UTC)}
	salesRecent   = dateRange{time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), salesNow}
)

func TestReadSalesSummaryComparesWonValueWithTheSameSpanLastMonth(t *testing.T) {
	sales := &dashboardCRMSalesFake{
		openDeals: 12, openValue: 4_500_000,
		won: map[dateRange]wonDeals{
			salesThisMTD:  {deals: 2, value: 300_000},
			salesPriorMTD: {deals: 1, value: 200_000},
			salesRecent:   {deals: 4, value: 900_000},
		},
	}
	service := NewService(nil, DashboardReaders{CRMSales: sales})

	got, err := service.readSalesSummary(context.Background(), 42, salesNow)
	if err != nil {
		t.Fatalf("readSalesSummary: %v", err)
	}
	if got.OpenDeals != 12 || got.OpenPipelineValue != 4_500_000 {
		t.Errorf("open pipeline = %d deals, %d minor; want 12, 4500000", got.OpenDeals, got.OpenPipelineValue)
	}
	if got.Won.Value != 300_000 || got.Won.DeltaPct != 50 {
		t.Errorf("Won = %+v, want 300000 up 50%%", got.Won)
	}
	if got.RecentWonDeals != 4 {
		t.Errorf("RecentWonDeals = %d, want the 4 won since the start of last month", got.RecentWonDeals)
	}
	want := []dateRange{salesThisMTD, salesPriorMTD, salesRecent}
	if len(sales.periods) != len(want) {
		t.Fatalf("won windows = %#v, want %#v", sales.periods, want)
	}
	for i := range want {
		if sales.periods[i] != want[i] {
			t.Errorf("won window %d = %#v, want %#v", i, sales.periods[i], want[i])
		}
	}
}

func TestReadSalesSummaryWithoutDealsIsAllZero(t *testing.T) {
	service := NewService(nil, DashboardReaders{CRMSales: &dashboardCRMSalesFake{}})

	got, err := service.readSalesSummary(context.Background(), 42, salesNow)
	if err != nil {
		t.Fatalf("readSalesSummary: %v", err)
	}
	if got.OpenDeals != 0 || got.OpenPipelineValue != 0 || got.Won.Value != 0 || got.Won.DeltaPct != 0 || got.RecentWonDeals != 0 {
		t.Errorf("summary = %+v, want zeros with no delta against an empty month", got)
	}
}

func TestReadSalesSummaryReportsReaderFailures(t *testing.T) {
	if _, err := NewService(nil, DashboardReaders{}).readSalesSummary(context.Background(), 42, salesNow); !errors.Is(err, errDashboardReadersUnavailable) {
		t.Errorf("without a CRM sales reader: error = %v, want %v", err, errDashboardReadersUnavailable)
	}
	boom := errors.New("crm unavailable")
	service := NewService(nil, DashboardReaders{CRMSales: &dashboardCRMSalesFake{err: boom}})
	if _, err := service.readSalesSummary(context.Background(), 42, salesNow); !errors.Is(err, boom) {
		t.Errorf("error = %v, want it to wrap %v", err, boom)
	}
}
