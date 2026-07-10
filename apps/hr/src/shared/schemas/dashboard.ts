import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : 0;
}, z.number());

export const hrDashboardStatsSchema = z.object({
  totalStaff: finiteNumber,
  onShiftToday: finiteNumber,
  openPositions: finiteNumber,
  monthlyPayroll: finiteNumber,
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
    shiftsScheduled: finiteNumber,
    newApplications: finiteNumber,
    pendingApprovals: finiteNumber,
  }),
});

export type HRDashboardStats = z.infer<typeof hrDashboardStatsSchema>;

export const dashboardWidgetSchema = z.object({
  id: finiteNumber,
  ownerId: finiteNumber,
  roleId: finiteNumber,
  widgetType: z.string(),
  positionX: finiteNumber,
  positionY: finiteNumber,
  width: finiteNumber,
  height: finiteNumber,
  config: z.unknown().default({}),
  createdAt: finiteNumber,
  updatedAt: finiteNumber,
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
