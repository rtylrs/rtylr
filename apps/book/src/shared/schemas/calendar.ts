import { z } from 'zod';

export const calendarEventSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(String),
  title: z.string().optional(),
  appointmentId: z.union([z.string(), z.number()]).transform(String).optional().nullable(),
  staffId: z.union([z.string(), z.number()]).transform(String).optional().nullable(),
  resourceId: z.union([z.string(), z.number()]).transform(String).optional().nullable(),
  start: z.string(),
  end: z.string(),
  status: z.enum(['pending', 'confirmed', 'completed', 'cancelled', 'blocked']).catch('confirmed'),
  color: z.string().optional().nullable(),
  allDay: z.boolean().optional(),
  notes: z.string().optional().nullable(),
});

export type CalendarEvent = z.infer<typeof calendarEventSchema>;
