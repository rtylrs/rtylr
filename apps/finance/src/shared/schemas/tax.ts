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
});

export type TaxConfigRaw = z.infer<typeof taxConfigSchema>;
export type TaxReportRaw = z.infer<typeof taxReportSchema>;
export type PLReportRaw = z.infer<typeof plReportSchema>;
