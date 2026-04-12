import { z } from 'zod';

export const productPerformanceSchema = z.object({
  totalProducts: z.number().optional(),
  avgRevenue: z.number().optional(),
  topCategory: z.string().optional().nullable(),
});

export const productRankingSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String).optional(),
  name: z.string().optional(),
  sku: z.string().optional().nullable(),
  revenue: z.number().optional(),
  unitsSold: z.number().optional(),
  rank: z.number().optional(),
  data: z.array(z.unknown()).optional(),
});

export const marginAnalysisSchema = z.object({
  avgMargin: z.number().optional(),
  averageMargin: z.number().optional(),
  totalRevenue: z.number().optional(),
  totalCogs: z.number().optional(),
  grossProfit: z.number().optional(),
});

export const categoryBreakdownSchema = z.object({
  category: z.string().optional(),
  revenue: z.number().optional(),
  pct: z.number().optional(),
  units: z.number().optional(),
  data: z.array(z.unknown()).optional(),
});

export type ProductPerformance = z.infer<typeof productPerformanceSchema>;
export type ProductRanking = z.infer<typeof productRankingSchema>;
export type MarginAnalysis = z.infer<typeof marginAnalysisSchema>;
export type CategoryBreakdown = z.infer<typeof categoryBreakdownSchema>;
