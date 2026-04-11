import api from '../lib/api';

export const insightsService = {
  // Dashboard / KPIs
  getKpis: () => api.get('/hub/kpis'),
  getDashboardSummary: () => api.get('/insights/dashboard'),

  // Revenue
  getRevenue: (params?: { period?: string; start?: string; end?: string }) =>
    api.get('/insights/revenue', { params }),
  getRevenueByApp: (params?: { period?: string; start?: string; end?: string }) =>
    api.get('/insights/revenue/by-app', { params }),
  getRevenueComparison: (params?: { period?: string }) =>
    api.get('/insights/revenue/comparison', { params }),

  // Customers
  getCustomerAnalytics: () => api.get('/insights/customers'),
  getTopCustomers: (params?: { limit?: number }) =>
    api.get('/insights/customers/top', { params }),
  getCustomerChurn: () => api.get('/insights/customers/churn'),
  getCustomerLtv: () => api.get('/insights/customers/ltv'),

  // Products
  getProductPerformance: () => api.get('/insights/products'),
  getBestSellers: (params?: { limit?: number }) =>
    api.get('/insights/products/best-sellers', { params }),
  getSlowMovers: (params?: { limit?: number }) =>
    api.get('/insights/products/slow-movers', { params }),
  getMarginAnalysis: () => api.get('/insights/products/margins'),
  getCategoryBreakdown: () => api.get('/insights/products/categories'),
  getErpProducts: () => api.get('/erp/products'),

  // Operations
  getOperationsSummary: () => api.get('/insights/operations'),
  getStaffUtilization: () => api.get('/insights/operations/utilization'),
  getShiftCoverage: () => api.get('/insights/operations/shifts'),
  getTimeOffTrends: () => api.get('/insights/operations/time-off'),
  getDeptHeadcount: () => api.get('/insights/operations/headcount'),
  getHrDashboard: () => api.get('/hr/dashboard'),

  // Financial
  getProfitLoss: (params?: { start?: string; end?: string }) =>
    api.get('/finance/pl', { params }),
  getExpenseBreakdown: (params?: { start?: string; end?: string }) =>
    api.get('/insights/financial/expenses', { params }),
  getInvoiceAging: () => api.get('/insights/financial/invoice-aging'),
  getCashFlow: (params?: { period?: string }) =>
    api.get('/insights/financial/cash-flow', { params }),

  // Reports
  getReportTemplates: () => api.get('/insights/reports/templates'),
  generateReport: (templateId: string, params?: Record<string, string>) =>
    api.post(`/insights/reports/generate/${templateId}`, params),
  exportReport: (templateId: string, format: 'csv' | 'pdf', params?: Record<string, string>) =>
    api.post(`/insights/reports/export/${templateId}`, { format, ...params }, { responseType: 'blob' }),

  // Settings
  getInsightsSettings: () => api.get('/insights/settings'),
  updateInsightsSettings: (data: any) => api.put('/insights/settings', data),
  getScheduledReports: () => api.get('/insights/settings/scheduled-reports'),
  updateScheduledReport: (id: string, data: any) => api.put(`/insights/settings/scheduled-reports/${id}`, data),
  createScheduledReport: (data: any) => api.post('/insights/settings/scheduled-reports', data),
  deleteScheduledReport: (id: string) => api.delete(`/insights/settings/scheduled-reports/${id}`),
};
