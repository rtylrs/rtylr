import { z } from 'zod';

export const departmentSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  description: z.string(),
  staffCount: z.number(),
  head: z.string(),
  avgSalary: z.number(),
  color: z.string().optional(),
});

export type Department = z.infer<typeof departmentSchema>;
