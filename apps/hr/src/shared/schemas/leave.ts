import { z } from 'zod';

export const leaveTypeSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  color: z.string().optional(),
  defaultDays: z.number(),
  requiresApproval: z.boolean(),
});

export const leaveRequestSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  leaveTypeId: z.number(),
  leaveTypeName: z.string(),
  startDate: z.string(),
  endDate: z.string(),
  daysCount: z.number(),
  status: z.enum(['Pending', 'Approved', 'Rejected']).catch('Pending'),
  reason: z.string(),
  notes: z.string().optional(),
  approvedBy: z.number().optional(),
  approvedAt: z.string().optional(),
  createdAt: z.string(),
});

export const leaveBalanceSchema = z.object({
  leaveTypeId: z.number(),
  leaveTypeName: z.string(),
  entitled: z.number(),
  used: z.number(),
  pending: z.number(),
  remaining: z.number(),
});

export const leaveCalendarEntrySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  leaveTypeName: z.string(),
  color: z.string().optional(),
  startDate: z.string(),
  endDate: z.string(),
});

export type LeaveType = z.infer<typeof leaveTypeSchema>;
export type LeaveRequest = z.infer<typeof leaveRequestSchema>;
export type LeaveBalance = z.infer<typeof leaveBalanceSchema>;
export type LeaveCalendarEntry = z.infer<typeof leaveCalendarEntrySchema>;
