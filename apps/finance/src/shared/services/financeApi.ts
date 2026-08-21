import { fromOneKitJson, oneKitResult, toOneKitJson } from '@voxire/api';
import type { InvoiceItem as InvoiceLine } from '@voxire/voxie/finance/invoices/v1/types';
import type { CreateInvoiceRequest } from '@voxire/voxie/finance/invoices/v1/types';
import type { CreateRecurringInvoiceRequest } from '@voxire/voxie/finance/recurring_invoice/v1/types';
import type { CreateBillRequest } from '@voxire/voxie/finance/bill/v1/types';
import type { CreateBankAccountRequest } from '@voxire/voxie/finance/bank_account/v1/types';
import type { CreateAccountRequest } from '@voxire/voxie/finance/account/v1/types';
import type { CreateJournalEntryRequest } from '@voxire/voxie/finance/journal/v1/types';
import { expensesClient, financeCompatibilityClient, invoicesClient } from '../lib/clients';

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
  status: 'Draft' | 'Sent' | 'Paid' | 'Overdue' | 'Cancelled';
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
  icon?: string;
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

function compatResult<T>(operation: () => Promise<unknown>) {
  return oneKitResult(operation).map((value) => fromOneKitJson<T>(value));
}

function generatedResult<T>(operation: () => Promise<unknown>) {
  return oneKitResult(operation).map((value) => value as T);
}

function body(value: unknown) {
  return { body: toOneKitJson(value) };
}

export const financeApi = {
  // Dashboard
  getDashboard: () =>
    compatResult<DashboardStats>(() => financeCompatibilityClient.getDashboard({})),

  // Invoices
  getInvoices: (params?: {
    page?: number;
    pageSize?: number;
    limit?: number;
    offset?: number;
    status?: string;
    start?: string;
    end?: string;
  }) => compatResult<Invoice[]>(() => financeCompatibilityClient.listInvoices(params ?? {})),

  getInvoiceDetail: (id: string) =>
    compatResult<Invoice>(() => financeCompatibilityClient.getInvoice({ id })),

  createInvoice: (data: CreateInvoiceRequest) =>
    compatResult<Invoice>(() => financeCompatibilityClient.createInvoice(body(data))),

  updateInvoice: (id: string, data: Partial<Invoice>) =>
    compatResult<Invoice>(() => financeCompatibilityClient.updateInvoice(body({ id, ...data }))),

  deleteInvoice: (id: string) =>
    generatedResult<unknown>(() => invoicesClient.deleteInvoice({ id })),

  updateInvoiceStatus: (id: string, status: Invoice['status']) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.updateInvoiceStatus(body({ id, status }))
    ),

  recordPayment: (id: string, amount: number, details?: Record<string, unknown>) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.recordPayment(
        body({ invoiceId: toNumericId(id), amount, ...details })
      )
    ),

  // Expenses
  getExpenses: (params?: {
    page?: number;
    pageSize?: number;
    limit?: number;
    offset?: number;
    category?: string;
    start?: string;
    end?: string;
  }) => compatResult<Expense[]>(() => financeCompatibilityClient.listExpenses(params ?? {})),

  createExpense: (data: Partial<Expense>) =>
    compatResult<Expense>(() =>
      financeCompatibilityClient.createExpense(body(toExpenseMutationPayload(data)))
    ),

  updateExpense: (id: string, data: Partial<Expense>) =>
    compatResult<Expense>(() => financeCompatibilityClient.updateExpense(body({ id, ...data }))),

  deleteExpense: (id: string) =>
    generatedResult<unknown>(() => expensesClient.deleteExpense({ id })),

  // Expense Categories
  getExpenseCategories: () =>
    compatResult<ExpenseCategory[]>(() => financeCompatibilityClient.listExpenseCategories({})),

  createExpenseCategory: (data: Partial<ExpenseCategory>) =>
    compatResult<ExpenseCategory>(() =>
      financeCompatibilityClient.createExpenseCategory(body(toExpenseCategoryMutationPayload(data)))
    ),

  updateExpenseCategory: (id: number, data: Partial<ExpenseCategory>) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.updateExpenseCategory({ id: String(id), ...body(data) })
    ),

  deleteExpenseCategory: (id: number) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.deleteExpenseCategory({ id: String(id) })
    ),

  // P&L
  getPL: (period: string) =>
    compatResult<PLReport>(() =>
      financeCompatibilityClient.getPL(
        plReportQueryParams({ period, date: new Date().toISOString().slice(0, 10) })
      )
    ),

  // Tax
  getTaxConfig: () => compatResult<TaxConfig[]>(() => financeCompatibilityClient.listTaxConfig({})),

  createTaxConfig: (data: TaxConfig) =>
    compatResult<TaxConfig>(() =>
      financeCompatibilityClient.createTaxConfig(body(toTaxConfigPayload(data)))
    ),

  getTaxReport: () => compatResult<TaxReport>(() => financeCompatibilityClient.getTaxReport({})),

  updateTaxConfig: (data: Record<string, unknown>) =>
    compatResult<unknown>(() => financeCompatibilityClient.updateTaxConfig(body(data))),

  // Recurring Invoices
  getRecurringInvoices: () =>
    compatResult<unknown>(() => financeCompatibilityClient.listRecurringInvoices({})),

  createRecurringInvoice: (data: CreateRecurringInvoiceRequest) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.createRecurringInvoice(body(toRecurringInvoicePayload(data)))
    ),

  updateRecurringInvoice: (id: string, data: Partial<CreateRecurringInvoiceRequest>) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.updateRecurringInvoice(body({ id, ...data }))
    ),

  triggerRecurringInvoice: (id: string) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.triggerRecurringInvoice({ id, body: {} })
    ),

  // Quotes
  getQuotes: () => compatResult<unknown>(() => financeCompatibilityClient.listQuotes({})),

  createQuote: (data: QuotePayload) =>
    compatResult<unknown>(() => financeCompatibilityClient.createQuote(body(data))),

  updateQuote: (id: string, data: Partial<QuotePayload>) =>
    compatResult<unknown>(() => financeCompatibilityClient.updateQuote(body({ id, ...data }))),

  deleteQuote: (id: string) =>
    compatResult<unknown>(() => financeCompatibilityClient.deleteQuote({ id })),

  // Bills
  getBills: () => compatResult<unknown>(() => financeCompatibilityClient.listBills({})),

  createBill: (data: CreateBillRequest) =>
    compatResult<unknown>(() => financeCompatibilityClient.createBill(body(data))),

  updateBill: (id: string, data: Partial<CreateBillRequest>) =>
    compatResult<unknown>(() => financeCompatibilityClient.updateBill(body({ id, ...data }))),

  // Bank Accounts
  getBankAccounts: () =>
    compatResult<unknown>(() => financeCompatibilityClient.listBankAccounts({})),

  createBankAccount: (data: CreateBankAccountRequest) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.createBankAccount(body(toBankAccountPayload(data)))
    ),

  // Reconciliation
  getReconciliation: () =>
    compatResult<unknown>(() => financeCompatibilityClient.getReconciliation({})),
  matchTransaction: (
    bankFeedId: string | number,
    targetType: 'invoice' | 'expense',
    targetId: string | number
  ) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.matchTransaction(
        body({ bankFeedId: Number(bankFeedId), targetType, targetId: Number(targetId) })
      )
    ),

  matchReconcile: (transactionId: string, targetId: string) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.matchReconcile(
        body({ transactionId, targetType: 'invoice', targetId })
      )
    ),

  getBankFeeds: (accountId?: string) =>
    compatResult<unknown>(() => financeCompatibilityClient.listBankFeeds({ accountId })),

  // Balance Sheet
  getBalanceSheet: () =>
    compatResult<unknown>(() => financeCompatibilityClient.getBalanceSheet({})),

  // Cash Flow
  getCashFlow: () => compatResult<unknown>(() => financeCompatibilityClient.getCashFlow({})),

  // Chart of Accounts
  getAccounts: () => compatResult<unknown>(() => financeCompatibilityClient.listAccounts({})),

  createAccount: (data: CreateAccountRequest) =>
    compatResult<unknown>(() => financeCompatibilityClient.createAccount(body(data))),

  // Journal
  getJournal: () => compatResult<unknown>(() => financeCompatibilityClient.getJournal({})),

  createJournalEntry: (data: CreateJournalEntryRequest) =>
    compatResult<unknown>(() => financeCompatibilityClient.createJournalEntry(body(data))),

  // Expense detail
  getExpenseDetail: (id: string) =>
    compatResult<Expense>(() => financeCompatibilityClient.listExpenses({ id })),
};
