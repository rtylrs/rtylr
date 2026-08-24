import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}, z.number());

/**
 * hub.dashboard.v1.GetKPIsResponse, and nothing else.
 *
 * This used to declare eighteen further fields - revenueToday, ordersToday,
 * newCustomers, revenueSeries, headcount, apiHealthPct and more - none of which
 * the contract carries. Every one was .optional() with a transform default of 0
 * or [], so a consumer reading them got a plausible zero rather than a type
 * error, and the Analytics screen showed "$0.00" and "0" regardless of business
 * activity (rtylr#947).
 *
 * They are gone so tsc catches the next drift instead of zod silently absorbing
 * it. Add a field here only when the contract has it.
 *
 * openInvoices was the last holdout, kept only because KpiInvoices read it. That
 * card now sources the count from insights.financial.v1 getInvoiceAging
 * (rtylr#1001), so this schema finally declares nothing the contract lacks.
 *
 * On scope: revenue and orderCount are today-scoped server-side (onion's
 * kpis.sql: SumRevenueToday, CountOrdersToday, both filtered to
 * created_at >= CURDATE()). staffCount, customerCount and lowStockCount are not
 * time-scoped at all, so a label implying a period would be wrong for those.
 */
export const dashKpisSchema = z
  .object({
    revenue: finiteNumber,
    orderCount: finiteNumber.optional(),
    staffCount: finiteNumber.optional(),
    customerCount: finiteNumber.optional(),
    lowStockCount: finiteNumber.optional(),
    activeApps: finiteNumber.optional(),
  })
  .transform((raw) => ({
    revenue: raw.revenue,
    orderCount: raw.orderCount ?? 0,
    staffCount: raw.staffCount ?? 0,
    customerCount: raw.customerCount ?? 0,
    lowStockCount: raw.lowStockCount ?? 0,
    activeApps: raw.activeApps ?? 0,
  }));

export type HubKpis = z.infer<typeof dashKpisSchema>;
