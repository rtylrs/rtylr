import { describe, it, expect } from 'vitest';
import {
  calculateOvertimeHours,
  calculateHourlyGross,
  calculateNetPay,
  calculateTaxDeduction,
  roundCurrency,
  sumPayrollEntries,
  annualToPerPeriod,
  STANDARD_HOURS_PER_WEEK,
  DEFAULT_OT_MULTIPLIER,
} from './payrollUtils';

describe('constants', () => {
  it('STANDARD_HOURS_PER_WEEK is 40', () => {
    expect(STANDARD_HOURS_PER_WEEK).toBe(40);
  });

  it('DEFAULT_OT_MULTIPLIER is 1.5', () => {
    expect(DEFAULT_OT_MULTIPLIER).toBe(1.5);
  });
});

describe('calculateOvertimeHours', () => {
  it('returns 0 for exactly 40 hours', () => {
    expect(calculateOvertimeHours(40)).toBe(0);
  });

  it('returns 0 for fewer than 40 hours', () => {
    expect(calculateOvertimeHours(35)).toBe(0);
  });

  it('returns excess hours above 40', () => {
    expect(calculateOvertimeHours(45)).toBe(5);
  });

  it('returns 0 for 0 hours', () => {
    expect(calculateOvertimeHours(0)).toBe(0);
  });

  it('uses a custom threshold', () => {
    expect(calculateOvertimeHours(50, 48)).toBe(2);
  });

  it('handles fractional hours', () => {
    expect(calculateOvertimeHours(40.5)).toBeCloseTo(0.5);
  });
});

describe('calculateHourlyGross', () => {
  it('calculates pay for exactly 40 regular hours', () => {
    expect(calculateHourlyGross(40, 20)).toBe(800);
  });

  it('calculates pay with overtime', () => {
    const result = calculateHourlyGross(45, 20);
    expect(result).toBe(800 + 5 * 20 * 1.5);
  });

  it('returns 0 for 0 hours', () => {
    expect(calculateHourlyGross(0, 20)).toBe(0);
  });

  it('uses custom threshold', () => {
    const result = calculateHourlyGross(50, 10, 48);
    expect(result).toBe(48 * 10 + 2 * 10 * 1.5);
  });

  it('uses custom overtime multiplier', () => {
    const result = calculateHourlyGross(42, 10, 40, 2);
    expect(result).toBe(400 + 2 * 10 * 2);
  });

  it('handles fractional hours', () => {
    const result = calculateHourlyGross(40.5, 20);
    expect(result).toBeCloseTo(800 + 0.5 * 20 * 1.5);
  });
});

describe('calculateNetPay', () => {
  it('returns basePay + overtime + bonuses - deductions', () => {
    expect(calculateNetPay(2000, 150, 100, 300)).toBe(1950);
  });

  it('returns 0 when deductions exceed gross', () => {
    expect(calculateNetPay(100, 0, 0, 500)).toBe(0);
  });

  it('handles all zeros', () => {
    expect(calculateNetPay(0, 0, 0, 0)).toBe(0);
  });

  it('handles no overtime or bonuses', () => {
    expect(calculateNetPay(3000, 0, 0, 500)).toBe(2500);
  });
});

describe('calculateTaxDeduction', () => {
  it('calculates 20% tax', () => {
    expect(calculateTaxDeduction(1000, 0.2)).toBeCloseTo(200);
  });

  it('returns 0 for 0 rate', () => {
    expect(calculateTaxDeduction(1000, 0)).toBe(0);
  });

  it('caps tax rate at 100%', () => {
    expect(calculateTaxDeduction(1000, 1)).toBe(1000);
  });

  it('clamps negative tax rate to 0', () => {
    expect(calculateTaxDeduction(1000, -0.1)).toBe(0);
  });

  it('clamps rate above 1 to 1', () => {
    expect(calculateTaxDeduction(1000, 1.5)).toBe(1000);
  });
});

describe('roundCurrency', () => {
  it('rounds to 2 decimal places', () => {
    expect(roundCurrency(1.456)).toBe(1.46);
  });

  it('does not change a value already at 2 dp', () => {
    expect(roundCurrency(10.25)).toBe(10.25);
  });

  it('handles whole numbers', () => {
    expect(roundCurrency(100)).toBe(100);
  });

  it('handles negative values', () => {
    expect(roundCurrency(-3.456)).toBe(-3.46);
  });
});

describe('sumPayrollEntries', () => {
  it('returns zeroes for an empty array', () => {
    const result = sumPayrollEntries([]);
    expect(result.totalNetPay).toBe(0);
    expect(result.totalBasePay).toBe(0);
  });

  it('sums a single entry correctly', () => {
    const result = sumPayrollEntries([
      { basePay: 2000, overtimePay: 150, bonuses: 100, deductions: 200, netPay: 2050 },
    ]);
    expect(result.totalBasePay).toBe(2000);
    expect(result.totalOvertimePay).toBe(150);
    expect(result.totalBonuses).toBe(100);
    expect(result.totalDeductions).toBe(200);
    expect(result.totalNetPay).toBe(2050);
  });

  it('sums multiple entries', () => {
    const entries = [
      { basePay: 1000, overtimePay: 0, bonuses: 50, deductions: 100, netPay: 950 },
      { basePay: 2000, overtimePay: 200, bonuses: 0, deductions: 300, netPay: 1900 },
    ];
    const result = sumPayrollEntries(entries);
    expect(result.totalBasePay).toBe(3000);
    expect(result.totalOvertimePay).toBe(200);
    expect(result.totalBonuses).toBe(50);
    expect(result.totalDeductions).toBe(400);
    expect(result.totalNetPay).toBe(2850);
  });
});

describe('annualToPerPeriod', () => {
  it('divides by 12 for monthly', () => {
    expect(annualToPerPeriod(60000, 'monthly')).toBeCloseTo(5000);
  });

  it('divides by 26 for biweekly', () => {
    expect(annualToPerPeriod(52000, 'biweekly')).toBe(2000);
  });

  it('divides by 52 for weekly', () => {
    expect(annualToPerPeriod(52000, 'weekly')).toBe(1000);
  });

  it('handles fractional results', () => {
    const result = annualToPerPeriod(50000, 'monthly');
    expect(result).toBeCloseTo(4166.67, 1);
  });
});
