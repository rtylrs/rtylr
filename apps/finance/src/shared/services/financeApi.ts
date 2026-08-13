import { safe } from '../lib/api';
import type { InvoiceItem as InvoiceLine } from '@voxire/voxie/finance/invoices/v1/types';
import type { CreateInvoiceRequest } from '@voxire/voxie/finance/invoices/v1/types';
import type { CreateRecurringInvoiceRequest } from '@voxire/voxie/finance/recurring_invoice/v1/types';
import type { CreateBillRequest } from '@voxire/voxie/finance/bill/v1/types';
import type { CreateBankAccountRequest } from '@voxire/voxie/finance/bank_account/v1/types';
import type { CreateAccountRequest } from '@voxire/voxie/finance/account/v1/types';
import type { CreateJournalEntryRequest } from '@voxire/voxie/finance/journal/v1/types';

export const plReportQueryParams = (params?: {
  period?: string;
  from?: string;
  to?: string;
  start?: string;
  end?: string;
  date?: string;
}) => {
  const from = params?.from ?? params?.start;
  const to = params?.to ?? params?.end;
  const date = params?.date ?? to ?? from ?? new Date().toISOString().slice(0, 10);
  const period = params?.period ?? (from && to ? `${from},${to}` : 'month');
  return { period, date };
};

// ── Types ────────────────────────────────────────────────────────

export interface DashboardStats {
  // Returned by the deployed backend DashboardResponse.
  totalReceivable: number;
  overdueAmount: number;
  totalRevenueMtd: number;
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

export interface QuotePayload {
  customerName: string;
  customerEmail?: string;
  status?: string;
  items?: string;
  subtotal?: number;
  tax?: number;
  discount?: number;
  total?: number;
  validUntil?: string;
  notes?: string;
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
  id?: string;
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
  period: string;
  taxCollected: number;
  taxOwed: number;
}

type ExpenseMutationPayload = Pick<
  Expense,
  'categoryId' | 'amount' | 'currency' | 'description' | 'vendor' | 'receiptUrl' | 'expenseDate'
>;

const toExpenseMutationPayload = (data: Partial<Expense>): Partial<ExpenseMutationPayload> => ({
  categoryId: data.categoryId,
  amount: data.amount,
  currency: data.currency ?? undefined,
  description: data.description,
  vendor: data.vendor ?? undefined,
  receiptUrl: data.receiptUrl ?? undefined,
  expenseDate: data.expenseDate,
});

const toExpenseCategoryMutationPayload = (data: Partial<ExpenseCategory>) => ({
  name: data.name,
  color: data.color,
  icon: 'icon' in data ? (data as Partial<ExpenseCategory> & { icon?: string }).icon : undefined,
});

const toTaxConfigPayload = (data: TaxConfig) => ({
  name: data.name,
  rate: data.rate,
  type: 'type' in data ? (data as TaxConfig & { type?: string }).type : undefined,
  isDefault:
    'isDefault' in data ? (data as TaxConfig & { isDefault?: boolean }).isDefault : undefined,
});

const toRecurringInvoicePayload = (data: CreateRecurringInvoiceRequest) => {
  const source = data as CreateRecurringInvoiceRequest & {
    templateInvoiceId?: number;
    customerId?: number;
    startDate?: string;
    nextInvoiceDate?: string;
    endDate?: string;
    isActive?: boolean;
  };
  return {
    templateInvoiceId: source.templateInvoiceId ?? source.customerId,
    frequency: source.frequency,
    nextInvoiceDate: source.nextInvoiceDate ?? source.startDate,
    endDate: source.endDate,
    isActive: source.isActive ?? true,
  };
};

const toBankAccountPayload = (data: CreateBankAccountRequest) => {
  const source = data as CreateBankAccountRequest & {
    institution?: string;
    bankName?: string;
    accountType?: string;
    balance?: number;
    currency?: string;
    isConnected?: boolean;
  };
  return {
    name: source.name,
    institution: source.institution ?? source.bankName,
    accountType: source.accountType,
    balance: source.balance ?? 0,
    currency: source.currency ?? 'USD',
    isConnected: source.isConnected ?? false,
  };
};

const toNumericId = (id: string): number => {
  const direct = Number(id);
  if (Number.isFinite(direct)) return direct;
  const digits = id.match(/\d+/g)?.join('');
  return digits ? Number(digits) : 0;
};

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
    safe.post('/finance/invoices/payment', { invoiceId: toNumericId(id), amount }),

  // Expenses
  getExpenses: () => safe.get<Expense[]>('/finance/expenses'),

  createExpense: (data: Partial<Expense>) =>
    safe.post<Expense>('/finance/expenses', toExpenseMutationPayload(data)),

  updateExpense: (id: string, data: Partial<Expense>) =>
    safe.put<Expense>('/finance/expenses', { id, ...data }),

  deleteExpense: (id: string) => safe.delete('/finance/expenses', { data: { id } }),

  // Expense Categories
  getExpenseCategories: () => safe.get<ExpenseCategory[]>('/finance/expense-categories'),

  createExpenseCategory: (data: Partial<ExpenseCategory>) =>
    safe.post<ExpenseCategory>(
      '/finance/expense-categories',
      toExpenseCategoryMutationPayload(data)
    ),

  // P&L
  getPL: (period: string) =>
    safe.get<PLReport>('/finance/report/v1/pl', {
      params: plReportQueryParams({ period, date: new Date().toISOString().slice(0, 10) }),
    }),

  // Tax
  getTaxConfig: () => safe.get<TaxConfig[]>('/finance/tax/config'),

  createTaxConfig: (data: TaxConfig) =>
    safe.post<TaxConfig>('/finance/tax/config', toTaxConfigPayload(data)),

  getTaxReport: () => safe.get<TaxReport>('/finance/tax/report'),

  // Recurring Invoices
  getRecurringInvoices: () => safe.get('/finance/recurring-invoices'),

  createRecurringInvoice: (data: CreateRecurringInvoiceRequest) =>
    safe.post('/finance/recurring-invoices', toRecurringInvoicePayload(data)),

  updateRecurringInvoice: (id: string, data: Partial<CreateRecurringInvoiceRequest>) =>
    safe.put('/finance/recurring-invoices', { id, ...data }),

  // Quotes
  getQuotes: () => safe.get('/finance/quotes'),

  createQuote: (data: QuotePayload) => safe.post('/finance/quotes', data),

  updateQuote: (id: string, data: Partial<QuotePayload>) =>
    safe.put('/finance/quotes', { id, ...data }),

  deleteQuote: (id: string) => safe.delete('/finance/quotes', { data: { id } }),

  // Bills
  getBills: () => safe.get('/finance/bills'),

  createBill: (data: CreateBillRequest) => safe.post('/finance/bills', data),

  updateBill: (id: string, data: Partial<CreateBillRequest>) =>
    safe.put('/finance/bills', { id, ...data }),

  // Bank Accounts
  getBankAccounts: () => safe.get('/finance/bank-accounts'),

  createBankAccount: (data: CreateBankAccountRequest) =>
    safe.post('/finance/bank-accounts', toBankAccountPayload(data)),

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
