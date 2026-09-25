import { z } from 'zod';

export const timesheetRowSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  staffId: z.number(),
  staffName: z.string(),
  weekOf: z.string(),
  hours: z.object({
    mon: z.number(),
    tue: z.number(),
    wed: z.number(),
    thu: z.number(),
    fri: z.number(),
    sat: z.number(),
    sun: z.number(),
  }),
  approved: z.boolean().optional(),
});

export type TimesheetRow = z.infer<typeof timesheetRowSchema>;
