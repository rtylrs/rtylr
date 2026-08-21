import { describe, expect, it, vi } from 'vitest';

const mockApi = vi.hoisted(() => {
  const methods: Record<string, ReturnType<typeof vi.fn>> = {};
  return new Proxy(methods, {
    get(target, property: string) {
      target[property] ??= vi.fn();
      return target[property];
    },
  });
});

vi.mock('./financeApi', () => ({
  financeApi: mockApi,
  plReportQueryParams: vi.fn(() => ({ period: 'month', date: '2026-01-01' })),
}));

import { financeService } from './financeService';

describe('financeService', () => {
  it('delegates invoice, expense, tax, and report operations to financeApi', () => {
    const invoiceParams = { status: 'paid' };
    const expenseParams = { category: 'office' };
    const tax = { rate: 15 };

    financeService.getDashboard();
    financeService.getInvoices(invoiceParams);
    financeService.getInvoice(3);
    financeService.updateInvoice(3, { amount: 500 });
    financeService.deleteInvoice(3);
    financeService.getExpenses(expenseParams);
    financeService.updateTaxConfig(tax);
    financeService.getPLReport({ start: '2026-01-01', end: '2026-03-31' });

    expect(mockApi.getDashboard).toHaveBeenCalledWith();
    expect(mockApi.getInvoices).toHaveBeenCalledWith(invoiceParams);
    expect(mockApi.getInvoiceDetail).toHaveBeenCalledWith('3');
    expect(mockApi.updateInvoice).toHaveBeenCalledWith('3', { amount: 500 });
    expect(mockApi.deleteInvoice).toHaveBeenCalledWith('3');
    expect(mockApi.getExpenses).toHaveBeenCalledWith(expenseParams);
    expect(mockApi.updateTaxConfig).toHaveBeenCalledWith(tax);
    expect(mockApi.getPL).toHaveBeenCalledWith('month');
  });
});
