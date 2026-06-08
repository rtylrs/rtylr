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
  recipients?: string[];
  format?: 'csv' | 'pdf';
  [key: string]: unknown;
}

export const insightsService = {
  // Dashboard / KPIs
  getKpis: () => safe.get('/hub/kpis'),
  getDashboardSummary: () => safe.get('/insights/dashboard'),

  // Revenue
  getRevenue: (params?: { period?: string; start?: string; end?: string }) =>
    safe.get('/insights/revenue', { params }),
  getRevenueByApp: (params?: { period?: string; start?: string; end?: string }) =>
    safe.get('/insights/revenue/by-app', { params }),
  getRevenueComparison: (params?: { period?: string }) =>
    safe.get('/insights/revenue/comparison', { params }),

  // Customers
  getCustomerAnalytics: () => safe.get('/insights/customers'),
  getTopCustomers: (params?: { limit?: number }) => safe.get('/insights/customers/top', { params }),
  getCustomerChurn: () => safe.get('/insights/customers/churn'),
  getCustomerLtv: () => safe.get('/insights/customers/ltv'),

  // Products
  getProductPerformance: () => safe.get('/insights/products'),
  getBestSellers: (params?: { limit?: number }) =>
    safe.get('/insights/products/best-sellers', { params }),
  getSlowMovers: (params?: { limit?: number }) =>
    safe.get('/insights/products/slow-movers', { params }),
  getMarginAnalysis: () => safe.get('/insights/products/margins'),
  getCategoryBreakdown: () => safe.get('/insights/products/categories'),
  getErpProducts: () => safe.get('/erp/products'),

  // Operations
  getOperationsSummary: () => safe.get('/insights/operations'),
  getStaffUtilization: () => safe.get('/insights/operations/utilization'),
  getShiftCoverage: () => safe.get('/insights/operations/shifts'),
  getTimeOffTrends: () => safe.get('/insights/operations/time-off'),
  getDeptHeadcount: () => safe.get('/insights/operations/headcount'),
  getHrDashboard: () => safe.get('/hr/dashboard'),

  // Financial
  getProfitLoss: (params?: { start?: string; end?: string }) => safe.get('/finance/pl', { params }),
  getExpenseBreakdown: (params?: { start?: string; end?: string }) =>
    safe.get('/insights/financial/expenses', { params }),
  getInvoiceAging: () => safe.get('/insights/financial/invoice-aging'),
  getCashFlow: (params?: { period?: string }) =>
    safe.get('/insights/financial/cash-flow', { params }),

  // Reports
  getReportTemplates: () => safe.get('/insights/reports/templates'),
  generateReport: (templateId: string, params?: Record<string, string>) => {
    if (templateId !== 'revenue') throw new Error(`Unsupported report template: ${templateId}`);
    return safe.post('/insights/reports/generate/revenue', params);
  },
  exportReport: (templateId: string, format: 'csv' | 'pdf', params?: Record<string, string>) => {
    if (templateId !== 'revenue') throw new Error(`Unsupported report template: ${templateId}`);
    return safe.get('/insights/reports/export/revenue', {
      params: { format, ...params },
      responseType: 'blob',
    });
  },

  // Settings
  getInsightsSettings: () => safe.get('/insights/settings'),
  updateInsightsSettings: (data: InsightsSettingsPayload) => safe.put('/insights/settings', data),
  getScheduledReports: () => safe.get('/insights/settings/scheduled-reports'),
  updateScheduledReport: (id: string, data: Partial<ScheduledReportPayload>) =>
    safe.put(`/insights/settings/scheduled-reports/${id}`, data),
  createScheduledReport: (data: ScheduledReportPayload) =>
    safe.post('/insights/settings/scheduled-reports', data),
  deleteScheduledReport: (id: string) => safe.delete(`/insights/settings/scheduled-reports/${id}`),
};
