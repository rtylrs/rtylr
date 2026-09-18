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

/**
 * One window in which a staff member can, or explicitly cannot, be scheduled.
 *
 * Mirrors `hr.shift.v1.AvailabilityPayload` with three deliberate differences:
 *
 * - int64 ids arrive as strings on the wire and are narrowed to numbers here,
 *   the same conversion `useTimeOff` applies to its staff ids.
 * - `staffName` is joined in from the staff service. The payload carries only
 *   `staff_id`, so the name cannot come from this endpoint.
 * - `isAvailable` defaults to false instead of being optional. The generated Go
 *   server tags every scalar `omitempty`, so a false `is_available` is dropped
 *   from the response body entirely, and the contract states that an absent key
 *   reads as false. Defaulting here keeps that rule in one place rather than at
 *   each call site.
 *
 * A row is either a recurring weekly pattern (`dayOfWeek` set) or a single-date
 * override (`specificDate` set), never both. That constraint is enforced by the
 * onion handler and, on the write path, by the `SetAvailabilityInput` union.
 */
export const availabilitySchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  // 0 = Monday through 6 = Sunday, matching the contract so the value indexes
  // the same Monday-first week the roster uses. Deliberately NOT JavaScript's
  // Date.getDay(), where 0 is Sunday.
  dayOfWeek: z.number().int().min(0).max(6).optional(),
  specificDate: z.string().optional(),
  isAvailable: z.boolean().default(false),
  // An endTime earlier than startTime means the window crosses midnight and
  // ends the following day, which is how a night shift is expressed.
  startTime: z.string(),
  endTime: z.string(),
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
