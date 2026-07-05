import { z } from 'zod';

export const profitLossSchema = z.object({
  revenue: z.number().optional(),
  totalRevenue: z.number().optional(),
  expenses: z.number().optional(),
  totalExpenses: z.number().optional(),
  cogs: z.number().optional(),
  grossProfit: z.number().optional(),
  gross_profit: z.number().optional(),
  operatingExpenses: z.number().optional(),
  operating_expenses: z.number().optional(),
  netProfit: z.number().optional(),
  net_profit: z.number().optional(),
  margin: z.number().optional(),
  margin_pct: z.number().optional(),
});

export const expenseBreakdownSchema = z.object({
  category: z.string().optional(),
  amount: z.number().optional(),
  pct: z.number().optional(),
  change: z.number().optional(),
  data: z.array(z.unknown()).optional(),
  categories: z.array(z.unknown()).optional(),
});

export const invoiceAgingSchema = z.object({
  current: z.number().optional(),
  days30: z.number().optional(),
  days60: z.number().optional(),
  days90: z.number().optional(),
  over90: z.number().optional(),
  data: z.array(z.unknown()).optional(),
  buckets: z.array(z.unknown()).optional(),
});

const cashFlowPointSchema = z.object({
  date: z.string().optional(),
  inflowMinorUnits: z.number().optional(),
  outflowMinorUnits: z.number().optional(),
});

export const cashFlowSchema = z.object({
  period: z.string().optional(),
  inflow: z.number().optional(),
  outflow: z.number().optional(),
  net: z.number().optional(),
  data: z.array(cashFlowPointSchema).optional(),
  periods: z.array(cashFlowPointSchema).optional(),
});

export type ProfitLoss = z.infer<typeof profitLossSchema>;
export type ExpenseBreakdown = z.infer<typeof expenseBreakdownSchema>;
export type InvoiceAging = z.infer<typeof invoiceAgingSchema>;
export type CashFlow = z.infer<typeof cashFlowSchema>;
