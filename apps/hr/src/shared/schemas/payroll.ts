import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  if (value === null || value === undefined || value === '') return value;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : value;
}, z.number());

export const payrollConfigSchema = z.object({
  grossModel: z.enum(['hourly', 'salary', 'both']),
  period: z.enum(['weekly', 'biweekly', 'monthly']),
  overtimeEnabled: z.boolean(),
  deductionsEnabled: z.boolean(),
  bonusesEnabled: z.boolean(),
});

export type PayrollConfigForm = z.infer<typeof payrollConfigSchema>;

export const payrollRunEntrySchema = z.object({
  id: finiteNumber,
  employeeId: finiteNumber,
  employeeName: z.string(),
  hoursWorked: finiteNumber,
  basePay: finiteNumber,
  overtimePay: finiteNumber,
  deductions: finiteNumber,
  bonuses: finiteNumber,
  netPay: finiteNumber,
});

export type PayrollRunEntry = z.infer<typeof payrollRunEntrySchema>;

export const payrollRunSchema = z.object({
  id: finiteNumber,
  periodStart: z.string(),
  periodEnd: z.string(),
  status: z.enum(['Draft', 'Pending', 'Approved', 'Paid']),
  totalNetPay: finiteNumber,
  totalBasePay: finiteNumber,
  totalOvertimePay: finiteNumber,
  totalDeductions: finiteNumber,
  employeeCount: finiteNumber,
  createdAt: z.string(),
  entries: z.array(payrollRunEntrySchema).optional(),
});

export type PayrollRun = z.infer<typeof payrollRunSchema>;
