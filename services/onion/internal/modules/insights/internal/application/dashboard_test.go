package application

import (
	"context"
	"testing"
	"time"
)

type dashboardFinanceFake struct {
	invoicePeriods []dateRange
	expensePeriods []dateRange
}

func (f *dashboardFinanceFake) InsightsPaidInvoiceTotal(_ context.Context, _ int64, from, to time.Time) (int64, error) {
	f.invoicePeriods = append(f.invoicePeriods, dateRange{from, to})
	if from.Month() == time.October {
		return 50000, nil
	}
	return 40000, nil
}

func (f *dashboardFinanceFake) InsightsExpenseTotal(_ context.Context, _ int64, from, to time.Time) (int64, error) {
	f.expensePeriods = append(f.expensePeriods, dateRange{from, to})
	if from.Month() == time.October {
		return 10000, nil
	}
	return 12000, nil
}

type dashboardPOSFake struct{ periods []dateRange }

func (f *dashboardPOSFake) CountOrdersCreatedBetween(_ context.Context, _ int64, from, to time.Time) (int64, error) {
	f.periods = append(f.periods, dateRange{from, to})
	if from.Month() == time.October {
		return 4, nil
	}
	return 2, nil
}

type dashboardCRMFake struct{ periods []dateRange }

func (f *dashboardCRMFake) CountCustomersCreatedBetween(_ context.Context, _ int64, from, to time.Time) (int64, error) {
	f.periods = append(f.periods, dateRange{from, to})
	if from.Month() == time.October {
		return 9, nil
	}
	return 5, nil
}

type dateRange struct{ from, to time.Time }

func TestReadDashboardUsesOwnerReadersAndPreservesReportWindows(t *testing.T) {
	finance := &dashboardFinanceFake{}
	pos := &dashboardPOSFake{}
	crm := &dashboardCRMFake{}
	service := NewService(nil, DashboardReaders{Finance: finance, POS: pos, CRM: crm})
	now := time.Date(2026, time.October, 15, 12, 0, 0, 0, time.UTC)

	got, err := service.readDashboard(context.Background(), 42, now)
	if err != nil {
		t.Fatalf("readDashboard: %v", err)
	}
	if got.Revenue.Value != 50000 || got.Revenue.DeltaPct != 25 {
		t.Errorf("Revenue = %+v", got.Revenue)
	}
	if got.Customers.Value != 9 || got.Customers.DeltaPct != 80 {
		t.Errorf("Customers = %+v", got.Customers)
	}
	if got.Orders.Value != 4 || got.Orders.DeltaPct != 100 {
		t.Errorf("Orders = %+v", got.Orders)
	}
	if got.Margin.Value != 80 || got.Margin.DeltaPct < 14.28 || got.Margin.DeltaPct > 14.29 {
		t.Errorf("Margin = %+v", got.Margin)
	}
	current := dateRange{time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC), now}
	previous := dateRange{time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, time.September, 15, 12, 0, 0, 0, time.UTC)}
	for name, ranges := range map[string][]dateRange{
		"invoice": finance.invoicePeriods,
		"expense": finance.expensePeriods,
		"POS":     pos.periods,
		"CRM":     crm.periods,
	} {
		if len(ranges) != 2 || ranges[0] != current || ranges[1] != previous {
			t.Errorf("%s ranges = %#v, want current and previous MTD %#v %#v", name, ranges, current, previous)
		}
	}
}

func TestReadDashboardRequiresEveryOwnerReader(t *testing.T) {
	service := NewService(nil, DashboardReaders{})
	if _, err := service.readDashboard(context.Background(), 42, time.Now()); err != errDashboardReadersUnavailable {
		t.Fatalf("readDashboard error = %v, want %v", err, errDashboardReadersUnavailable)
	}
}
