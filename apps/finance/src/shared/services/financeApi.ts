import { safe } from '../lib/api';
import type { InvoiceItem as InvoiceLine } from '@voxire/contracts/finance/v1/invoices';
import type { CreateRecurringInvoiceRequest } from '@voxire/contracts/finance/v1/recurring_invoices';
import type { CreateQuoteRequest } from '@voxire/contracts/finance/v1/quotes';
import type { CreateBillRequest } from '@voxire/contracts/finance/v1/bills';
import type { CreateBankAccountRequest } from '@voxire/contracts/finance/v1/bank_accounts';
import type { CreateAccountRequest } from '@voxire/contracts/finance/v1/accounts';
import type { CreateJournalEntryRequest } from '@voxire/contracts/finance/v1/journal';

// ── Types ────────────────────────────────────────────────────────

export interface DashboardStats {
  totalRevenueMTD: number;
  totalExpensesMTD: number;
  netProfit: number;
  outstandingInvoicesTotal: number;
  outstandingInvoicesCount: number;
  totalRevenueMTDChange: string;
  totalExpensesMTDChange: string;
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
  date: string;
  description: string;
  category: string;
  amount: number;
  receipt: boolean;
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

  createInvoice: (data: Partial<Invoice>) => safe.post<Invoice>('/finance/invoices', data),

  updateInvoice: (id: string, data: Partial<Invoice>) =>
    safe.post<Invoice>('/finance/invoices', { id, ...data }),

  deleteInvoice: (id: string) => safe.post('/finance/invoices', { id, _action: 'delete' }),

  updateInvoiceStatus: (id: string, status: Invoice['status']) =>
    safe.post('/finance/invoices/status', { id, status }),

  recordPayment: (id: string, amount: number) =>
    safe.post('/finance/invoices/payment', { id, amount }),

  // Expenses
  getExpenses: () => safe.get<Expense[]>('/finance/expenses'),

  createExpense: (data: Partial<Expense>) => safe.post<Expense>('/finance/expenses', data),

  updateExpense: (id: string, data: Partial<Expense>) =>
    safe.post<Expense>('/finance/expenses', { id, ...data }),

  deleteExpense: (id: string) => safe.post('/finance/expenses', { id, _action: 'delete' }),

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
    safe.post('/finance/recurring-invoices', { id, ...data }),

  // Quotes
  getQuotes: () => safe.get('/finance/quotes'),

  createQuote: (data: CreateQuoteRequest) => safe.post('/finance/quotes', data),

  updateQuote: (id: string, data: Partial<CreateQuoteRequest>) =>
    safe.post('/finance/quotes', { id, ...data }),

  deleteQuote: (id: string) => safe.post('/finance/quotes', { id, _action: 'delete' }),

  // Bills
  getBills: () => safe.get('/finance/bills'),

  createBill: (data: CreateBillRequest) => safe.post('/finance/bills', data),

  updateBill: (id: string, data: Partial<CreateBillRequest>) =>
    safe.post('/finance/bills', { id, ...data }),

  // Bank Accounts
  getBankAccounts: () => safe.get('/finance/bank-accounts'),

  createBankAccount: (data: CreateBankAccountRequest) => safe.post('/finance/bank-accounts', data),

  // Reconciliation
  getReconciliation: () => safe.get('/finance/reconciliation'),
  matchTransaction: (transactionId: string, matchId: string) =>
    safe.post('/finance/reconciliation/match', {
      transaction_id: transactionId,
      match_id: matchId,
    }),
  autoMatchTransactions: () => safe.post('/finance/reconciliation/match', { auto: true }),

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
