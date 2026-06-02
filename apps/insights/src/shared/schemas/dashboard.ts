import { z } from 'zod';

// insightsKpisSchema validates GET /hub/kpis (hubv1.GetKPIsResponse). Backend
// json tags revenue/order_count/staff_count/customer_count/low_stock_count/
// active_apps are converted to camelCase by the @voxire/api interceptor.
export const insightsKpisSchema = z.object({
  revenue: z.number().optional(),
  orderCount: z.number().optional(),
  staffCount: z.number().optional(),
  customerCount: z.number().optional(),
  lowStockCount: z.number().optional(),
  activeApps: z.number().optional(),
});

// metricCardSchema mirrors the domain.{Money,Count,Pct}Card envelope returned
// by GET /insights/dashboard: each card is { value, delta_pct } (delta_pct ->
// deltaPct after camelCase conversion).
const metricCardSchema = z.object({
  value: z.number().optional(),
  deltaPct: z.number().optional(),
});

// insightsDashboardSummarySchema validates GET /insights/dashboard
// (GetDashboardResponse). The backend returns four metric cards keyed
// revenue/customers/orders/margin, not flat numeric fields.
export const insightsDashboardSummarySchema = z.object({
  revenue: metricCardSchema.optional(),
  customers: metricCardSchema.optional(),
  orders: metricCardSchema.optional(),
  margin: metricCardSchema.optional(),
});

export type InsightsKpis = z.infer<typeof insightsKpisSchema>;
export type InsightsDashboardSummary = z.infer<typeof insightsDashboardSummarySchema>;
