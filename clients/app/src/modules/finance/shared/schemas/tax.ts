import { z } from 'zod';

export const taxConfigSchema = z.object({
  id: z.string().optional().nullable(),
  name: z.string().default(''),
  rate: z.number().default(0),
});

export const taxReportSchema = z.object({
  period: z.string().default(''),
  taxCollected: z.number().default(0),
  taxOwed: z.number().default(0),
});

export const plReportSchema = z.object({
  period: z.string().default(''),
  revenue: z.number().default(0),
  cogs: z.number().default(0),
  operatingExpenses: z.number().default(0),
  grossProfit: z.number().default(0),
  netProfit: z.number().default(0),
  expensesByCategory: z
    .array(
      z.object({
        categoryId: z.coerce.number().optional(),
        name: z.string().default(''),
        color: z.string().default(''),
        amount: z.number().default(0),
      })
    )
    .default([]),
  pendingExpenses: z.number().default(0),
  pendingExpenseCount: z.coerce.number().default(0),
});
