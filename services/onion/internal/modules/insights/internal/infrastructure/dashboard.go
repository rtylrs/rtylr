package infrastructure

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/platform/logger"
)

// DashboardInput captures the MTD / prior-MTD window used to build the
// four headline cards. Computed in the service layer, never from HTTP.
type DashboardInput struct {
	Now time.Time
}

// GetDashboard returns the {revenue, customers, orders, margin} snapshot
// for the given owner. Every underlying query is a cross-schema read; a
// fresh tenant with no invoices/orders/customers gets back all zeros,
// not an error, per onion/CLAUDE.md §8 fallback rules.
func (r *Repository) GetDashboard(ctx context.Context, businessID int64, in DashboardInput) (domain.DashboardSnapshot, error) {
	thisFrom, thisTo, priorFrom, priorTo := mtdRanges(in.Now)

	revThis, err := r.sumPaidInvoicesMinorUnits(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	revPrior, err := r.sumPaidInvoicesMinorUnits(ctx, businessID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}

	ordersThis, err := r.countPosOrders(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	ordersPrior, err := r.countPosOrders(ctx, businessID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}

	custThis, err := r.countCrmCustomersCreated(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	custPrior, err := r.countCrmCustomersCreated(ctx, businessID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}

	// Margin % = (revenue - COGS) / revenue * 100, where COGS here is the
	// summed expenses over the same window. This is a proxy, not a GAAP
	// margin; the finance module owns the strict calculation.
	expThis, err := r.sumExpensesMinorUnits(ctx, businessID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	expPrior, err := r.sumExpensesMinorUnits(ctx, businessID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	marginThis := marginPct(revThis, expThis)
	marginPrior := marginPct(revPrior, expPrior)

	return domain.DashboardSnapshot{
		Revenue: domain.MoneyCard{
			Value:    revThis,
			DeltaPct: deltaPct(revThis, revPrior),
		},
		Customers: domain.CountCard{
			Value:    custThis,
			DeltaPct: deltaPct(custThis, custPrior),
		},
		Orders: domain.CountCard{
			Value:    ordersThis,
			DeltaPct: deltaPct(ordersThis, ordersPrior),
		},
		Margin: domain.PctCard{
			Value:    marginPct(revThis, expThis),
			DeltaPct: deltaPctFloat(marginThis, marginPrior),
		},
	}, nil
}

func marginPct(revenueMinor, expenseMinor int64) float64 {
	if revenueMinor <= 0 {
		return 0
	}
	return (float64(revenueMinor-expenseMinor) / float64(revenueMinor)) * 100
}

func (r *Repository) sumPaidInvoicesMinorUnits(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	// cross-schema read from finance_invoices — paid invoice revenue powers the dashboard revenue card
	q, args, err := sq.
		Select("COALESCE(SUM(total), 0)").
		From("finance_invoices").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.Eq{"status": "paid"}).
		Where(sq.GtOrEq{"COALESCE(paid_date, issue_date)": from.Format("2006-01-02")}).
		Where(sq.LtOrEq{"COALESCE(paid_date, issue_date)": to.Format("2006-01-02")}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build sum paid invoices query failed", logger.Error(err))
		return 0, err
	}
	var total interface{}
	if err := r.pool.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("insights: sum paid invoices failed", logger.Error(err))
		return 0, err
	}
	return decimalToMinorUnits(total), nil
}

func (r *Repository) countPosOrders(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	// cross-schema read from pos_orders — order volume powers the dashboard orders card
	q, args, err := sq.
		Select("COUNT(*)").
		From("pos_orders").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.GtOrEq{"created_at": from}).
		Where(sq.LtOrEq{"created_at": to}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build count pos orders query failed", logger.Error(err))
		return 0, err
	}
	var n int64
	if err := r.pool.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
		logger.Ctx(ctx).Error("insights: count pos orders failed", logger.Error(err))
		return 0, err
	}
	return n, nil
}

func (r *Repository) countCrmCustomersCreated(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	// cross-schema read from crm_customers — new customer signups for the dashboard customers card
	q, args, err := sq.
		Select("COUNT(*)").
		From("crm_customers").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.GtOrEq{"created_at": from}).
		Where(sq.LtOrEq{"created_at": to}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build count crm customers query failed", logger.Error(err))
		return 0, err
	}
	var n int64
	if err := r.pool.QueryRowContext(ctx, q, args...).Scan(&n); err != nil {
		logger.Ctx(ctx).Error("insights: count crm customers failed", logger.Error(err))
		return 0, err
	}
	return n, nil
}

func (r *Repository) sumExpensesMinorUnits(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	// cross-schema read from finance_expenses — expense totals feed the margin card
	q, args, err := sq.
		Select("COALESCE(SUM(amount), 0)").
		From("finance_expenses").
		Where(sq.Eq{"business_id": businessID}).
		Where(sq.GtOrEq{"expense_date": from.Format("2006-01-02")}).
		Where(sq.LtOrEq{"expense_date": to.Format("2006-01-02")}).
		PlaceholderFormat(sq.Question).
		ToSql()
	if err != nil {
		logger.Ctx(ctx).Error("insights: build sum expenses query failed", logger.Error(err))
		return 0, err
	}
	var total interface{}
	if err := r.pool.QueryRowContext(ctx, q, args...).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("insights: sum expenses failed", logger.Error(err))
		return 0, err
	}
	return decimalToMinorUnits(total), nil
}
