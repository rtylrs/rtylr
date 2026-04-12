import { z } from 'zod';

export const insightsKpisSchema = z.object({
  // Fields accessed by Dashboard.tsx
  totalRevenue: z.number().optional(),
  ordersThisMonth: z.number().optional(),
  customerCount: z.number().optional(),
  staffCount: z.number().optional(),
  avgOrderValue: z.number().optional(),
  topProduct: z
    .object({ name: z.string().optional(), revenue: z.number().optional() })
    .optional()
    .nullable(),
  // Additional hub KPI fields
  revenueToday: z.number().optional(),
  revenueTrendPct: z.number().optional(),
  ordersToday: z.number().optional(),
  ordersTrendPct: z.number().optional(),
  headcount: z.number().optional(),
  activeStaff: z.number().optional(),
  lowStockItems: z.number().optional(),
  openInvoices: z.number().optional(),
  newCustomers: z.number().optional(),
  apiHealthPct: z.number().optional(),
  revenueSeries: z.array(z.number()).optional(),
  ordersSeries: z.array(z.number()).optional(),
});

export const insightsDashboardSummarySchema = z.object({
  totalRevenue: z.number().optional(),
  ordersThisMonth: z.number().optional(),
  customerCount: z.number().optional(),
  staffCount: z.number().optional(),
  avgOrderValue: z.number().optional(),
  topProduct: z
    .object({ name: z.string().optional(), revenue: z.number().optional() })
    .optional()
    .nullable(),
  revenueChange: z.number().optional(),
  ordersChange: z.number().optional(),
  newCustomers: z.number().optional(),
  customersChange: z.number().optional(),
  aovChange: z.number().optional(),
});

export type InsightsKpis = z.infer<typeof insightsKpisSchema>;
export type InsightsDashboardSummary = z.infer<typeof insightsDashboardSummarySchema>;
