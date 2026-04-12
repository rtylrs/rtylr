import { z } from 'zod';

export const bookDashboardStatsSchema = z.object({
  todayCount: z.number().optional(),
  todayChange: z.string().optional(),
  upcomingCount: z.number().optional(),
  upcomingChange: z.string().optional(),
  cancellationRate: z.number().optional(),
  cancellationChange: z.string().optional(),
  utilization: z.number().optional(),
  utilizationChange: z.string().optional(),
  totalAppointments: z.number().optional(),
  newCustomers: z.number().optional(),
  revenue: z.number().optional(),
  revenueToday: z.number().optional(),
});

export type BookDashboardStats = z.infer<typeof bookDashboardStatsSchema>;
