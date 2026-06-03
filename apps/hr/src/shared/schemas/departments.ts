import { z } from 'zod';

export const departmentSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  description: z.string().optional().default(''),
  staffCount: z.number().optional().default(0),
  head: z.string().optional().default(''),
  avgSalary: z.number().optional().default(0),
  color: z.string().optional(),
});

export type Department = z.infer<typeof departmentSchema>;
