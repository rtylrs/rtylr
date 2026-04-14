import { z } from 'zod';

export const overtimeRequestSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  date: z.string(),
  plannedMinutes: z.number(),
  actualMinutes: z.number().optional(),
  reason: z.string(),
  status: z.enum(['Pending', 'Approved', 'Rejected']).catch('Pending'),
  approvedBy: z.number().optional(),
  approvedAt: z.string().optional(),
  createdAt: z.string(),
});

export const overtimeConfigSchema = z.object({
  enabled: z.boolean(),
  dailyThresholdMinutes: z.number(),
  weeklyThresholdMinutes: z.number(),
  requiresPreApproval: z.boolean(),
  multiplier: z.number(),
});

export type OvertimeRequest = z.infer<typeof overtimeRequestSchema>;
export type OvertimeConfig = z.infer<typeof overtimeConfigSchema>;
