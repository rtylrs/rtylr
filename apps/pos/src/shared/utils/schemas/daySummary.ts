import { z } from 'zod';

export const daySummarySchema = z.object({
  date: z.string().default(() => new Date().toISOString().slice(0, 10)),
  grossSales: z.number().default(0),
  netSales: z.number().default(0),
  taxCollected: z.number().default(0),
  tips: z.number().default(0),
  refunds: z.number().default(0),
  voids: z.number().default(0),
  orderCount: z.number().default(0),
  averageTicket: z.number().default(0),
  cashDrops: z.number().default(0),
  openingFloat: z.number().default(0),
  closingCash: z.number().default(0),
});

export type DaySummaryRaw = z.infer<typeof daySummarySchema>;
