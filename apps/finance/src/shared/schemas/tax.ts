import { z } from 'zod';

export const taxConfigSchema = z.object({
  id: z.number().optional().nullable(),
  name: z.string().default(''),
  rate: z.number().default(0),
});

const taxPeriodSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  period: z.string().default(''),
  revenue: z.number().default(0),
  taxableAmount: z.number().default(0),
  taxRate: z.number().default(0),
  collected: z.number().default(0),
  status: z.enum(['Filed', 'Pending', 'Due']).catch('Pending'),
  dueDate: z.string().default(''),
});

export const taxReportSchema = z.object({
  vatSummary: z
    .object({
      collected: z.number().default(0),
      owed: z.number().default(0),
      paid: z.number().default(0),
    })
    .default({ collected: 0, owed: 0, paid: 0 }),
  periods: z.array(taxPeriodSchema).default([]),
  configs: z.array(taxConfigSchema).default([]),
});

export const plReportSchema = z.object({
  period: z.string().default(''),
  revenue: z.number().default(0),
  cogs: z.number().default(0),
  operatingExpenses: z.number().default(0),
  grossProfit: z.number().default(0),
  netProfit: z.number().default(0),
});

export type TaxConfigRaw = z.infer<typeof taxConfigSchema>;
export type TaxReportRaw = z.infer<typeof taxReportSchema>;
export type PLReportRaw = z.infer<typeof plReportSchema>;
