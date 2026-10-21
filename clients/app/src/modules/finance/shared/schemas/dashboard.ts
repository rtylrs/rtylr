import { z } from 'zod';

// int64 fields arrive as strings from the generated client and money as
// numbers; both are normalised to finite numbers, anything else to zero.
const finiteNumber = z.preprocess((value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}, z.number());

const monthSchema = z.object({
  month: z.string().default(''),
  revenue: finiteNumber.default(0),
  expenses: finiteNumber.default(0),
});

const activitySchema = z.object({
  kind: z.enum(['expense', 'payment']).catch('expense'),
  id: z.union([z.string(), z.number()]).transform(String).default(''),
  description: z.string().default(''),
  amount: finiteNumber.default(0),
  date: z.string().default(''),
  status: z.string().default(''),
});

const overdueInvoiceSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String).default(''),
  invoiceNumber: z.string().default(''),
  clientName: z.string().default(''),
  amountDue: finiteNumber.default(0),
  daysOverdue: finiteNumber.default(0),
});

/** Mirrors finance.dashboard.v1 GetFinanceDashboardResponse. */
export const dashboardStatsSchema = z.object({
  totalRevenueMtd: finiteNumber.default(0),
  totalExpensesMtd: finiteNumber.default(0),
  netProfitMtd: finiteNumber.default(0),
  totalReceivable: finiteNumber.default(0),
  overdueAmount: finiteNumber.default(0),
  pendingExpenses: finiteNumber.default(0),
  pendingExpenseCount: finiteNumber.default(0),
  monthly: z.array(monthSchema).default([]),
  recentActivity: z.array(activitySchema).default([]),
  overdueInvoices: z.array(overdueInvoiceSchema).default([]),
  overdueInvoiceCount: finiteNumber.default(0),
});

export type DashboardStats = z.infer<typeof dashboardStatsSchema>;
export type DashboardActivity = DashboardStats['recentActivity'][number];
export type DashboardMonth = DashboardStats['monthly'][number];
