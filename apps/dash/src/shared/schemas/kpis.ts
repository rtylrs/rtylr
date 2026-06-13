import { z } from 'zod';

// The actual /hub/kpis response only carries the first six fields. Requiring
// the rest made the whole query throw → every dashboard card rendered $0. All
// fields are now defaulted so parse never fails; the dashboard widgets read the
// real fields, while the legacy fields (no backend source yet) stay defaulted
// for the Analytics screen until they're wired on the backend.
export const dashKpisSchema = z.object({
  // Real backend fields:
  revenue: z.number().default(0),
  orderCount: z.number().default(0),
  staffCount: z.number().default(0),
  customerCount: z.number().default(0),
  lowStockCount: z.number().default(0),
  activeApps: z.number().default(0),
  // Legacy fields with no backend source yet (default so consumers don't break):
  revenueToday: z.number().default(0),
  revenueTrendPct: z.number().default(0),
  ordersToday: z.number().default(0),
  ordersTrendPct: z.number().default(0),
  headcount: z.number().default(0),
  activeStaff: z.number().default(0),
  lowStockItems: z.number().default(0),
  openInvoices: z.number().default(0),
  newCustomers: z.number().default(0),
  apiHealthPct: z.number().default(0),
  revenueSeries: z.array(z.number()).default([]),
  ordersSeries: z.array(z.number()).default([]),
});

export type HubKpis = z.infer<typeof dashKpisSchema>;
