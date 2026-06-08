import { safe } from '../lib/api';
import type { InvoiceItem as InvoiceLine } from '@voxire/contracts/finance/invoices/v1/types';
import type { CreateInvoiceRequest } from '@voxire/contracts/finance/invoices/v1/create_invoice';
import type { CreateRecurringInvoiceRequest } from '@voxire/contracts/finance/recurring_invoice/v1/create_recurring_invoice';
import type { CreateQuoteRequest } from '@voxire/contracts/finance/quote/v1/create_quote';
import type { CreateBillRequest } from '@voxire/contracts/finance/bill/v1/create_bill';
import type { CreateBankAccountRequest } from '@voxire/contracts/finance/bank_account/v1/create_bank_account';
import type { CreateAccountRequest } from '@voxire/contracts/finance/account/v1/create_account';
import type { CreateJournalEntryRequest } from '@voxire/contracts/finance/journal/v1/create_journal_entry';

// ── Types ────────────────────────────────────────────────────────

export interface DashboardStats {
  // Returned by the deployed backend DashboardResponse.
  totalReceivable: number;
  overdueAmount: number;
  totalExpensesMtd: number;
  netProfitMtd: number;
  // Not yet returned by the backend; safe defaults applied by the schema.
  outstandingInvoicesCount: number;
  totalRevenueMtdChange: string;
  totalExpensesMtdChange: string;
  netProfitChange: string;
  recentTransactions: Transaction[];
  overdueInvoices: OverdueInvoice[];
  barData: BarDataPoint[];
}

export interface Transaction {
  id: string;
  description: string;
  amount: string;
  type: 'credit' | 'debit';
  date: string;
}

export interface OverdueInvoice {
  id: string;
  client: string;
  amount: string;
  daysOverdue: number;
}

export interface BarDataPoint {
  month: string;
  revenue: number;
  expenses: number;
}

export interface Invoice {
  id: string;
  client: string;
  clientEmail?: string;
  amount: number;
  currency?: string | null;
  paid?: number;
  status: 'Draft' | 'Sent' | 'Paid' | 'Overdue';
  dueDate: string;
  issuedDate: string;
  items?: InvoiceLine[];
  notes?: string;
  discount?: number;
  tax?: number;
  subtotal?: number;
  number?: string;
  history?: { label: string; at: string }[];
}

export interface Expense {
  id: string;
  ownerId?: number | null;
  categoryId?: number | null;
  description: string;
  amount: number;
  currency?: string | null;
  vendor?: string | null;
  receiptUrl?: string | null;
  expenseDate: string;
  status: string;
  notes?: string | null;
  createdAt?: string | null;
  updatedAt?: string | null;
}

export interface ExpenseCategory {
  id: number;
  name: string;
  color: string;
  spent?: number;
  budget?: number;
  count?: number;
}

export interface PLReport {
  period: string;
  revenue: number;
  cogs: number;
  operatingExpenses: number;
  grossProfit: number;
  netProfit: number;
}

export interface TaxConfig {
  id?: number;
  name: string;
  rate: number;
}

export interface TaxPeriod {
  id: string;
  period: string;
  revenue: number;
  taxableAmount: number;
  taxRate: number;
  collected: number;
  status: 'Filed' | 'Pending' | 'Due';
  dueDate: string;
}

export interface TaxReport {
  vatSummary: { collected: number; owed: number; paid: number };
  periods: TaxPeriod[];
  configs: TaxConfig[];
}

// ── Service ──────────────────────────────────────────────────────

export const financeApi = {
  // Dashboard
  getDashboard: () => safe.get<DashboardStats>('/finance/dashboard'),

  // Invoices
  getInvoices: () => safe.get<Invoice[]>('/finance/invoices'),

  getInvoiceDetail: (id: string) =>
    safe.get<Invoice>('/finance/invoices/detail', { params: { id } }),

  createInvoice: (data: CreateInvoiceRequest) => safe.post<Invoice>('/finance/invoices', data),

  updateInvoice: (id: string, data: Partial<Invoice>) =>
    safe.put<Invoice>('/finance/invoices', { id, ...data }),

  deleteInvoice: (id: string) => safe.delete('/finance/invoices', { data: { id } }),

  updateInvoiceStatus: (id: string, status: Invoice['status']) =>
    safe.put('/finance/invoices/status', { id, status }),

  recordPayment: (id: string, amount: number) =>
    safe.post('/finance/invoices/payment', { id, amount }),

  // Expenses
  getExpenses: () => safe.get<Expense[]>('/finance/expenses'),

  createExpense: (data: Partial<Expense>) => safe.post<Expense>('/finance/expenses', data),

  updateExpense: (id: string, data: Partial<Expense>) =>
    safe.put<Expense>('/finance/expenses', { id, ...data }),

  deleteExpense: (id: string) => safe.delete('/finance/expenses', { data: { id } }),

  // Expense Categories
  getExpenseCategories: () => safe.get<ExpenseCategory[]>('/finance/expense-categories'),

  createExpenseCategory: (data: Partial<ExpenseCategory>) =>
    safe.post<ExpenseCategory>('/finance/expense-categories', data),

  // P&L
  getPL: (period: string) => safe.get<PLReport>('/finance/pl', { params: { period } }),

  // Tax
  getTaxConfig: () => safe.get<TaxConfig[]>('/finance/tax/config'),

  createTaxConfig: (data: TaxConfig) => safe.post<TaxConfig>('/finance/tax/config', data),

  getTaxReport: () => safe.get<TaxReport>('/finance/tax/report'),

  // Recurring Invoices
  getRecurringInvoices: () => safe.get('/finance/recurring-invoices'),

  createRecurringInvoice: (data: CreateRecurringInvoiceRequest) =>
    safe.post('/finance/recurring-invoices', data),

  updateRecurringInvoice: (id: string, data: Partial<CreateRecurringInvoiceRequest>) =>
    safe.put('/finance/recurring-invoices', { id, ...data }),

  // Quotes
  getQuotes: () => safe.get('/finance/quotes'),

  createQuote: (data: CreateQuoteRequest) => safe.post('/finance/quotes', data),

  updateQuote: (id: string, data: Partial<CreateQuoteRequest>) =>
    safe.put('/finance/quotes', { id, ...data }),

  deleteQuote: (id: string) => safe.delete('/finance/quotes', { data: { id } }),

  // Bills
  getBills: () => safe.get('/finance/bills'),

  createBill: (data: CreateBillRequest) => safe.post('/finance/bills', data),

  updateBill: (id: string, data: Partial<CreateBillRequest>) =>
    safe.put('/finance/bills', { id, ...data }),

  // Bank Accounts
  getBankAccounts: () => safe.get('/finance/bank-accounts'),

  createBankAccount: (data: CreateBankAccountRequest) => safe.post('/finance/bank-accounts', data),

  // Reconciliation
  getReconciliation: () => safe.get('/finance/reconciliation'),
  matchTransaction: (
    bankFeedId: string | number,
    targetType: 'invoice' | 'expense',
    targetId: string | number
  ) =>
    safe.post('/finance/reconciliation/match', {
      bankFeedId: Number(bankFeedId),
      targetType,
      targetId: Number(targetId),
    }),

  // Balance Sheet
  getBalanceSheet: () => safe.get('/finance/balance-sheet'),

  // Cash Flow
  getCashFlow: () => safe.get('/finance/cash-flow'),

  // Chart of Accounts
  getAccounts: () => safe.get('/finance/accounts'),

  createAccount: (data: CreateAccountRequest) => safe.post('/finance/accounts', data),

  // Journal
  getJournal: () => safe.get('/finance/journal'),

  createJournalEntry: (data: CreateJournalEntryRequest) => safe.post('/finance/journal', data),

  // Expense detail
  getExpenseDetail: (id: string) => safe.get('/finance/expenses', { params: { id } }),
};
