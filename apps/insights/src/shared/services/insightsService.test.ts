import { describe, it, expect, vi, beforeEach } from 'vitest';
import { ResultAsync } from 'neverthrow';

vi.mock('../lib/api', () => ({
  safe: {
    get: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: [] }), () => ({}) as any)),
    post: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    put: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
    delete: vi.fn(() => ResultAsync.fromPromise(Promise.resolve({ data: {} }), () => ({}) as any)),
  },
}));

import { insightsService } from './insightsService';
import { safe } from '../lib/api';

describe('insightsService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Dashboard / KPIs
  it('getKpis calls GET /hub/kpis', () => {
    insightsService.getKpis();
    expect(safe.get).toHaveBeenCalledWith('/hub/kpis');
  });

  it('getDashboardSummary calls GET /insights/dashboard', () => {
    insightsService.getDashboardSummary();
    expect(safe.get).toHaveBeenCalledWith('/insights/dashboard');
  });

  // Revenue
  it('getRevenue calls GET /insights/revenue with params', () => {
    const params = { period: 'monthly' };
    insightsService.getRevenue(params);
    expect(safe.get).toHaveBeenCalledWith('/insights/revenue', { params });
  });

  it('getRevenueByApp calls GET /insights/revenue/by-app', () => {
    insightsService.getRevenueByApp();
    expect(safe.get).toHaveBeenCalledWith('/insights/revenue/by-app', { params: undefined });
  });

  it('getRevenueComparison calls GET /insights/revenue/comparison', () => {
    insightsService.getRevenueComparison({ period: 'weekly' });
    expect(safe.get).toHaveBeenCalledWith('/insights/revenue/comparison', {
      params: { period: 'weekly' },
    });
  });

  // Customers
  it('getCustomerAnalytics calls GET /insights/customers', () => {
    insightsService.getCustomerAnalytics();
    expect(safe.get).toHaveBeenCalledWith('/insights/customers');
  });

  it('getTopCustomers calls GET /insights/customers/top', () => {
    insightsService.getTopCustomers({ limit: 5 });
    expect(safe.get).toHaveBeenCalledWith('/insights/customers/top', { params: { limit: 5 } });
  });

  it('getCustomerChurn calls GET /insights/customers/churn', () => {
    insightsService.getCustomerChurn();
    expect(safe.get).toHaveBeenCalledWith('/insights/customers/churn');
  });

  // Products
  it('getProductPerformance calls GET /insights/products', () => {
    insightsService.getProductPerformance();
    expect(safe.get).toHaveBeenCalledWith('/insights/products');
  });

  it('getBestSellers calls GET /insights/products/best-sellers', () => {
    insightsService.getBestSellers({ limit: 10 });
    expect(safe.get).toHaveBeenCalledWith('/insights/products/best-sellers', {
      params: { limit: 10 },
    });
  });

  it('getErpProducts calls GET /erp/products', () => {
    insightsService.getErpProducts();
    expect(safe.get).toHaveBeenCalledWith('/erp/products');
  });

  // Operations
  it('getOperationsSummary calls GET /insights/operations', () => {
    insightsService.getOperationsSummary();
    expect(safe.get).toHaveBeenCalledWith('/insights/operations');
  });

  it('getStaffUtilization calls GET /insights/operations/utilization', () => {
    insightsService.getStaffUtilization();
    expect(safe.get).toHaveBeenCalledWith('/insights/operations/utilization');
  });

  it('getHrDashboard calls GET /hr/dashboard', () => {
    insightsService.getHrDashboard();
    expect(safe.get).toHaveBeenCalledWith('/hr/dashboard');
  });

  // Financial
  it('getProfitLoss calls GET /finance/pl with params', () => {
    const params = { start: '2026-01-01', end: '2026-03-31' };
    insightsService.getProfitLoss(params);
    expect(safe.get).toHaveBeenCalledWith('/finance/pl', { params });
  });

  it('getInvoiceAging calls GET /insights/financial/invoice-aging', () => {
    insightsService.getInvoiceAging();
    expect(safe.get).toHaveBeenCalledWith('/insights/financial/invoice-aging');
  });

  // Reports
  it('getReportTemplates calls GET /insights/reports/templates', () => {
    insightsService.getReportTemplates();
    expect(safe.get).toHaveBeenCalledWith('/insights/reports/templates');
  });

  it('generateReport calls POST /insights/reports/generate/:templateId', () => {
    const params = { start: '2026-01-01' };
    insightsService.generateReport('revenue', params);
    expect(safe.post).toHaveBeenCalledWith('/insights/reports/generate/revenue', params);
  });

  it('exportReport calls GET /insights/reports/export/revenue with format', () => {
    insightsService.exportReport('revenue', 'csv', { start: '2026-01-01' });
    expect(safe.get).toHaveBeenCalledWith('/insights/reports/export/revenue', {
      params: { format: 'csv', start: '2026-01-01' },
      responseType: 'blob',
    });
  });

  // Settings
  it('getInsightsSettings calls GET /insights/settings', () => {
    insightsService.getInsightsSettings();
    expect(safe.get).toHaveBeenCalledWith('/insights/settings');
  });

  it('updateInsightsSettings calls PUT /insights/settings', () => {
    const data = { defaultCurrency: 'USD' };
    insightsService.updateInsightsSettings(data);
    expect(safe.put).toHaveBeenCalledWith('/insights/settings', data);
  });

  it('createScheduledReport calls POST /insights/settings/scheduled-reports', () => {
    const data = { name: 'Weekly Revenue' };
    insightsService.createScheduledReport(data);
    expect(safe.post).toHaveBeenCalledWith('/insights/settings/scheduled-reports', data);
  });

  it('deleteScheduledReport calls DELETE /insights/settings/scheduled-reports/:id', () => {
    insightsService.deleteScheduledReport('abc');
    expect(safe.delete).toHaveBeenCalledWith('/insights/settings/scheduled-reports/abc');
  });
});
