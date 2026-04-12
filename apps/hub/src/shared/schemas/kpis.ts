import { z } from 'zod';

export const hubKpisSchema = z.object({
  revenueToday: z.number(),
  revenueTrendPct: z.number(),
  ordersToday: z.number(),
  ordersTrendPct: z.number(),
  headcount: z.number(),
  activeStaff: z.number(),
  lowStockItems: z.number(),
  openInvoices: z.number(),
  newCustomers: z.number(),
  apiHealthPct: z.number(),
  revenueSeries: z.array(z.number()),
  ordersSeries: z.array(z.number()),
});

export type HubKpis = z.infer<typeof hubKpisSchema>;
