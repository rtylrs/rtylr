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

import { financeApi } from './financeApi';
import { safe } from '../lib/api';

describe('financeApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  // Dashboard
  it('getDashboard calls GET /finance/dashboard', () => {
    financeApi.getDashboard();
    expect(safe.get).toHaveBeenCalledWith('/finance/dashboard');
  });

  // Invoices
  it('getInvoices calls GET /finance/invoices', () => {
    financeApi.getInvoices();
    expect(safe.get).toHaveBeenCalledWith('/finance/invoices');
  });

  it('getInvoiceDetail calls GET /finance/invoices/detail with id', () => {
    financeApi.getInvoiceDetail('INV-001');
    expect(safe.get).toHaveBeenCalledWith('/finance/invoices/detail', {
      params: { id: 'INV-001' },
    });
  });

  it('createInvoice calls POST /finance/invoices', () => {
    const data = { client: 'Acme', amount: 1000 };
    financeApi.createInvoice(data);
    expect(safe.post).toHaveBeenCalledWith('/finance/invoices', data);
  });

  it('updateInvoice calls POST /finance/invoices with id', () => {
    financeApi.updateInvoice('INV-001', { amount: 1500 });
    expect(safe.post).toHaveBeenCalledWith('/finance/invoices', { id: 'INV-001', amount: 1500 });
  });

  it('deleteInvoice calls POST /finance/invoices with delete action', () => {
    financeApi.deleteInvoice('INV-001');
    expect(safe.post).toHaveBeenCalledWith('/finance/invoices', {
      id: 'INV-001',
      _action: 'delete',
    });
  });

  it('updateInvoiceStatus calls POST /finance/invoices/status', () => {
    financeApi.updateInvoiceStatus('INV-001', 'Paid');
    expect(safe.post).toHaveBeenCalledWith('/finance/invoices/status', {
      id: 'INV-001',
      status: 'Paid',
    });
  });

  it('recordPayment calls POST /finance/invoices/payment', () => {
    financeApi.recordPayment('INV-001', 500);
    expect(safe.post).toHaveBeenCalledWith('/finance/invoices/payment', {
      id: 'INV-001',
      amount: 500,
    });
  });

  // Expenses
  it('getExpenses calls GET /finance/expenses', () => {
    financeApi.getExpenses();
    expect(safe.get).toHaveBeenCalledWith('/finance/expenses');
  });

  it('createExpense calls POST /finance/expenses', () => {
    const data = { description: 'Office', amount: 50 };
    financeApi.createExpense(data);
    expect(safe.post).toHaveBeenCalledWith('/finance/expenses', data);
  });

  it('deleteExpense calls POST /finance/expenses with delete action', () => {
    financeApi.deleteExpense('EXP-1');
    expect(safe.post).toHaveBeenCalledWith('/finance/expenses', { id: 'EXP-1', _action: 'delete' });
  });

  it('getExpenseCategories calls GET /finance/expense-categories', () => {
    financeApi.getExpenseCategories();
    expect(safe.get).toHaveBeenCalledWith('/finance/expense-categories');
  });

  // P&L
  it('getPL calls GET /finance/pl with period param', () => {
    financeApi.getPL('2026-Q1');
    expect(safe.get).toHaveBeenCalledWith('/finance/pl', { params: { period: '2026-Q1' } });
  });

  // Tax
  it('getTaxConfig calls GET /finance/tax/config', () => {
    financeApi.getTaxConfig();
    expect(safe.get).toHaveBeenCalledWith('/finance/tax/config');
  });

  it('createTaxConfig calls POST /finance/tax/config', () => {
    const data = { name: 'VAT', rate: 20 };
    financeApi.createTaxConfig(data);
    expect(safe.post).toHaveBeenCalledWith('/finance/tax/config', data);
  });

  it('getTaxReport calls GET /finance/tax/report', () => {
    financeApi.getTaxReport();
    expect(safe.get).toHaveBeenCalledWith('/finance/tax/report');
  });

  // Recurring Invoices
  it('getRecurringInvoices calls GET /finance/recurring-invoices', () => {
    financeApi.getRecurringInvoices();
    expect(safe.get).toHaveBeenCalledWith('/finance/recurring-invoices');
  });

  it('createRecurringInvoice calls POST /finance/recurring-invoices', () => {
    const data = { client: 'Acme' };
    financeApi.createRecurringInvoice(data);
    expect(safe.post).toHaveBeenCalledWith('/finance/recurring-invoices', data);
  });

  // Quotes
  it('getQuotes calls GET /finance/quotes', () => {
    financeApi.getQuotes();
    expect(safe.get).toHaveBeenCalledWith('/finance/quotes');
  });

  it('deleteQuote calls POST /finance/quotes with delete action', () => {
    financeApi.deleteQuote('Q-1');
    expect(safe.post).toHaveBeenCalledWith('/finance/quotes', { id: 'Q-1', _action: 'delete' });
  });

  // Bills
  it('getBills calls GET /finance/bills', () => {
    financeApi.getBills();
    expect(safe.get).toHaveBeenCalledWith('/finance/bills');
  });

  // Bank Accounts
  it('getBankAccounts calls GET /finance/bank-accounts', () => {
    financeApi.getBankAccounts();
    expect(safe.get).toHaveBeenCalledWith('/finance/bank-accounts');
  });

  // Balance Sheet & Cash Flow
  it('getBalanceSheet calls GET /finance/balance-sheet', () => {
    financeApi.getBalanceSheet();
    expect(safe.get).toHaveBeenCalledWith('/finance/balance-sheet');
  });

  it('getCashFlow calls GET /finance/cash-flow', () => {
    financeApi.getCashFlow();
    expect(safe.get).toHaveBeenCalledWith('/finance/cash-flow');
  });

  // Accounts & Journal
  it('getAccounts calls GET /finance/accounts', () => {
    financeApi.getAccounts();
    expect(safe.get).toHaveBeenCalledWith('/finance/accounts');
  });

  it('createJournalEntry calls POST /finance/journal', () => {
    const data = { entries: [] };
    financeApi.createJournalEntry(data);
    expect(safe.post).toHaveBeenCalledWith('/finance/journal', data);
  });

  // Expense detail
  it('getExpenseDetail calls GET /finance/expenses with id param', () => {
    financeApi.getExpenseDetail('EXP-1');
    expect(safe.get).toHaveBeenCalledWith('/finance/expenses', { params: { id: 'EXP-1' } });
  });
});
