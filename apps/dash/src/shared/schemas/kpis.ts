import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}, z.number());
const numberSeries = z.array(finiteNumber).catch([]);

export const dashKpisSchema = z
  .object({
    revenue: finiteNumber,
    orderCount: finiteNumber.optional(),
    order_count: finiteNumber.optional(),
    staffCount: finiteNumber.optional(),
    staff_count: finiteNumber.optional(),
    customerCount: finiteNumber.optional(),
    customer_count: finiteNumber.optional(),
    lowStockCount: finiteNumber.optional(),
    low_stock_count: finiteNumber.optional(),
    activeApps: finiteNumber.optional(),
    active_apps: finiteNumber.optional(),
    revenueToday: finiteNumber.optional(),
    revenue_today: finiteNumber.optional(),
    revenueTrendPct: finiteNumber.optional(),
    revenue_trend_pct: finiteNumber.optional(),
    ordersToday: finiteNumber.optional(),
    orders_today: finiteNumber.optional(),
    ordersTrendPct: finiteNumber.optional(),
    orders_trend_pct: finiteNumber.optional(),
    headcount: finiteNumber,
    activeStaff: finiteNumber.optional(),
    active_staff: finiteNumber.optional(),
    lowStockItems: finiteNumber.optional(),
    low_stock_items: finiteNumber.optional(),
    openInvoices: finiteNumber.optional(),
    open_invoices: finiteNumber.optional(),
    newCustomers: finiteNumber.optional(),
    new_customers: finiteNumber.optional(),
    apiHealthPct: finiteNumber.optional(),
    api_health_pct: finiteNumber.optional(),
    revenueSeries: numberSeries.optional(),
    revenue_series: numberSeries.optional(),
    ordersSeries: numberSeries.optional(),
    orders_series: numberSeries.optional(),
  })
  .transform((raw) => ({
    revenue: raw.revenue,
    orderCount: raw.orderCount ?? raw.order_count ?? 0,
    staffCount: raw.staffCount ?? raw.staff_count ?? 0,
    customerCount: raw.customerCount ?? raw.customer_count ?? 0,
    lowStockCount: raw.lowStockCount ?? raw.low_stock_count ?? 0,
    activeApps: raw.activeApps ?? raw.active_apps ?? 0,
    revenueToday: raw.revenueToday ?? raw.revenue_today ?? 0,
    revenueTrendPct: raw.revenueTrendPct ?? raw.revenue_trend_pct ?? 0,
    ordersToday: raw.ordersToday ?? raw.orders_today ?? 0,
    ordersTrendPct: raw.ordersTrendPct ?? raw.orders_trend_pct ?? 0,
    headcount: raw.headcount,
    activeStaff: raw.activeStaff ?? raw.active_staff ?? 0,
    lowStockItems: raw.lowStockItems ?? raw.low_stock_items ?? 0,
    openInvoices: raw.openInvoices ?? raw.open_invoices ?? 0,
    newCustomers: raw.newCustomers ?? raw.new_customers ?? 0,
    apiHealthPct: raw.apiHealthPct ?? raw.api_health_pct ?? 0,
    revenueSeries: raw.revenueSeries ?? raw.revenue_series ?? [],
    ordersSeries: raw.ordersSeries ?? raw.orders_series ?? [],
  }));

export type HubKpis = z.infer<typeof dashKpisSchema>;
