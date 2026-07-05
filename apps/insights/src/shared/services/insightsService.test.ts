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
  it('getKpis calls GET /api/hub/dashboard/v1/kpis', () => {
    insightsService.getKpis();
    expect(safe.get).toHaveBeenCalledWith('/api/hub/dashboard/v1/kpis');
  });

  it('getDashboardSummary calls GET /api/insights/dashboard/v1/summary', () => {
    insightsService.getDashboardSummary();
    expect(safe.get).toHaveBeenCalledWith('/api/insights/dashboard/v1/summary');
  });

  // Revenue
  it('getRevenue calls GET /api/insights/revenue/v1/overview with params', () => {
    const params = { period: 'monthly' };
    insightsService.getRevenue(params);
    expect(safe.get).toHaveBeenCalledWith('/api/insights/revenue/v1/overview', { params });
  });

  it('getRevenueByApp calls GET /api/insights/revenue/v1/by-app', () => {
    insightsService.getRevenueByApp();
    expect(safe.get).toHaveBeenCalledWith('/api/insights/revenue/v1/by-app');
  });

  it('getRevenueComparison calls GET /api/insights/revenue/v1/comparison', () => {
    insightsService.getRevenueComparison({ period: 'weekly' });
    expect(safe.get).toHaveBeenCalledWith('/api/insights/revenue/v1/comparison', {
      params: { compare_to: 'previous_period', period: 'weekly' },
    });
  });

  // Customers
  it('getCustomerAnalytics calls GET /api/insights/customer/v1/analytics', () => {
    insightsService.getCustomerAnalytics();
    expect(safe.get).toHaveBeenCalledWith('/api/insights/customer/v1/analytics');
  });

  it('getTopCustomers calls GET /api/insights/customer/v1/top', () => {
    insightsService.getTopCustomers({ limit: 5 });
    expect(safe.get).toHaveBeenCalledWith('/api/insights/customer/v1/top', { params: { limit: 5 } });
  });

  it('getCustomerChurn calls GET /api/insights/customer/v1/churn', () => {
    insightsService.getCustomerChurn();
    expect(safe.get).toHaveBeenCalledWith('/api/insights/customer/v1/churn');
  });

  // Products
  it('getProductPerformance calls GET /api/insights/product/v1/performance', () => {
    insightsService.getProductPerformance();
    expect(safe.get).toHaveBeenCalledWith('/api/insights/product/v1/performance');
  });

  it('getBestSellers calls GET /api/insights/product/v1/best-sellers', () => {
    insightsService.getBestSellers({ limit: 10 });
    expect(safe.get).toHaveBeenCalledWith('/api/insights/product/v1/best-sellers', {
      params: { limit: 10 },
    });
  });

  it('getErpProducts calls GET /api/erp/product/v1/products', () => {
    insightsService.getErpProducts();
    expect(safe.get).toHaveBeenCalledWith('/api/erp/product/v1/products');
  });

  // Operations
  it('getOperationsSummary calls GET /api/insights/operation/v1/summary', () => {
    insightsService.getOperationsSummary();
    expect(safe.get).toHaveBeenCalledWith('/api/insights/operation/v1/summary');
  });

  it('getStaffUtilization calls GET /api/insights/operation/v1/utilization', () => {
    insightsService.getStaffUtilization();
    expect(safe.get).toHaveBeenCalledWith('/api/insights/operation/v1/utilization');
  });

  it('getHrDashboard calls GET /api/hr/dashboard/v1/dashboard', () => {
    insightsService.getHrDashboard();
    expect(safe.get).toHaveBeenCalledWith('/api/hr/dashboard/v1/dashboard');
  });

  // Financial
  it('getProfitLoss calls GET /api/finance/report/v1/pl with params', () => {
    const params = { start: '2026-01-01', end: '2026-03-31' };
    insightsService.getProfitLoss(params);
    expect(safe.get).toHaveBeenCalledWith('/api/finance/report/v1/pl', {
      params: { period: '2026-01-01,2026-03-31', date: '2026-03-31' },
    });
  });

  it('getInvoiceAging calls GET /api/insights/financial/v1/invoice-aging', () => {
    insightsService.getInvoiceAging();
    expect(safe.get).toHaveBeenCalledWith('/api/insights/financial/v1/invoice-aging');
  });

  // Reports
  it('getReportTemplates calls GET /api/insights/report/v1/templates', () => {
    insightsService.getReportTemplates();
    expect(safe.get).toHaveBeenCalledWith('/api/insights/report/v1/templates');
  });

  it('generateReport calls POST /api/insights/report/v1/reports/revenue/generate', () => {
    const params = { start: '2026-01-01' };
    insightsService.generateReport('revenue', params);
    expect(safe.post).toHaveBeenCalledWith('/api/insights/report/v1/reports/revenue/generate', params);
  });

  it('exportReport calls GET /api/insights/report/v1/reports/revenue/export', () => {
    insightsService.exportReport('revenue', 'csv', { start: '2026-01-01' });
    expect(safe.get).toHaveBeenCalledWith('/api/insights/report/v1/reports/revenue/export', {
      params: { start: '2026-01-01' },
    });
  });

  // Settings
  it('getInsightsSettings calls GET /api/insights/setting/v1/settings', () => {
    insightsService.getInsightsSettings();
    expect(safe.get).toHaveBeenCalledWith('/api/insights/setting/v1/settings');
  });

  it('updateInsightsSettings calls PUT /api/insights/setting/v1/settings', () => {
    const data = { defaultCurrency: 'USD' };
    insightsService.updateInsightsSettings(data);
    expect(safe.put).toHaveBeenCalledWith('/api/insights/setting/v1/settings', {
      settings: {
        default_period: 'this_month',
        currency: 'USD',
        compare_previous: false,
        dashboard_widgets: [],
        favorite_reports: [],
      },
    });
  });

  it('createScheduledReport calls POST /api/insights/report/v1/scheduled', () => {
    const data = { templateId: 'revenue', frequency: 'weekly', channel: 'email', recipient: 'a@b.com' };
    insightsService.createScheduledReport(data);
    expect(safe.post).toHaveBeenCalledWith('/api/insights/report/v1/scheduled', {
      template_slug: 'revenue',
      frequency: 'weekly',
      channel: 'email',
      recipient_email: 'a@b.com',
      is_active: true,
    });
  });

  it('deleteScheduledReport calls DELETE /api/insights/report/v1/scheduled/:id', () => {
    insightsService.deleteScheduledReport('abc');
    expect(safe.delete).toHaveBeenCalledWith('/api/insights/report/v1/scheduled/abc');
  });
});
