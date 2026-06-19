package infrastructure

import (
	"context"

	"github.com/voxire/onion/internal/platform/logger"
)

// RevenueByAppRow is an internal row type: one entry per "app surface" that
// generated revenue over the reporting window (MTD by default).
type RevenueByAppRow struct {
	App              string
	AmountMinorUnits int64
}

// GetRevenueByApp returns MTD revenue aggregated by source: POS sales,
// Finance one-off invoices, and ERP purchase-order fulfilment proxy (total
// of paid, supplier-sourced inventory transactions). The frontend uses
// this to render the revenue-split donut.
//
// The ERP row is intentionally not computed from purchase orders because
// those are outflows; instead it is left at zero until a sellable ERP
// channel exists. CRM row is always zero because CRM itself does not
// capture payments — it surfaces crm_customers as a dimension only.
func (r *Repository) GetRevenueByApp(ctx context.Context, businessID int64) ([]RevenueByAppRow, error) {
	thisFrom, thisTo, _, _ := mtdRanges(nowFn())

	// cross-schema read from pos_orders — POS-tracked paid order revenue for the split donut
	const posQ = `
		SELECT COALESCE(SUM(total), 0)
		FROM pos_orders
		WHERE business_id = ?
		  AND payment_status = 'paid'
		  AND created_at >= ?
		  AND created_at <= ?
	`
	var posTotal interface{}
	if err := r.pool.QueryRowContext(ctx, posQ, businessID, thisFrom, thisTo).Scan(&posTotal); err != nil {
		logger.Ctx(ctx).Error("insights: revenue by app (pos) failed", logger.Error(err))
		return nil, err
	}

	// cross-schema read from finance_invoices — one-off invoice revenue for the split donut
	const finQ = `
		SELECT COALESCE(SUM(total), 0)
		FROM finance_invoices
		WHERE business_id = ?
		  AND status = 'paid'
		  AND COALESCE(paid_date, issue_date) >= ?
		  AND COALESCE(paid_date, issue_date) <= ?
	`
	var finTotal interface{}
	if err := r.pool.QueryRowContext(ctx, finQ, businessID, thisFrom.Format("2006-01-02"), thisTo.Format("2006-01-02")).Scan(&finTotal); err != nil {
		logger.Ctx(ctx).Error("insights: revenue by app (finance) failed", logger.Error(err))
		return nil, err
	}

	return []RevenueByAppRow{
		{App: "pos", AmountMinorUnits: decimalToMinorUnits(posTotal)},
		{App: "finance", AmountMinorUnits: decimalToMinorUnits(finTotal)},
		{App: "erp", AmountMinorUnits: 0},
		{App: "crm", AmountMinorUnits: 0},
	}, nil
}
