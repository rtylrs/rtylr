import { z } from 'zod';

export const hrDashboardStatsSchema = z.object({
  totalStaff: z.number(),
  onShiftToday: z.number(),
  openPositions: z.number(),
  monthlyPayroll: z.number(),
  recentActivity: z
    .array(
      z.object({
        name: z.string(),
        action: z.string(),
        time: z.string(),
        type: z.enum(['in', 'out']).catch('in'),
        role: z.string(),
      })
    )
    .default([]),
  todayAtGlance: z.object({
    shiftsScheduled: z.number(),
    newApplications: z.number(),
    pendingApprovals: z.number(),
  }),
});

export type HRDashboardStats = z.infer<typeof hrDashboardStatsSchema>;
