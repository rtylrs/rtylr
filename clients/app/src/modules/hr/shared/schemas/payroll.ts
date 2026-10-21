import { z } from 'zod';

const finiteNumber = z.preprocess((value) => {
  if (value === null || value === undefined || value === '') return value;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : value;
}, z.number());

/**
 * Search params for the pay slip view route (`/hr/payroll/my-payslips/view`).
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

/**
 * Path params for the payroll run detail route (`/hr/payroll/$id`).
 *
 * Unlike search params, path params are never JSON-parsed by the router -
 * `params.parse` always receives raw strings - so a string-to-number
 * transform is required here. It is still not `z.coerce`: coercion widens
 * the schema's input to `unknown`, and `Number('')` is `0`, which
 * `.pipe(z.number().int().positive())` would have to reject anyway.
 *
 * Malformed segments resolve to `undefined` instead of throwing, mirroring
 * `paySlipSearchSchema`'s `.catch(undefined)`: `params.parse` runs before the
 * route renders, so a throw would replace the feature's own not-found state
 * with the router error boundary. Run ids are backend serial primary keys,
 * so `-3` or `12.5` is always malformed input; resolving them to `undefined`
 * keeps `useGetPayrollRun` disabled (`enabled: !!id`) and no request is
 * issued.
 *
 * The validation lives inside the transform (rather than
 * `.pipe(...).optional().catch(...)`) for two typing constraints: the
 * router's `ValidateParsedParams` requires every path segment to stay a
 * REQUIRED key of the parsed type (`id: number | undefined`, never
 * `id?: number`), and Zod's `.pipe`/hr`.catch` cannot produce that shape
 * without an unsafe widening cast.
 */
export const payrollRunParamsSchema = z.object({
  id: z.string().transform((value): number | undefined => {
    const parsed = Number(value);
    // Same rule as `paySlipSearchSchema.runId`: a positive integer, or nothing.
    return Number.isInteger(parsed) && parsed > 0 ? parsed : undefined;
  }),
});

export type PayrollRunParams = z.infer<typeof payrollRunParamsSchema>;

export const payrollConfigSchema = z.object({
  period: z.enum(['weekly', 'biweekly', 'monthly']),
  overtimeEnabled: z.boolean(),
  payDayOfMonth: z.number().int().min(1).max(28),
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
