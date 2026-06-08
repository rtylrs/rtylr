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

import { financeService } from './financeService';
import { safe } from '../lib/api';

describe('financeService', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('getDashboard calls GET /finance/dashboard', () => {
    financeService.getDashboard();
    expect(safe.get).toHaveBeenCalledWith('/finance/dashboard');
  });

  // Invoices
  it('getInvoices calls GET /finance/invoices with params', () => {
    const params = { status: 'paid' };
    financeService.getInvoices(params);
    expect(safe.get).toHaveBeenCalledWith('/finance/invoices', { params });
  });

  it('createInvoice calls POST /finance/invoices', () => {
    const data = { client: 'Acme' };
    financeService.createInvoice(data);
    expect(safe.post).toHaveBeenCalledWith('/finance/invoices', data);
  });

  it('getInvoice calls GET /finance/invoices/detail with id param', () => {
    financeService.getInvoice(3);
    expect(safe.get).toHaveBeenCalledWith('/finance/invoices/detail', { params: { id: 3 } });
  });

  it('updateInvoice calls PUT /finance/invoices with id', () => {
    financeService.updateInvoice(3, { amount: 500 });
    expect(safe.put).toHaveBeenCalledWith('/finance/invoices', { id: 3, amount: 500 });
  });

  it('deleteInvoice calls DELETE /finance/invoices with id', () => {
    financeService.deleteInvoice(3);
    expect(safe.delete).toHaveBeenCalledWith('/finance/invoices', { data: { id: 3 } });
  });

  it('updateInvoiceStatus calls PUT /finance/invoices/status with id', () => {
    financeService.updateInvoiceStatus(3, 'paid');
    expect(safe.put).toHaveBeenCalledWith('/finance/invoices/status', { id: 3, status: 'paid' });
  });

  it('recordPayment calls POST /finance/invoices/payment with invoiceId', () => {
    const data = { amount: 100 };
    financeService.recordPayment(3, data);
    expect(safe.post).toHaveBeenCalledWith('/finance/invoices/payment', {
      invoiceId: 3,
      amount: 100,
    });
  });

  // Expenses
  it('getExpenses calls GET /finance/expenses with params', () => {
    financeService.getExpenses({ category: 'office' });
    expect(safe.get).toHaveBeenCalledWith('/finance/expenses', { params: { category: 'office' } });
  });

  it('createExpense calls POST /finance/expenses', () => {
    const data = { description: 'Lunch' };
    financeService.createExpense(data);
    expect(safe.post).toHaveBeenCalledWith('/finance/expenses', data);
  });

  it('updateExpense calls PUT /finance/expenses/:id', () => {
    financeService.updateExpense(2, { amount: 25 });
    expect(safe.put).toHaveBeenCalledWith('/finance/expenses/2', { amount: 25 });
  });

  it('deleteExpense calls DELETE /finance/expenses/:id', () => {
    financeService.deleteExpense(2);
    expect(safe.delete).toHaveBeenCalledWith('/finance/expenses/2');
  });

  it('getExpenseCategories calls GET /finance/expenses/categories', () => {
    financeService.getExpenseCategories();
    expect(safe.get).toHaveBeenCalledWith('/finance/expenses/categories');
  });

  // Tax & P&L
  it('getTaxConfig calls GET /finance/tax', () => {
    financeService.getTaxConfig();
    expect(safe.get).toHaveBeenCalledWith('/finance/tax');
  });

  it('updateTaxConfig calls PUT /finance/tax', () => {
    const data = { rate: 15 };
    financeService.updateTaxConfig(data);
    expect(safe.put).toHaveBeenCalledWith('/finance/tax', data);
  });

  it('getPLReport calls GET /finance/pl with params', () => {
    const params = { period: '2026-Q1' };
    financeService.getPLReport(params);
    expect(safe.get).toHaveBeenCalledWith('/finance/pl', { params });
  });
});
