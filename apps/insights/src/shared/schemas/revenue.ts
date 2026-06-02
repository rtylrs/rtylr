import { z } from 'zod';

export const revenueDataSchema = z.object({
  period: z.string().optional(),
  total: z.number().optional(),
  change: z.number().optional(),
  changePct: z.number().optional(),
  data: z.array(z.object({ date: z.string(), amountMinorUnits: z.number() })).optional(),
  series: z.array(z.object({ date: z.string(), amountMinorUnits: z.number() })).optional(),
});

export const revenueByAppSchema = z.object({
  app: z.string().optional(),
  amountMinorUnits: z.number().optional(),
  pct: z.number().optional(),
  data: z.array(z.unknown()).optional(),
});

export const revenueComparisonSchema = z.object({
  current: z.number().optional(),
  previous: z.number().optional(),
  change: z.number().optional(),
  changePercent: z.number().optional(),
  changePct: z.number().optional(),
  difference: z.number().optional(),
});

export type RevenueData = z.infer<typeof revenueDataSchema>;
export type RevenueByApp = z.infer<typeof revenueByAppSchema>;
export type RevenueComparison = z.infer<typeof revenueComparisonSchema>;
