// Query key factories for @tanstack/react-query used in the insights app.
// Centralizing query keys prevents typos and makes cache invalidation
// explicit. All keys follow the pattern ['scope', 'resource', params].

export const insightsKeys = {
  all: ['insights'] as const,

  dashboard: () => [...insightsKeys.all, 'dashboard'] as const,

  revenue: () => [...insightsKeys.all, 'revenue'] as const,
  revenueSeries: (params: { from: string; to: string; granularity?: string }) =>
    [...insightsKeys.revenue(), 'series', params] as const,
  revenueByApp: (params: { from: string; to: string }) =>
    [...insightsKeys.revenue(), 'by-app', params] as const,
  revenueComparison: (params: { from: string; to: string }) =>
    [...insightsKeys.revenue(), 'comparison', params] as const,

  customers: () => [...insightsKeys.all, 'customers'] as const,
  customerKpi: () => [...insightsKeys.customers(), 'kpi'] as const,
  topCustomers: (limit?: number) => [...insightsKeys.customers(), 'top', { limit }] as const,
  customerLtv: () => [...insightsKeys.customers(), 'ltv'] as const,
  customerChurn: () => [...insightsKeys.customers(), 'churn'] as const,

  products: () => [...insightsKeys.all, 'products'] as const,
  productKpi: () => [...insightsKeys.products(), 'kpi'] as const,
  bestSellers: (limit?: number) => [...insightsKeys.products(), 'best-sellers', { limit }] as const,
  productMargins: () => [...insightsKeys.products(), 'margins'] as const,
  slowMovers: (limit?: number) => [...insightsKeys.products(), 'slow-movers', { limit }] as const,

  financial: () => [...insightsKeys.all, 'financial'] as const,
  cashFlow: (params: { from: string; to: string }) =>
    [...insightsKeys.financial(), 'cash-flow', params] as const,
  expensesByCategory: (params: { from: string; to: string }) =>
    [...insightsKeys.financial(), 'expenses', params] as const,
  invoiceAging: () => [...insightsKeys.financial(), 'invoice-aging'] as const,

  operations: () => [...insightsKeys.all, 'operations'] as const,
  operationsKpi: () => [...insightsKeys.operations(), 'kpi'] as const,
  headcount: () => [...insightsKeys.operations(), 'headcount'] as const,
  shiftMetrics: () => [...insightsKeys.operations(), 'shifts'] as const,
  timeOff: () => [...insightsKeys.operations(), 'time-off'] as const,

  reports: () => [...insightsKeys.all, 'reports'] as const,
  reportTemplates: () => [...insightsKeys.reports(), 'templates'] as const,
  scheduledReports: () => [...insightsKeys.reports(), 'scheduled'] as const,

  settings: () => [...insightsKeys.all, 'settings'] as const,
} as const;
