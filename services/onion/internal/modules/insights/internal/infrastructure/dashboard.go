package infrastructure

import (
	"context"
	"time"

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
func (r *Repository) GetDashboard(ctx context.Context, ownerID int64, in DashboardInput) (domain.DashboardSnapshot, error) {
	thisFrom, thisTo, priorFrom, priorTo := mtdRanges(in.Now)

	revThis, err := r.sumPaidInvoicesMinorUnits(ctx, ownerID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	revPrior, err := r.sumPaidInvoicesMinorUnits(ctx, ownerID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}

	ordersThis, err := r.countPosOrders(ctx, ownerID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	ordersPrior, err := r.countPosOrders(ctx, ownerID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}

	custThis, err := r.countCrmCustomersCreated(ctx, ownerID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	custPrior, err := r.countCrmCustomersCreated(ctx, ownerID, priorFrom, priorTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}

	// Margin % = (revenue - COGS) / revenue * 100, where COGS here is the
	// summed expenses over the same window. This is a proxy, not a GAAP
	// margin; the finance module owns the strict calculation.
	expThis, err := r.sumExpensesMinorUnits(ctx, ownerID, thisFrom, thisTo)
	if err != nil {
		return domain.DashboardSnapshot{}, err
	}
	expPrior, err := r.sumExpensesMinorUnits(ctx, ownerID, priorFrom, priorTo)
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

func (r *Repository) sumPaidInvoicesMinorUnits(ctx context.Context, ownerID int64, from, to time.Time) (int64, error) {
	// cross-schema read from finance_invoices — paid invoice revenue powers the dashboard revenue card
	const q = `
		SELECT COALESCE(SUM(total), 0)
		FROM finance_invoices
		WHERE owner_id = ?
		  AND status = 'paid'
		  AND COALESCE(paid_date, issue_date) >= ?
		  AND COALESCE(paid_date, issue_date) <= ?
	`
	row := r.pool.QueryRowContext(ctx, q, ownerID, from.Format("2006-01-02"), to.Format("2006-01-02"))
	var total interface{}
	if err := row.Scan(&total); err != nil {
		logger.Ctx(ctx).Error("insights: sum paid invoices failed", logger.Error(err))
		return 0, err
	}
	return decimalToMinorUnits(total), nil
}

func (r *Repository) countPosOrders(ctx context.Context, ownerID int64, from, to time.Time) (int64, error) {
	// cross-schema read from pos_orders — order volume powers the dashboard orders card
	const q = `
		SELECT COUNT(*)
		FROM pos_orders
		WHERE owner_id = ?
		  AND created_at >= ?
		  AND created_at <= ?
	`
	var n int64
	if err := r.pool.QueryRowContext(ctx, q, ownerID, from, to).Scan(&n); err != nil {
		logger.Ctx(ctx).Error("insights: count pos orders failed", logger.Error(err))
		return 0, err
	}
	return n, nil
}

func (r *Repository) countCrmCustomersCreated(ctx context.Context, ownerID int64, from, to time.Time) (int64, error) {
	// cross-schema read from crm_customers — new customer signups for the dashboard customers card
	const q = `
		SELECT COUNT(*)
		FROM crm_customers
		WHERE owner_id = ?
		  AND created_at >= ?
		  AND created_at <= ?
	`
	var n int64
	if err := r.pool.QueryRowContext(ctx, q, ownerID, from, to).Scan(&n); err != nil {
		logger.Ctx(ctx).Error("insights: count crm customers failed", logger.Error(err))
		return 0, err
	}
	return n, nil
}

func (r *Repository) sumExpensesMinorUnits(ctx context.Context, ownerID int64, from, to time.Time) (int64, error) {
	// cross-schema read from finance_expenses — expense totals feed the margin card
	const q = `
		SELECT COALESCE(SUM(amount), 0)
		FROM finance_expenses
		WHERE owner_id = ?
		  AND expense_date >= ?
		  AND expense_date <= ?
	`
	var total interface{}
	if err := r.pool.QueryRowContext(ctx, q, ownerID, from.Format("2006-01-02"), to.Format("2006-01-02")).Scan(&total); err != nil {
		logger.Ctx(ctx).Error("insights: sum expenses failed", logger.Error(err))
		return 0, err
	}
	return decimalToMinorUnits(total), nil
}
