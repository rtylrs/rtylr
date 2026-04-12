import { z } from 'zod';

export const storeDashboardStatsSchema = z.object({
  totalProducts: z.number(),
  ordersToday: z.number(),
  revenueToday: z.number(),
  activeVisitors: z.number(),
});

export type StoreDashboardStatsParsed = z.infer<typeof storeDashboardStatsSchema>;
