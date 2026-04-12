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

export const dashboardStatsSchema = z.object({
  totalRevenueMTD: z.number().default(0),
  totalExpensesMTD: z.number().default(0),
  netProfit: z.number().default(0),
  outstandingInvoicesTotal: z.number().default(0),
  outstandingInvoicesCount: z.number().default(0),
  totalRevenueMTDChange: z.string().default('0%'),
  totalExpensesMTDChange: z.string().default('0%'),
  netProfitChange: z.string().default('0%'),
  recentTransactions: z.array(transactionSchema).default([]),
  overdueInvoices: z.array(overdueInvoiceSchema).default([]),
  barData: z.array(barDataPointSchema).default([]),
});

export type DashboardStatsRaw = z.infer<typeof dashboardStatsSchema>;
