import { z } from 'zod';

export const operationsSummarySchema = z.object({
  headcount: z.number().optional(),
  totalStaff: z.number().optional(),
  activeStaff: z.number().optional(),
  avgHoursWorked: z.number().optional(),
  avgUtilization: z.number().optional(),
  shiftCoverage: z.number().optional(),
  pendingTimeOff: z.number().optional(),
  openRoles: z.number().optional(),
});

export const staffUtilizationSchema = z.object({
  staffId: z.union([z.string(), z.number()]).transform(String).optional(),
  name: z.string().optional(),
  utilization: z.number().optional(),
  hoursScheduled: z.number().optional(),
  hoursWorked: z.number().optional(),
  average: z.number().optional(),
  data: z.array(z.unknown()).optional(),
  staff: z.array(z.unknown()).optional(),
});

export const shiftCoverageSchema = z.object({
  date: z.string().optional(),
  coverage: z.number().optional(),
  coveredPercent: z.number().optional(),
  scheduled: z.number().optional(),
  actual: z.number().optional(),
});

export const timeOffTrendSchema = z.object({
  period: z.string().optional(),
  requests: z.number().optional(),
  pending: z.number().optional(),
  approved: z.number().optional(),
  denied: z.number().optional(),
  data: z.array(z.unknown()).optional(),
  recent: z.array(z.unknown()).optional(),
});

export const deptHeadcountSchema = z.object({
  department: z.string().optional(),
  headcount: z.number().optional(),
  change: z.number().optional(),
  data: z.array(z.unknown()).optional(),
});

export const hrDashboardSchema = z.object({
  totalEmployees: z.number().optional(),
  newHires: z.number().optional(),
  attrition: z.number().optional(),
  openPositions: z.number().optional(),
});

export type OperationsSummary = z.infer<typeof operationsSummarySchema>;
export type StaffUtilization = z.infer<typeof staffUtilizationSchema>;
export type ShiftCoverage = z.infer<typeof shiftCoverageSchema>;
export type TimeOffTrend = z.infer<typeof timeOffTrendSchema>;
export type DeptHeadcount = z.infer<typeof deptHeadcountSchema>;
export type HrDashboard = z.infer<typeof hrDashboardSchema>;
