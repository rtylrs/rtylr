/**
 * Pure payroll calculation helpers used throughout the HR payroll module.
 * All functions are side-effect free and operate only on plain numbers.
 */

/** Standard full-time hours per week in most jurisdictions. */
export const STANDARD_HOURS_PER_WEEK = 40;

/** Default overtime multiplier (1.5x). */
export const DEFAULT_OT_MULTIPLIER = 1.5;

/**
 * Calculates the number of overtime hours given total weekly hours.
 * Returns 0 when hoursWorked is at or below the threshold.
 */
export function calculateOvertimeHours(
  hoursWorked: number,
  threshold = STANDARD_HOURS_PER_WEEK
): number {
  return Math.max(0, hoursWorked - threshold);
}

/**
 * Calculates gross pay for an hourly employee.
 * Regular hours are capped at threshold; anything above is overtime.
 */
export function calculateHourlyGross(
  hoursWorked: number,
  hourlyRate: number,
  threshold = STANDARD_HOURS_PER_WEEK,
  otMultiplier = DEFAULT_OT_MULTIPLIER
): number {
  const regularHours = Math.min(hoursWorked, threshold);
  const otHours = calculateOvertimeHours(hoursWorked, threshold);
  const regularPay = regularHours * hourlyRate;
  const otPay = otHours * hourlyRate * otMultiplier;
  return regularPay + otPay;
}

/**
 * Calculates net pay after deductions and after adding bonuses.
 * Returns 0 if the result would be negative.
 */
export function calculateNetPay(
  basePay: number,
  overtimePay: number,
  bonuses: number,
  deductions: number
): number {
  return Math.max(0, basePay + overtimePay + bonuses - deductions);
}

/**
 * Calculates a simple flat-rate tax deduction amount.
 * taxRate should be a decimal fraction e.g. 0.2 for 20%.
 */
export function calculateTaxDeduction(grossPay: number, taxRate: number): number {
  return grossPay * Math.max(0, Math.min(1, taxRate));
}

/**
 * Rounds a currency value to 2 decimal places (cents).
 */
export function roundCurrency(amount: number): number {
  return Math.round(amount * 100) / 100;
}

/**
 * Sums an array of PayrollRunEntry-like objects to produce roll-up totals.
 */
export function sumPayrollEntries(
  entries: Array<{
    basePay: number;
    overtimePay: number;
    bonuses: number;
    deductions: number;
    netPay: number;
  }>
): {
  totalBasePay: number;
  totalOvertimePay: number;
  totalBonuses: number;
  totalDeductions: number;
  totalNetPay: number;
} {
  return entries.reduce(
    (acc, e) => ({
      totalBasePay: acc.totalBasePay + e.basePay,
      totalOvertimePay: acc.totalOvertimePay + e.overtimePay,
      totalBonuses: acc.totalBonuses + e.bonuses,
      totalDeductions: acc.totalDeductions + e.deductions,
      totalNetPay: acc.totalNetPay + e.netPay,
    }),
    {
      totalBasePay: 0,
      totalOvertimePay: 0,
      totalBonuses: 0,
      totalDeductions: 0,
      totalNetPay: 0,
    }
  );
}

/**
 * Converts an annual salary to a per-period amount.
 * period: 'monthly' = /12, 'biweekly' = /26, 'weekly' = /52.
 */
export function annualToPerPeriod(
  annualSalary: number,
  period: 'weekly' | 'biweekly' | 'monthly'
): number {
  const divisors = { weekly: 52, biweekly: 26, monthly: 12 };
  return annualSalary / divisors[period];
}

/**
 * Calculates employer social security contribution up to a wage cap.
 */
export function employerSocialSecurity(grossPay: number, wageCap: number, rate: number): number {
  if (rate < 0 || rate > 1) return 0;
  return Math.round(Math.min(grossPay, wageCap) * rate * 100) / 100;
}

/**
 * Calculates total employer cost including an overhead fraction.
 */
export function totalEmployerCost(grossPay: number, overhead: number): number {
  return Math.round(grossPay * (1 + Math.max(0, overhead)) * 100) / 100;
}

/**
 * Returns true when payroll period start is before end.
 */
export function isValidPayrollPeriod(startIso: string, endIso: string): boolean {
  return new Date(startIso) < new Date(endIso);
}
