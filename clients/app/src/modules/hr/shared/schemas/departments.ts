import { z } from 'zod';

export const createDepartmentSchema = z.object({
  name: z.string().min(1),
  description: z.string().optional(),
});

export type CreateDepartmentForm = z.infer<typeof createDepartmentSchema>;

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
