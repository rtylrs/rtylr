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
  source: z.enum(['manual', 'timer', 'flow_import']),
});

export type TimeEntryFormData = z.infer<typeof timeEntryFormSchema>;

export const flowImportFormSchema = z
  .object({
    employeeId: z.number().int().positive(),
    periodStart: z.string().min(1),
    periodEnd: z.string().min(1),
  })
  .refine((data) => data.periodStart <= data.periodEnd, {
    message: 'periodEnd must be on or after periodStart',
    path: ['periodEnd'],
  });

export type FlowImportFormData = z.infer<typeof flowImportFormSchema>;
