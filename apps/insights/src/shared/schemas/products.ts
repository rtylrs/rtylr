import { z } from 'zod';

export const productPerformanceSchema = z.object({
  totalSkus: z.number().optional(),
  activeSkus: z.number().optional(),
  outOfStock: z.number().optional(),
  avgMarginPct: z.number().optional(),
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
  categoryId: z.union([z.string(), z.number()]).transform(String).optional(),
  categoryName: z.string().optional(),
  unitsSold: z.number().optional(),
  revenueMinorUnits: z.number().optional(),
  data: z.array(z.unknown()).optional(),
});

export type ProductPerformance = z.infer<typeof productPerformanceSchema>;
export type ProductRanking = z.infer<typeof productRankingSchema>;
export type MarginAnalysis = z.infer<typeof marginAnalysisSchema>;
export type CategoryBreakdown = z.infer<typeof categoryBreakdownSchema>;
