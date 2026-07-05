import { safe } from '../lib/api';

export interface InsightsSettingsPayload {
  defaultCurrency?: string;
  fiscalYearStart?: string;
  timezone?: string;
  [key: string]: unknown;
}

export interface ScheduledReportPayload {
  templateId?: string;
  frequency?: string;
  cadence?: string;
  channel?: string;
  recipient?: string;
  recipients?: string[];
  isActive?: boolean;
  format?: 'csv' | 'pdf';
  [key: string]: unknown;
}

const toInsightsPeriod = (params?: { period?: string; start?: string; end?: string }) => {
  if (params?.period != null) return { period: params.period };
  if (params?.start || params?.end) return { period: `${params.start ?? ''},${params.end ?? ''}` };
  return undefined;
};

export const insightsService = {
  // Dashboard / KPIs
  getKpis: () => safe.get('/api/hub/dashboard/v1/kpis'),
  getDashboardSummary: () => safe.get('/api/insights/dashboard/v1/summary'),

  // Revenue
  getRevenue: (params?: { period?: string; start?: string; end?: string }) =>
    safe.get('/api/insights/revenue/v1/overview', { params: toInsightsPeriod(params) }),
  getRevenueByApp: (_params?: { period?: string; start?: string; end?: string }) =>
    safe.get('/api/insights/revenue/v1/by-app'),
  getRevenueComparison: (params?: { period?: string; start?: string; end?: string }) =>
    safe.get('/api/insights/revenue/v1/comparison', {
      params: { compare_to: 'previous_period', ...toInsightsPeriod(params) },
    }),

  // Customers
  getCustomerAnalytics: () => safe.get('/api/insights/customer/v1/analytics'),
  getTopCustomers: (params?: { limit?: number }) =>
    safe.get('/api/insights/customer/v1/top', { params }),
  getCustomerChurn: () => safe.get('/api/insights/customer/v1/churn'),
  getCustomerLtv: () => safe.get('/api/insights/customer/v1/ltv'),

  // Products
  getProductPerformance: () => safe.get('/api/insights/product/v1/performance'),
  getBestSellers: (params?: { limit?: number }) =>
    safe.get('/api/insights/product/v1/best-sellers', { params }),
  getSlowMovers: (params?: { limit?: number; days?: number }) =>
    safe.get('/api/insights/product/v1/slow-movers', {
      params: { days: params?.days ?? params?.limit ?? 60 },
    }),
  getMarginAnalysis: () => safe.get('/api/insights/product/v1/margins'),
  getCategoryBreakdown: () => safe.get('/api/insights/product/v1/categories'),
  getErpProducts: () => safe.get('/api/erp/product/v1/products'),

  // Operations
  getOperationsSummary: () => safe.get('/api/insights/operation/v1/summary'),
  getStaffUtilization: () => safe.get('/api/insights/operation/v1/utilization'),
  getShiftCoverage: (params?: { period?: string; start?: string; end?: string }) =>
    safe.get('/api/insights/operation/v1/shifts', { params: toInsightsPeriod(params) }),
  getTimeOffTrends: () => safe.get('/api/insights/operation/v1/time-off'),
  getDeptHeadcount: () => safe.get('/api/insights/operation/v1/headcount'),
  getHrDashboard: () => safe.get('/api/hr/dashboard/v1/dashboard'),

  // Financial
  getProfitLoss: (params?: { start?: string; end?: string }) => {
    const date = params?.end ?? params?.start ?? new Date().toISOString().slice(0, 10);
    const period =
      params?.start && params?.end
        ? `${params.start},${params.end}`
        : (params?.start ?? params?.end ?? 'month');
    return safe.get('/api/finance/report/v1/pl', { params: { period, date } });
  },
  getExpenseBreakdown: (params?: { start?: string; end?: string; period?: string }) =>
    safe.get('/api/insights/financial/v1/expenses', { params: toInsightsPeriod(params) }),
  getInvoiceAging: () => safe.get('/api/insights/financial/v1/invoice-aging'),
  getCashFlow: (params?: { period?: string; start?: string; end?: string }) =>
    safe.get('/api/insights/financial/v1/cash-flow', { params: toInsightsPeriod(params) }),

  // Reports
  getReportTemplates: () => safe.get('/api/insights/report/v1/templates'),
  generateReport: (templateId: string, params?: Record<string, string>) => {
    if (templateId !== 'revenue') throw new Error(`Unsupported report template: ${templateId}`);
    return safe.post('/api/insights/report/v1/reports/revenue/generate', params);
  },
  exportReport: (templateId: string, format: 'csv' | 'pdf', params?: Record<string, string>) => {
    if (templateId !== 'revenue') throw new Error(`Unsupported report template: ${templateId}`);
    if (format !== 'csv') throw new Error(`Unsupported export format: ${format}`);
    return safe.get<{ csv: string; filename: string; contentType: string }>(
      '/api/insights/report/v1/reports/revenue/export',
      { params }
    );
  },

  // Settings
  getInsightsSettings: () => safe.get('/api/insights/setting/v1/settings'),
  updateInsightsSettings: (data: InsightsSettingsPayload) =>
    safe.put('/api/insights/setting/v1/settings', {
      settings: {
        default_period: 'this_month',
        currency: data.defaultCurrency ?? 'USD',
        compare_previous: false,
        dashboard_widgets: [],
        favorite_reports: [],
      },
    }),
  getScheduledReports: () => safe.get('/api/insights/report/v1/scheduled'),
  updateScheduledReport: (id: string, data: Partial<ScheduledReportPayload>) =>
    safe.put(`/api/insights/report/v1/scheduled/${id}`, {
      template_slug: data.templateId,
      frequency: data.frequency ?? data.cadence,
      channel: data.channel ?? 'email',
      recipient_email: Array.isArray(data.recipients) ? data.recipients[0] : data.recipient,
      is_active: data.isActive ?? true,
    }),
  createScheduledReport: (data: ScheduledReportPayload) =>
    safe.post('/api/insights/report/v1/scheduled', {
      template_slug: data.templateId,
      frequency: data.frequency ?? data.cadence,
      channel: data.channel ?? 'email',
      recipient_email: Array.isArray(data.recipients) ? data.recipients[0] : data.recipient,
      is_active: data.isActive ?? true,
    }),
  deleteScheduledReport: (id: string) => safe.delete(`/api/insights/report/v1/scheduled/${id}`),
};
