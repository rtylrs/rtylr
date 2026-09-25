package application

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/voxire/onion/internal/modules/invoices/internal/domain"
	"github.com/voxire/onion/internal/platform/cache/cacher"
	cacheKeys "github.com/voxire/onion/internal/platform/cache/keys"
)

func (s *Service) GetDashboard(ctx context.Context, businessID int64) (domain.Dashboard, error) {
	key := cacheKeys.Scoped(invoicesScope, businessID, "dashboard")

	return cacher.Of[domain.Dashboard]().GetOrComputeSingle(ctx, key,
		func(ctx context.Context) (domain.Dashboard, error) {
			return s.fetchDashboard(ctx, businessID)
		},
		cacher.WithExpiry(ctx, time.Minute),
	)
}

func (s *Service) fetchDashboard(ctx context.Context, businessID int64) (domain.Dashboard, error) {
	now := time.Now()
	mtdFrom := fmt.Sprintf("%d-%02d-01", now.Year(), now.Month())
	mtdTo := now.Format("2006-01-02")

	var (
		totalReceivable  float64
		overdueAmount    float64
		totalExpensesMTD float64
		revenueMTD       float64
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		var err error
		totalReceivable, err = s.repo.SumTotalReceivable(groupCtx, businessID)
		return err
	})
	group.Go(func() error {
		var err error
		overdueAmount, err = s.repo.SumOverdueAmount(groupCtx, businessID)
		return err
	})
	group.Go(func() error {
		var err error
		totalExpensesMTD, err = s.repo.SumExpenses(groupCtx, businessID, mtdFrom, mtdTo)
		return err
	})
	group.Go(func() error {
		var err error
		revenueMTD, err = s.repo.SumPaidInvoices(groupCtx, businessID, mtdFrom, mtdTo)
		return err
	})
	if err := group.Wait(); err != nil {
		return domain.Dashboard{}, err
	}

	netProfitMTD := revenueMTD - totalExpensesMTD

	return domain.Dashboard{
		TotalReceivable:  totalReceivable,
		OverdueAmount:    overdueAmount,
		TotalExpensesMTD: totalExpensesMTD,
		TotalRevenueMTD:  revenueMTD,
		NetProfitMTD:     netProfitMTD,
	}, nil
}

func (s *Service) GetTaxReport(ctx context.Context, businessID int64, period, date string) (domain.TaxReport, error) {
	dateFrom, dateTo := resolvePeriodRange(period, date)

	taxCollected, err := s.repo.SumTaxCollected(ctx, businessID, dateFrom, dateTo)
	if err != nil {
		return domain.TaxReport{}, err
	}

	taxOwed, err := s.repo.SumTaxOwed(ctx, businessID, dateFrom, dateTo)
	if err != nil {
		return domain.TaxReport{}, err
	}

	return domain.TaxReport{
		Period:       period,
		TaxCollected: taxCollected,
		TaxOwed:      taxOwed,
	}, nil
}

// resolvePeriodRange converts a period keyword + anchor date into a date range.
// date should be in YYYY-MM-DD format; if empty, today is used.
func resolvePeriodRange(period, date string) (string, string) {
	var anchor time.Time

	if date != "" {
		if t, err := time.Parse("2006-01-02", date); err == nil {
			anchor = t
		}
	}

	if anchor.IsZero() {
		anchor = time.Now()
	}

	switch period {
	case "month":
		from := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, -1)

		return from.Format("2006-01-02"), to.Format("2006-01-02")

	case "quarter":
		quarter := (int(anchor.Month()) - 1) / 3
		fromMonth := time.Month(quarter*3 + 1)
		from := time.Date(anchor.Year(), fromMonth, 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 3, -1)

		return from.Format("2006-01-02"), to.Format("2006-01-02")

	case "year":
		from := time.Date(anchor.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		to := time.Date(anchor.Year(), 12, 31, 0, 0, 0, 0, time.UTC)

		return from.Format("2006-01-02"), to.Format("2006-01-02")

	default:
		// default to current month
		from := time.Date(anchor.Year(), anchor.Month(), 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, -1)

		return from.Format("2006-01-02"), to.Format("2006-01-02")
	}
}

// CentsToFloat converts integer cents to standard float amount.
func CentsToFloat(cents int64) float64 {
	return float64(cents) / 100.0
}

// ListInvoicesByCustomer returns one page of a customer's invoices. Keep this
// compatibility wrapper for the legacy read surface; the RPC path uses the
// filtered variant above.
func (s *Service) ListInvoicesByCustomer(ctx context.Context, businessID, customerID, page, pageSize int64) ([]domain.Invoice, bool, error) {
	return s.ListInvoicesForCustomer(ctx, businessID, customerID, page, pageSize, "", "", "", "")
}

// CountInvoicesByCustomer reports how many invoices carry the customer link.
func (s *Service) CountInvoicesByCustomer(ctx context.Context, businessID, customerID int64) (int64, error) {
	return s.repo.CountInvoicesByCustomer(ctx, businessID, customerID)
}
