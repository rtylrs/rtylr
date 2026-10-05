package application

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// dashboardCacheTTL keeps the expensive cross-schema rollup behind a
// 5-minute cache, per the task spec. Mutations in downstream modules
// (invoices, orders, customers) do not busts this key — acceptable for
// an analytics surface where 5-minute freshness is the product requirement.
const dashboardCacheTTL = 5 * time.Minute

var errDashboardReadersUnavailable = errors.New("insights dashboard readers are not configured")

// GetDashboard returns the four headline cards for the given owner.
// Result is cached per-owner under `insights:{businessID}:dashboard`.
func (s *Service) GetDashboard(ctx context.Context, businessID int64) (domain.DashboardSnapshot, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "dashboard")
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	return cacher.Of[domain.DashboardSnapshot]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (domain.DashboardSnapshot, error) {
		return s.readDashboard(ctx, businessID, time.Now())
	}, cacher.WithExpiry(ctx, dashboardCacheTTL))
}

func (s *Service) readDashboard(ctx context.Context, businessID int64, now time.Time) (domain.DashboardSnapshot, error) {
	if s.dashboard.Finance == nil || s.dashboard.POS == nil || s.dashboard.CRM == nil {
		return domain.DashboardSnapshot{}, errDashboardReadersUnavailable
	}
	thisFrom, thisTo, priorFrom, priorTo := dashboardMTDRanges(now)

	revThis, err := s.dashboard.Finance.InsightsPaidInvoiceTotal(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, fmt.Errorf("read current paid invoice total: %w", err)
	}
	revPrior, err := s.dashboard.Finance.InsightsPaidInvoiceTotal(ctx, businessID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, fmt.Errorf("read previous paid invoice total: %w", err)
	}
	ordersThis, err := s.dashboard.POS.CountOrdersCreatedBetween(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, fmt.Errorf("read current order count: %w", err)
	}
	ordersPrior, err := s.dashboard.POS.CountOrdersCreatedBetween(ctx, businessID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, fmt.Errorf("read previous order count: %w", err)
	}
	customersThis, err := s.dashboard.CRM.CountCustomersCreatedBetween(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, fmt.Errorf("read current customer count: %w", err)
	}
	customersPrior, err := s.dashboard.CRM.CountCustomersCreatedBetween(ctx, businessID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, fmt.Errorf("read previous customer count: %w", err)
	}
	expensesThis, err := s.dashboard.Finance.InsightsExpenseTotal(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, fmt.Errorf("read current expense total: %w", err)
	}
	expensesPrior, err := s.dashboard.Finance.InsightsExpenseTotal(ctx, businessID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, fmt.Errorf("read previous expense total: %w", err)
	}
	marginThis := dashboardMarginPct(revThis, expensesThis)
	marginPrior := dashboardMarginPct(revPrior, expensesPrior)
	return domain.DashboardSnapshot{
		Revenue:   domain.MoneyCard{Value: revThis, DeltaPct: dashboardDeltaPct(revThis, revPrior)},
		Customers: domain.CountCard{Value: customersThis, DeltaPct: dashboardDeltaPct(customersThis, customersPrior)},
		Orders:    domain.CountCard{Value: ordersThis, DeltaPct: dashboardDeltaPct(ordersThis, ordersPrior)},
		Margin:    domain.PctCard{Value: marginThis, DeltaPct: dashboardDeltaPctFloat(marginThis, marginPrior)},
	}, nil
}

func dashboardMTDRanges(now time.Time) (thisFrom, thisTo, priorFrom, priorTo time.Time) {
	thisFrom = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	thisTo = now
	priorFrom = thisFrom.AddDate(0, -1, 0)
	priorTo = priorFrom.Add(now.Sub(thisFrom))
	return thisFrom, thisTo, priorFrom, priorTo
}

func dashboardDeltaPct(current, prior int64) float64 {
	if prior == 0 {
		return 0
	}
	return float64(current-prior) / float64(prior) * 100
}

func dashboardDeltaPctFloat(current, prior float64) float64 {
	if prior == 0 {
		return 0
	}
	return (current - prior) / prior * 100
}

func dashboardMarginPct(revenueMinor, expenseMinor int64) float64 {
	if revenueMinor <= 0 {
		return 0
	}
	return float64(revenueMinor-expenseMinor) / float64(revenueMinor) * 100
}
