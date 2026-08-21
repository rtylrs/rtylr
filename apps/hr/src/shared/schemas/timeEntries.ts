import { z } from 'zod';

export const timeEntryFormSchema = z.object({
  employeeId: z.number().int().positive(),
  date: z.string().min(1),
  durationMinutes: z.number().int().min(1),
  startTime: z.string().optional(),
  endTime: z.string().optional(),
  projectId: z.number().optional(),
  taskId: z.number().optional(),
  categoryId: z.number().optional(),
  description: z.string(),
  isBillable: z.boolean(),
  // 'flow_import' is gone with the Flow import button: nothing produces it now.
  // 'flow' is deliberately absent too, since only the backend writes that value
  // when Flow logs a work-day line; this schema validates form input.
  source: z.enum(['manual', 'timer']),
});

export type TimeEntryFormData = z.infer<typeof timeEntryFormSchema>;
