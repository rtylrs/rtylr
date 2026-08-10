import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  if (value === null || value === undefined || value === '') return value;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : value;
}, z.number());

/**
 * Search params for the pay slip view route (`/payroll/my-payslips/view`).
 *
 * No `z.coerce` here, for two reasons. The router's `defaultParseSearch` already
 * JSON-parses query values, so `?runId=12` arrives as a number. And `z.coerce`
 * widens the schema's input type to `unknown`, which would silently disable
 * type checking on the `search` prop at every `<Link>` call site.
 *
 * `.catch(undefined)` keeps the route tolerant: a missing, empty or malformed
 * `runId` resolves to `undefined` rather than throwing. `validateSearch` runs
 * before the route renders, so a throw would swap the "pay slip not found"
 * empty state for the router error boundary on any stale or hand-edited URL.
 * `useGetMyPayrollRun` treats `undefined` as "do not fetch", which is what
 * produces that empty state.
 */
export const paySlipSearchSchema = z.object({
  runId: z.number().int().positive().optional().catch(undefined),
});

export type PaySlipSearch = z.infer<typeof paySlipSearchSchema>;

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
