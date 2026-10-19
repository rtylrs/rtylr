import { fromOneKitJson, oneKitResult, toOneKitJson } from '@voxire/api';
import type { InvoiceItem as InvoiceLine } from '@voxire/voxie/finance/invoices/v1/types';
import type { CreateInvoiceRequest } from '@voxire/voxie/finance/invoices/v1/types';
import type { CreateRecurringInvoiceRequest } from '@voxire/voxie/finance/recurring_invoice/v1/types';
import type { CreateBillRequest } from '@voxire/voxie/finance/bill/v1/types';
import type { CreateBankAccountRequest } from '@voxire/voxie/finance/bank_account/v1/types';
import type { CreateAccountRequest } from '@voxire/voxie/finance/account/v1/types';
import type { CreateJournalEntryRequest } from '@voxire/voxie/finance/journal/v1/types';
import { expensesClient, financeCompatibilityClient, invoicesClient } from '../lib/clients';
import { todayIso } from '../lib/date';

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
  const date = params?.date ?? to ?? from ?? todayIso();
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
  discountType?: string | null;
  discountAmount?: number;
  tax?: number;
  subtotal?: number;
  number?: string;
  history?: { id?: string; label: string; at: string }[];
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

/**
 * Wraps a mutation payload for a OneKit `@body("body")` request.
 *
 * The generic is an assertion, not a check. toOneKitJson only renames keys to
 * the wire's snake_case, so the value's shape is whatever the caller passed;
 * there is nothing here that could validate it against the request type. It was
 * previously `{ body: unknown }`, which compiled against the invented compat
 * client because that client's parameters were loose. Now that these calls reach
 * the real generated requests, the assertion is what carries the shape - and the
 * generated types are at least checking the surrounding fields.
 */
function flat<T>(value: unknown): T {
  return toOneKitJson(value) as T;
}

function withId<B>(id: string, value: unknown): { id: string; body: B } {
  return { id, body: toOneKitJson(value) as B };
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
    compatResult<Invoice>(() => financeCompatibilityClient.createInvoice(flat(data))),

  updateInvoice: (id: string, data: Partial<Invoice>) =>
    compatResult<Invoice>(() => financeCompatibilityClient.updateInvoice(withId(id, data))),

  deleteInvoice: (id: string) =>
    generatedResult<unknown>(() => invoicesClient.deleteInvoice({ id })),

  updateInvoiceStatus: (id: string, status: Invoice['status']) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.updateInvoiceStatus(withId(id, { status }))
    ),

  recordPayment: (id: string, amount: number, details?: Record<string, unknown>) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.recordPayment(
        flat({ invoiceId: toNumericId(id), amount, ...details })
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
      financeCompatibilityClient.createExpense(flat(toExpenseMutationPayload(data)))
    ),

  updateExpense: (id: string, data: Partial<Expense>) =>
    compatResult<Expense>(() => financeCompatibilityClient.updateExpense(withId(id, data))),

  deleteExpense: (id: string) =>
    generatedResult<unknown>(() => expensesClient.deleteExpense({ id })),

  // P&L
  getPL: (period: string) =>
    compatResult<PLReport>(() =>
      financeCompatibilityClient.getPL(plReportQueryParams({ period, date: todayIso() }))
    ),

  // Tax
  getTaxConfig: () => compatResult<TaxConfig[]>(() => financeCompatibilityClient.listTaxConfig({})),

  createTaxConfig: (data: TaxConfig) =>
    compatResult<TaxConfig>(() =>
      financeCompatibilityClient.createTaxConfig(flat(toTaxConfigPayload(data)))
    ),

  getTaxReport: () =>
    compatResult<TaxReport>(() => financeCompatibilityClient.getTaxReport({ date: todayIso() })),

  updateTaxConfig: (data: Record<string, unknown>) =>
    compatResult<unknown>(() => financeCompatibilityClient.updateTaxConfig(flat(data))),

  // Recurring Invoices
  getRecurringInvoices: () =>
    compatResult<unknown>(() => financeCompatibilityClient.listRecurringInvoices({})),

  createRecurringInvoice: (data: CreateRecurringInvoiceRequest) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.createRecurringInvoice(flat(toRecurringInvoicePayload(data)))
    ),

  updateRecurringInvoice: (id: string, data: Partial<CreateRecurringInvoiceRequest>) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.updateRecurringInvoice(withId(id, data))
    ),

  // Quotes
  getQuotes: () => compatResult<unknown>(() => financeCompatibilityClient.listQuotes({})),

  createQuote: (data: QuotePayload) =>
    compatResult<unknown>(() => financeCompatibilityClient.createQuote(flat(data))),

  updateQuote: (id: string, data: Partial<QuotePayload>) =>
    compatResult<unknown>(() => financeCompatibilityClient.updateQuote(withId(id, data))),

  deleteQuote: (id: string) =>
    compatResult<unknown>(() => financeCompatibilityClient.deleteQuote({ id })),

  // Bills
  getBills: () => compatResult<unknown>(() => financeCompatibilityClient.listBills({})),

  createBill: (data: CreateBillRequest) =>
    compatResult<unknown>(() => financeCompatibilityClient.createBill(flat(data))),

  updateBill: (id: string, data: Partial<CreateBillRequest>) =>
    compatResult<unknown>(() => financeCompatibilityClient.updateBill(withId(id, data))),

  // Bank Accounts
  getBankAccounts: () =>
    compatResult<unknown>(() => financeCompatibilityClient.listBankAccounts({})),

  createBankAccount: (data: CreateBankAccountRequest) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.createBankAccount(flat(toBankAccountPayload(data)))
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
        flat({ bankFeedId: Number(bankFeedId), targetType, targetId: Number(targetId) })
      )
    ),

  matchReconcile: (transactionId: string, targetId: string) =>
    compatResult<unknown>(() =>
      financeCompatibilityClient.matchReconcile(
        flat({ transactionId, targetType: 'invoice', targetId })
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
    compatResult<unknown>(() => financeCompatibilityClient.createAccount(flat(data))),

  // Journal
  getJournal: () => compatResult<unknown>(() => financeCompatibilityClient.getJournal({})),

  createJournalEntry: (data: CreateJournalEntryRequest) =>
    compatResult<unknown>(() => financeCompatibilityClient.createJournalEntry(flat(data))),

  // Expense detail
  getExpenseDetail: (id: string) =>
    compatResult<Expense>(() => financeCompatibilityClient.getExpense({ id })),
};
