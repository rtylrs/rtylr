import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}, z.number());

const transactionSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  description: z.string().default(''),
  amount: z.string().default('0'),
  type: z.enum(['credit', 'debit']).catch('credit'),
  date: z.string().default(''),
});

const overdueInvoiceSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  client: z.string().default(''),
  amount: z.string().default('0'),
  daysOverdue: finiteNumber,
});

const barDataPointSchema = z.object({
  month: z.string().default(''),
  revenue: finiteNumber,
  expenses: finiteNumber,
});

/**
 * Mirrors the DEPLOYED onion finance DashboardResponse
 * (invoices module, dashboard_handler.go): total_receivable, overdue_amount,
 * total_expenses_mtd, net_profit_mtd -> camelCased by the @voxire/api interceptor.
 *
 * The richer fields (revenue/profit changes, recent transactions, overdue list,
 * bar series) are not yet returned by the backend; they remain optional with
 * safe defaults so the dashboard degrades gracefully until the backend is
 * extended. See the backend followup to populate them server-side.
 */
export const dashboardStatsSchema = z
  .object({
    totalReceivable: finiteNumber.optional(),
    total_receivable: finiteNumber.optional(),
    overdueAmount: finiteNumber.optional(),
    overdue_amount: finiteNumber.optional(),
    totalExpensesMtd: finiteNumber.optional(),
    total_expenses_mtd: finiteNumber.optional(),
    totalRevenueMtd: finiteNumber.optional(),
    total_revenue_mtd: finiteNumber.optional(),
    netProfitMtd: finiteNumber.optional(),
    net_profit_mtd: finiteNumber.optional(),

    // Optional / not yet sent by the backend.
    outstandingInvoicesCount: finiteNumber.optional(),
    outstanding_invoices_count: finiteNumber.optional(),
    totalRevenueMtdChange: z.string().optional(),
    total_revenue_mtd_change: z.string().optional(),
    totalExpensesMtdChange: z.string().optional(),
    total_expenses_mtd_change: z.string().optional(),
    netProfitChange: z.string().optional(),
    net_profit_change: z.string().optional(),
    recentTransactions: z.array(transactionSchema).optional(),
    recent_transactions: z.array(transactionSchema).optional(),
    overdueInvoices: z.array(overdueInvoiceSchema).optional(),
    overdue_invoices: z.array(overdueInvoiceSchema).optional(),
    barData: z.array(barDataPointSchema).optional(),
    bar_data: z.array(barDataPointSchema).optional(),
  })
  .transform((raw) => ({
    totalReceivable: raw.totalReceivable ?? raw.total_receivable ?? 0,
    overdueAmount: raw.overdueAmount ?? raw.overdue_amount ?? 0,
    totalExpensesMtd: raw.totalExpensesMtd ?? raw.total_expenses_mtd ?? 0,
    totalRevenueMtd: raw.totalRevenueMtd ?? raw.total_revenue_mtd ?? 0,
    netProfitMtd: raw.netProfitMtd ?? raw.net_profit_mtd ?? 0,
    outstandingInvoicesCount: raw.outstandingInvoicesCount ?? raw.outstanding_invoices_count ?? 0,
    totalRevenueMtdChange: raw.totalRevenueMtdChange ?? raw.total_revenue_mtd_change ?? '+0%',
    totalExpensesMtdChange: raw.totalExpensesMtdChange ?? raw.total_expenses_mtd_change ?? '+0%',
    netProfitChange: raw.netProfitChange ?? raw.net_profit_change ?? '+0%',
    recentTransactions: raw.recentTransactions ?? raw.recent_transactions ?? [],
    overdueInvoices: raw.overdueInvoices ?? raw.overdue_invoices ?? [],
    barData: raw.barData ?? raw.bar_data ?? [],
  }));
