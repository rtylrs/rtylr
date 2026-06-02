import { z } from 'zod';

export const customerAnalyticsSchema = z.object({
  total: z.number().optional(),
  newThisMonth: z.number().optional(),
  active: z.number().optional(),
  churned: z.number().optional(),
});

export const topCustomerSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String).optional(),
  name: z.string().optional(),
  email: z.string().optional().nullable(),
  totalSpend: z.number().optional(),
  orderCount: z.number().optional(),
  lastOrderAt: z.string().optional().nullable(),
  data: z.array(z.unknown()).optional(),
});

export const customerChurnSchema = z.object({
  rate: z.number().optional(),
  churnRate: z.number().optional(),
  churned: z.number().optional(),
  atRisk: z.number().optional(),
  retained: z.number().optional(),
});

export const customerLtvSchema = z.object({
  average: z.number().optional(),
  avgLtv: z.number().optional(),
  medianLtv: z.number().optional(),
  topDecileLtv: z.number().optional(),
});

export type CustomerAnalytics = z.infer<typeof customerAnalyticsSchema>;
export type TopCustomer = z.infer<typeof topCustomerSchema>;
export type CustomerChurn = z.infer<typeof customerChurnSchema>;
export type CustomerLtv = z.infer<typeof customerLtvSchema>;
