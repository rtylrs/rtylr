package application

import (
	"context"
	"fmt"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
)

// GetSalesSummary returns the CRM deal cards for the overview, cached like
// the headline cards.
func (s *Service) GetSalesSummary(ctx context.Context, businessID int64) (domain.SalesSummary, error) {
	key, err := s.insightsCacheKey(ctx, businessID, "dashboard-sales")
	if err != nil {
		return domain.SalesSummary{}, err
	}
	return cacher.Of[domain.SalesSummary]().GetOrComputeSingle(ctx, key+"|domain=v2", func(ctx context.Context) (domain.SalesSummary, error) {
		return s.readSalesSummary(ctx, businessID, time.Now())
	}, cacher.WithExpiry(ctx, dashboardCacheTTL))
}

func (s *Service) readSalesSummary(ctx context.Context, businessID int64, now time.Time) (domain.SalesSummary, error) {
	sales := s.dashboard.CRMSales
	if sales == nil {
		return domain.SalesSummary{}, errDashboardReadersUnavailable
	}
	thisFrom, thisTo, priorFrom, priorTo := dashboardMTDRanges(now)

	openDeals, openValue, err := sales.OpenDealPipeline(ctx, businessID)
	if err != nil {
		return domain.SalesSummary{}, fmt.Errorf("read open deal pipeline: %w", err)
	}
	_, wonThis, err := sales.WonDealsBetween(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.SalesSummary{}, fmt.Errorf("read current won deal value: %w", err)
	}
	_, wonPrior, err := sales.WonDealsBetween(ctx, businessID, priorFrom, priorTo)
	if err != nil {
		return domain.SalesSummary{}, fmt.Errorf("read previous won deal value: %w", err)
	}
	recentWon, _, err := sales.WonDealsBetween(ctx, businessID, priorFrom, thisTo)
	if err != nil {
		return domain.SalesSummary{}, fmt.Errorf("read recent won deal count: %w", err)
	}
	contracts, mrr, mrrAtMonthStart, err := sales.RecurringRevenue(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.SalesSummary{}, fmt.Errorf("read recurring revenue: %w", err)
	}
	return domain.SalesSummary{
		OpenPipelineValue: openValue,
		OpenDeals:         openDeals,
		Won:               domain.MoneyCard{Value: wonThis, DeltaPct: dashboardDeltaPct(wonThis, wonPrior)},
		RecentWonDeals:    recentWon,
		MRR:               domain.MoneyCard{Value: mrr, DeltaPct: dashboardDeltaPct(mrr, mrrAtMonthStart)},
		ActiveContracts:   contracts,
	}, nil
}
