import { z } from 'zod';

export const payrollRowSchema = z.object({
  id: z.union([z.string(), z.number()]).transform(Number),
  name: z.string(),
  role: z.string(),
  hoursWorked: z.number(),
  basePay: z.number(),
  overtime: z.number(),
  deductions: z.number(),
  status: z.enum(['Pending', 'Paid']).optional(),
});

export type PayrollRow = z.infer<typeof payrollRowSchema>;
