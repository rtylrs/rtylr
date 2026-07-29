package infrastructure

import (
	"context"
	"database/sql"
	"time"

	"github.com/voxire/onion/internal/modules/insights/internal/domain"
	"github.com/voxire/onion/internal/modules/insights/internal/infrastructure/persistence/sqlc"
	"github.com/voxire/onion/internal/platform/logger"
)

type DashboardInput struct {
	Now time.Time
}

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
		Revenue:   domain.MoneyCard{Value: revThis, DeltaPct: deltaPct(revThis, revPrior)},
		Customers: domain.CountCard{Value: custThis, DeltaPct: deltaPct(custThis, custPrior)},
		Orders:    domain.CountCard{Value: ordersThis, DeltaPct: deltaPct(ordersThis, ordersPrior)},
		Margin:    domain.PctCard{Value: marginPct(revThis, expThis), DeltaPct: deltaPctFloat(marginThis, marginPrior)},
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
	total, err := r.queries.SumPaidInvoices(ctx, sqlc.SumPaidInvoicesParams{
		BusinessID: businessID,
		PaidDate:   sql.NullTime{Time: from, Valid: true},
		PaidDate_2: sql.NullTime{Time: to, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: sum paid invoices failed", logger.Error(err))
		return 0, err
	}
	return decimalToMinorUnits(total), nil
}

func (r *Repository) countPosOrders(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	// cross-schema read from pos_orders — order volume powers the dashboard orders card
	n, err := r.queries.CountPosOrders(ctx, sqlc.CountPosOrdersParams{
		BusinessID:  businessID,
		CreatedAt:   sql.NullTime{Time: from, Valid: true},
		CreatedAt_2: sql.NullTime{Time: to, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: count pos orders failed", logger.Error(err))
		return 0, err
	}
	return n, nil
}

func (r *Repository) countCrmCustomersCreated(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	// cross-schema read from crm_customers — new customer signups for the dashboard customers card
	n, err := r.queries.CountCrmCustomersCreated(ctx, sqlc.CountCrmCustomersCreatedParams{
		BusinessID:  businessID,
		CreatedAt:   sql.NullTime{Time: from, Valid: true},
		CreatedAt_2: sql.NullTime{Time: to, Valid: true},
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: count crm customers failed", logger.Error(err))
		return 0, err
	}
	return n, nil
}

func (r *Repository) sumExpensesMinorUnits(ctx context.Context, businessID int64, from, to time.Time) (int64, error) {
	// cross-schema read from finance_expenses — expense totals feed the margin card
	total, err := r.queries.SumExpenses(ctx, sqlc.SumExpensesParams{
		BusinessID:    businessID,
		ExpenseDate:   from,
		ExpenseDate_2: to,
	})
	if err != nil {
		logger.Ctx(ctx).Error("insights: sum expenses failed", logger.Error(err))
		return 0, err
	}
	return decimalToMinorUnits(total), nil
}
