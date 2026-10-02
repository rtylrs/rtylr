import { createOneKitClient, createRestRouteClient, oneKitResult, unwrap } from '@voxire/api';
import { UserServiceClient } from '@voxire/voxie/hub/user/v1/client';
import { MembersServiceClient } from '@voxire/voxie/hub/team/members/v1/client';
import { InvitationsServiceClient } from '@voxire/voxie/hub/team/invitations/v1/client';
import { FinancialServiceClient } from '@voxire/voxie/insights/financial/v1/client';
import { CustomerServiceClient } from '@voxire/voxie/crm/customer/v1/client';
import { DashboardServiceClient as FinanceDashboardServiceClient } from '@voxire/voxie/finance/dashboard/v1/client';
import { AccountServiceClient } from '@voxire/voxie/finance/account/v1/client';
import { BankAccountServiceClient } from '@voxire/voxie/finance/bank_account/v1/client';
import { BillServiceClient } from '@voxire/voxie/finance/bill/v1/client';
import { ExpensesServiceClient } from '@voxire/voxie/finance/expenses/v1/client';
import { InvoicesServiceClient } from '@voxire/voxie/finance/invoices/v1/client';
import { InvoicePaymentServiceClient } from '@voxire/voxie/finance/invoices/payments/v1/client';
import { JournalServiceClient } from '@voxire/voxie/finance/journal/v1/client';
import { QuoteServiceClient } from '@voxire/voxie/finance/quote/v1/client';
import { ReconciliationServiceClient } from '@voxire/voxie/finance/reconciliation/v1/client';
import { RecurringInvoiceServiceClient } from '@voxire/voxie/finance/recurring_invoice/v1/client';
import { ReportServiceClient } from '@voxire/voxie/finance/report/v1/client';
import { StatementServiceClient } from '@voxire/voxie/finance/statement/v1/client';
import { TaxServiceClient } from '@voxire/voxie/finance/tax/v1/client';

export const userClient = createOneKitClient(UserServiceClient, 'finance');
export const financeDashboardClient = createOneKitClient(FinanceDashboardServiceClient, 'finance');
export const invoicesClient = createOneKitClient(InvoicesServiceClient, 'finance');
export const expensesClient = createOneKitClient(ExpensesServiceClient, 'finance');

export const teamMembersClient = createOneKitClient(MembersServiceClient, 'finance');
export const teamInvitationsClient = createOneKitClient(InvitationsServiceClient, 'finance');
export const insightsFinancialClient = createOneKitClient(FinancialServiceClient, 'finance');
export const crmCustomerClient = createOneKitClient(CustomerServiceClient, 'finance');

const accountClient = createOneKitClient(AccountServiceClient, 'finance');
const bankAccountClient = createOneKitClient(BankAccountServiceClient, 'finance');
const billClient = createOneKitClient(BillServiceClient, 'finance');
const invoicePaymentClient = createOneKitClient(InvoicePaymentServiceClient, 'finance');
const journalClient = createOneKitClient(JournalServiceClient, 'finance');
const quoteClient = createOneKitClient(QuoteServiceClient, 'finance');
const reconciliationClient = createOneKitClient(ReconciliationServiceClient, 'finance');
const recurringInvoiceClient = createOneKitClient(RecurringInvoiceServiceClient, 'finance');
const reportClient = createOneKitClient(ReportServiceClient, 'finance');
const statementClient = createOneKitClient(StatementServiceClient, 'finance');
const taxClient = createOneKitClient(TaxServiceClient, 'finance');
const rest = createRestRouteClient('finance');

/**
 * c6e4ce609 routed 41 of this app's calls through a single
 * `@voxire/voxie/finance/compat/v1` client. That module has never existed at any
 * voxie commit, so the app could not be type-checked or built (#894).
 *
 * It was never needed. Every capability behind it already had its own contract:
 * account, bank_account, bill, expenses, invoices, invoices/payments, journal,
 * quote, reconciliation, recurring_invoice, report, statement and tax. The
 * umbrella was invented alongside thirteen real services.
 *
 * This object keeps the same 41 method names so no call site had to change, and
 * dispatches each to the service that actually declares it. The name is kept for
 * the same reason; it should shrink to nothing as call sites move to the real
 * client directly, which is a mechanical follow-up rather than part of this fix.
 *
 * Method names differ from the invented ones in a few places, and where they do
 * the contract wins: `getJournal` is `listJournalEntries`, `getReconciliation`
 * is `getOverview`, `getPL` is `getPLReport`, and both `matchTransaction` and
 * `matchReconcile` are the single `match`.
 */
export const financeCompatibilityClient = {
  // Accounts
  listAccounts: (...args: Parameters<typeof accountClient.listAccounts>) =>
    accountClient.listAccounts(...args),
  createAccount: (...args: Parameters<typeof accountClient.createAccount>) =>
    accountClient.createAccount(...args),

  // Bank accounts
  listBankAccounts: (...args: Parameters<typeof bankAccountClient.listBankAccounts>) =>
    bankAccountClient.listBankAccounts(...args),
  createBankAccount: (...args: Parameters<typeof bankAccountClient.createBankAccount>) =>
    bankAccountClient.createBankAccount(...args),

  // Bills
  listBills: (...args: Parameters<typeof billClient.listBills>) => billClient.listBills(...args),
  createBill: (...args: Parameters<typeof billClient.createBill>) => billClient.createBill(...args),
  updateBill: (...args: Parameters<typeof billClient.updateBill>) => billClient.updateBill(...args),

  // Expenses
  listExpenses: (...args: Parameters<typeof expensesClient.listExpenses>) =>
    expensesClient.listExpenses(...args),
  createExpense: (...args: Parameters<typeof expensesClient.createExpense>) =>
    expensesClient.createExpense(...args),
  updateExpense: (...args: Parameters<typeof expensesClient.updateExpense>) =>
    expensesClient.updateExpense(...args),

  // Expense categories: the read lives on expenses, the detail rows on report.
  listExpenseCategories: (...args: Parameters<typeof reportClient.listExpenseCategoryDetails>) =>
    reportClient.listExpenseCategoryDetails(...args),
  createExpenseCategory: (...args: Parameters<typeof reportClient.createExpenseCategoryDetail>) =>
    reportClient.createExpenseCategoryDetail(...args),

  // Invoices
  listInvoices: (...args: Parameters<typeof invoicesClient.listInvoices>) =>
    invoicesClient.listInvoices(...args),
  getInvoice: (...args: Parameters<typeof invoicesClient.getInvoice>) =>
    invoicesClient.getInvoice(...args),
  createInvoice: (...args: Parameters<typeof invoicesClient.createInvoice>) =>
    invoicesClient.createInvoice(...args),
  updateInvoice: (...args: Parameters<typeof invoicesClient.updateInvoice>) =>
    invoicesClient.updateInvoice(...args),
  updateInvoiceStatus: (...args: Parameters<typeof invoicesClient.updateInvoiceStatus>) =>
    invoicesClient.updateInvoiceStatus(...args),
  recordPayment: (...args: Parameters<typeof invoicePaymentClient.recordPayment>) =>
    invoicePaymentClient.recordPayment(...args),

  // Journal
  getJournal: (...args: Parameters<typeof journalClient.listJournalEntries>) =>
    journalClient.listJournalEntries(...args),
  createJournalEntry: (...args: Parameters<typeof journalClient.createJournalEntry>) =>
    journalClient.createJournalEntry(...args),

  // Quotes
  listQuotes: (...args: Parameters<typeof quoteClient.listQuotes>) =>
    quoteClient.listQuotes(...args),
  createQuote: (...args: Parameters<typeof quoteClient.createQuote>) =>
    quoteClient.createQuote(...args),

  // Reconciliation
  getReconciliation: (...args: Parameters<typeof reconciliationClient.getOverview>) =>
    reconciliationClient.getOverview(...args),
  matchTransaction: (...args: Parameters<typeof reconciliationClient.match>) =>
    reconciliationClient.match(...args),
  matchReconcile: (...args: Parameters<typeof reconciliationClient.match>) =>
    reconciliationClient.match(...args),

  // Recurring invoices
  listRecurringInvoices: (
    ...args: Parameters<typeof recurringInvoiceClient.listRecurringInvoices>
  ) => recurringInvoiceClient.listRecurringInvoices(...args),
  createRecurringInvoice: (
    ...args: Parameters<typeof recurringInvoiceClient.createRecurringInvoice>
  ) => recurringInvoiceClient.createRecurringInvoice(...args),
  updateRecurringInvoice: (
    ...args: Parameters<typeof recurringInvoiceClient.updateRecurringInvoice>
  ) => recurringInvoiceClient.updateRecurringInvoice(...args),

  // Reports
  getPL: (...args: Parameters<typeof reportClient.getPLReport>) =>
    reportClient.getPLReport(...args),
  getTaxReport: (...args: Parameters<typeof reportClient.getTaxReport>) =>
    reportClient.getTaxReport(...args),
  listTaxConfig: (...args: Parameters<typeof reportClient.listTaxConfig>) =>
    reportClient.listTaxConfig(...args),
  createTaxConfig: (...args: Parameters<typeof reportClient.createTaxConfig>) =>
    reportClient.createTaxConfig(...args),

  // Statements
  getBalanceSheet: (...args: Parameters<typeof statementClient.getBalanceSheet>) =>
    statementClient.getBalanceSheet(...args),
  getCashFlow: (...args: Parameters<typeof statementClient.getCashFlow>) =>
    statementClient.getCashFlow(...args),

  // Tax
  updateTaxConfig: (...args: Parameters<typeof taxClient.updateTaxConfig>) =>
    taxClient.updateTaxConfig(...args),

  // Dashboard
  getDashboard: (...args: Parameters<typeof financeDashboardClient.getFinanceDashboard>) =>
    financeDashboardClient.getFinanceDashboard(...args),
  getExpense: (...args: Parameters<typeof expensesClient.getExpense>) =>
    expensesClient.getExpense(...args),

  // ── No contract declares these six ────────────────────────────────────────
  //
  // c6e4ce609 invented them along with the umbrella, and unlike the rest they
  // have no service to dispatch to. They go over REST at the paths onion
  // actually registers as literal handlers, which is verifiable:
  // /finance/expense-categories, /finance/quotes/{id} and /finance/bank-feeds
  // are all in onion's route table.

  updateExpenseCategory: (req: { id: string; body?: unknown }) =>
    unwrap(rest.put<unknown>('/finance/expense-categories', { id: req.id, ...(req.body ?? {}) })),

  deleteExpenseCategory: (req: { id: string }) =>
    unwrap(rest.delete<unknown>('/finance/expense-categories', { params: { id: req.id } })),

  updateQuote: (req: { id: string; body?: unknown }) =>
    unwrap(rest.put<unknown>(`/finance/quotes/${req.id}`, req.body ?? {})),

  deleteQuote: (req: { id: string }) => unwrap(rest.delete<unknown>(`/finance/quotes/${req.id}`)),

  listBankFeeds: (req?: Record<string, unknown>) =>
    unwrap(rest.get<unknown>('/finance/bank-feeds', { params: req ?? {} })),
};

export const listCrmCustomers = () => oneKitResult(() => crmCustomerClient.listCustomers({}));
