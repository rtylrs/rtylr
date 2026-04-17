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

export const dashboardWidgetSchema = z.object({
  id: z.number(),
  ownerId: z.number(),
  roleId: z.number(),
  widgetType: z.string(),
  positionX: z.number(),
  positionY: z.number(),
  width: z.number(),
  height: z.number(),
  config: z.unknown().default({}),
  createdAt: z.number(),
  updatedAt: z.number(),
});

export const dashboardDataSchema = z.object({
  widgetType: z.string(),
  data: z.unknown().default({}),
});

export const getDashboardResponseSchema = z.object({
  widgets: z.array(dashboardWidgetSchema).default([]),
  data: z.record(z.string(), dashboardDataSchema).default({}),
});

export type DashboardWidget = z.infer<typeof dashboardWidgetSchema>;
export type DashboardData = z.infer<typeof dashboardDataSchema>;
export type GetDashboardResponse = z.infer<typeof getDashboardResponseSchema>;
