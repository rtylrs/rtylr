import { z } from 'zod';

export const shiftSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  role: z.string(),
  start: z.string(),
  end: z.string(),
  status: z.enum(['Scheduled', 'In Progress', 'Completed']).catch('Scheduled'),
  day: z.string(),
  staffId: z.number().optional(),
});

export const scheduleShiftSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  role: z.string(),
  day: z.string(),
  start: z.string(),
  end: z.string(),
  color: z.string().optional(),
  published: z.boolean().optional(),
});

export const availabilitySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  day: z.string(),
  canWork: z.boolean(),
  startTime: z.string().optional(),
  endTime: z.string().optional(),
});

export const swapRequestSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  requesterId: z.number(),
  requester: z.string(),
  targetId: z.number(),
  target: z.string(),
  shiftDay: z.string(),
  reason: z.string(),
  status: z.enum(['Pending', 'Accepted', 'Rejected']).catch('Pending'),
});

export const breakLogSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  start: z.string(),
  end: z.string().optional(),
  durationMin: z.number().optional(),
  type: z.enum(['Lunch', 'Short', 'Other']).catch('Other'),
});

export const overtimeAlertSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  hoursThisWeek: z.number(),
  threshold: z.number(),
  triggeredAt: z.string(),
});

export type Shift = z.infer<typeof shiftSchema>;
export type ScheduleShift = z.infer<typeof scheduleShiftSchema>;
export type Availability = z.infer<typeof availabilitySchema>;
export type SwapRequest = z.infer<typeof swapRequestSchema>;
export type BreakLog = z.infer<typeof breakLogSchema>;
export type OvertimeAlert = z.infer<typeof overtimeAlertSchema>;
