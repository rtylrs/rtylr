import { z } from 'zod';

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
  daysOverdue: z.number().default(0),
});

const barDataPointSchema = z.object({
  month: z.string().default(''),
  revenue: z.number().default(0),
  expenses: z.number().default(0),
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
export const dashboardStatsSchema = z.object({
  totalReceivable: z.number().default(0),
  overdueAmount: z.number().default(0),
  totalExpensesMtd: z.number().default(0),
  netProfitMtd: z.number().default(0),

  // Optional / not yet sent by the backend.
  outstandingInvoicesCount: z.number().default(0),
  totalRevenueMtdChange: z.string().default('+0%'),
  totalExpensesMtdChange: z.string().default('+0%'),
  netProfitChange: z.string().default('+0%'),
  recentTransactions: z.array(transactionSchema).default([]),
  overdueInvoices: z.array(overdueInvoiceSchema).default([]),
  barData: z.array(barDataPointSchema).default([]),
});

export type DashboardStatsRaw = z.infer<typeof dashboardStatsSchema>;
