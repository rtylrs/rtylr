import { z } from 'zod';

export const payrollConfigSchema = z.object({
  grossModel: z.enum(['hourly', 'salary', 'both']),
  period: z.enum(['weekly', 'biweekly', 'monthly']),
  overtimeEnabled: z.boolean(),
  deductionsEnabled: z.boolean(),
  bonusesEnabled: z.boolean(),
});

export type PayrollConfigForm = z.infer<typeof payrollConfigSchema>;

export const payrollRunEntrySchema = z.object({
  id: z.number(),
  employeeId: z.number(),
  employeeName: z.string(),
  hoursWorked: z.number(),
  basePay: z.number(),
  overtimePay: z.number(),
  deductions: z.number(),
  bonuses: z.number(),
  netPay: z.number(),
});

export type PayrollRunEntry = z.infer<typeof payrollRunEntrySchema>;

export const payrollRunSchema = z.object({
  id: z.number(),
  periodStart: z.string(),
  periodEnd: z.string(),
  status: z.enum(['Draft', 'Pending', 'Approved', 'Paid']),
  totalNetPay: z.number(),
  totalBasePay: z.number(),
  totalOvertimePay: z.number(),
  totalDeductions: z.number(),
  employeeCount: z.number(),
  createdAt: z.string(),
  entries: z.array(payrollRunEntrySchema).optional(),
});

export type PayrollRun = z.infer<typeof payrollRunSchema>;
