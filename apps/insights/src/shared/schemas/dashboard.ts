import { z } from 'zod';

// insightsKpisSchema validates GET /hub/kpis (hubv1.GetKPIsResponse). Backend
// json tags revenue/order_count/staff_count/customer_count/low_stock_count/
// active_apps are converted to camelCase by the @voxire/api interceptor. All
// six are proto3 scalar fields, so the wire always carries a number (zero when
// unset) — they are required here so a dropped field fails loudly.
export const insightsKpisSchema = z.object({
  revenue: z.number(),
  orderCount: z.number(),
  staffCount: z.number(),
  customerCount: z.number(),
  lowStockCount: z.number(),
  activeApps: z.number(),
});

// metricCardSchema mirrors the domain.{Money,Count,Pct}Card envelope returned
// by GET /insights/dashboard: each card is { value, delta_pct } (delta_pct ->
// deltaPct after camelCase conversion). Both fields are non-pointer Go numbers
// and are always serialized, so both are required.
const metricCardSchema = z.object({
  value: z.number(),
  deltaPct: z.number(),
});

// insightsDashboardSummarySchema validates GET /insights/dashboard
// (GetDashboardResponse). The backend always returns all four metric cards
// keyed revenue/customers/orders/margin (non-pointer struct fields), so each
// card is required.
export const insightsDashboardSummarySchema = z.object({
  revenue: metricCardSchema,
  customers: metricCardSchema,
  orders: metricCardSchema,
  margin: metricCardSchema,
});

export type InsightsKpis = z.infer<typeof insightsKpisSchema>;
export type InsightsDashboardSummary = z.infer<typeof insightsDashboardSummarySchema>;
