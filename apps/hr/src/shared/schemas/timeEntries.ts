import { z } from 'zod';

export const timeEntryFormSchema = z.object({
  date: z.string().min(1),
  durationMinutes: z.number().int().min(1),
  startTime: z.string().optional(),
  endTime: z.string().optional(),
  projectId: z.number().optional(),
  taskId: z.number().optional(),
  categoryId: z.number().optional(),
  description: z.string(),
  isBillable: z.boolean(),
  source: z.enum(['manual', 'timer', 'flow_import']),
});

export type TimeEntryFormData = z.infer<typeof timeEntryFormSchema>;
