// Domain constants for the insights app.

export const DEFAULT_PAGE_SIZE = 25;
export const STALE_TIME_MS = 5 * 60 * 1000;

export const CHART_COLORS = [
  '#6366f1',
  '#10b981',
  '#f59e0b',
  '#ef4444',
  '#3b82f6',
  '#8b5cf6',
  '#ec4899',
  '#14b8a6',
] as const;
export type ChartColor = (typeof CHART_COLORS)[number];

export const METRIC_CATEGORIES = ['revenue', 'orders', 'customers', 'staff'] as const;
export type MetricCategory = (typeof METRIC_CATEGORIES)[number];
export const METRIC_CATEGORY_LABELS: Record<MetricCategory, string> = {
  revenue: 'Revenue',
  orders: 'Orders',
  customers: 'Customers',
  staff: 'Staff',
};

export const REVENUE_GRANULARITIES = ['daily', 'weekly', 'monthly'] as const;
export type RevenueGranularity = (typeof REVENUE_GRANULARITIES)[number];
export const REVENUE_GRANULARITY_LABELS: Record<RevenueGranularity, string> = {
  daily: 'Daily',
  weekly: 'Weekly',
  monthly: 'Monthly',
};

export const MIN_TREND_POINTS = 2;
export const TOP_LIST_LIMIT = 10;
