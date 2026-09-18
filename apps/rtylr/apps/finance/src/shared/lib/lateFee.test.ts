import { describe, it, expect } from 'vitest';
import { computeLateFee } from './lateFee';
import type { LateFeeInput } from './lateFee';

// Helper: build a "today" date that is `daysAfterDue` days after the
// provided dueDate string, so tests are deterministic regardless of wall
// clock time.
function todayOffset(dueDate: string, daysAfterDue: number): Date {
  // Built from local parts. `new Date('2026-04-01')` is UTC midnight, so
  // reading it back with a local getter lands on 31 March west of Greenwich
  // and every offset here came out a day short.
  const [year, month, day] = dueDate.split('-').map(Number);
  return new Date(year, month - 1, day + daysAfterDue);
}

describe('computeLateFee - not yet late', () => {
  it('returns isLate=false when today equals dueDate', () => {
    const result = computeLateFee({
      amount: 1000,
      dueDate: '2026-04-01',
      today: new Date('2026-04-01'),
      feePercent: 1.5,
      graceDays: 0,
    });
    expect(result.isLate).toBe(false);
    expect(result.fee).toBe(0);
    expect(result.totalOwed).toBe(1000);
    expect(result.daysLate).toBe(0);
  });

  it('returns isLate=false when still within grace period', () => {
    const today = todayOffset('2026-04-01', 5); // 5 days after due
    const result = computeLateFee({
      amount: 500,
      dueDate: '2026-04-01',
      today,
      feePercent: 2,
      graceDays: 7, // 5 < 7 → not late
    });
    expect(result.isLate).toBe(false);
    expect(result.fee).toBe(0);
    expect(result.daysLate).toBe(0);
  });

  it('returns isLate=false when today is before dueDate', () => {
    const result = computeLateFee({
      amount: 2000,
      dueDate: '2026-05-01',
      today: new Date('2026-04-15'),
      feePercent: 1,
      graceDays: 0,
    });
    expect(result.isLate).toBe(false);
    expect(result.totalOwed).toBe(2000);
  });
});

describe('computeLateFee - simple (non-compounding)', () => {
  it('applies a flat percentage fee after grace period', () => {
    const today = todayOffset('2026-04-01', 10); // 10 days late, grace = 5 → 5 days late
    const result = computeLateFee({
      amount: 1000,
      dueDate: '2026-04-01',
      today,
      feePercent: 5,
      graceDays: 5,
      compoundingMonthly: false,
    });
    expect(result.isLate).toBe(true);
    expect(result.daysLate).toBe(5);
    expect(result.fee).toBeCloseTo(50); // 5% of 1000
    expect(result.totalOwed).toBeCloseTo(1050);
  });

  it('charges a flat fee regardless of how many days late (non-compounding)', () => {
    const today = todayOffset('2026-04-01', 60); // 60 days late
    const result1 = computeLateFee({
      amount: 1000,
      dueDate: '2026-04-01',
      today,
      feePercent: 1.5,
      graceDays: 0,
      compoundingMonthly: false,
    });
    const today2 = todayOffset('2026-04-01', 30);
    const result2 = computeLateFee({
      amount: 1000,
      dueDate: '2026-04-01',
      today: today2,
      feePercent: 1.5,
      graceDays: 0,
      compoundingMonthly: false,
    });
    // Non-compounding: fee is always feePercent% of amount, not time-dependent.
    expect(result1.fee).toBeCloseTo(result2.fee);
  });
});

describe('computeLateFee - compounding monthly (default)', () => {
  it('compounds over one month (exactly 30 days late)', () => {
    const today = todayOffset('2026-04-01', 30);
    const result = computeLateFee({
      amount: 1000,
      dueDate: '2026-04-01',
      today,
      feePercent: 1.5,
      graceDays: 0,
      compoundingMonthly: true,
    });
    // 1 month → fee = 1000 * ((1.015)^1 - 1) = 15
    expect(result.isLate).toBe(true);
    expect(result.fee).toBeCloseTo(15);
    expect(result.totalOwed).toBeCloseTo(1015);
  });

  it('compounds over two months (61 days late → ceil(61/30)=3 months)', () => {
    const today = todayOffset('2026-04-01', 61);
    const result = computeLateFee({
      amount: 1000,
      dueDate: '2026-04-01',
      today,
      feePercent: 1.5,
      graceDays: 0,
      compoundingMonthly: true,
    });
    // ceil(61/30) = 3 months
    const expected = 1000 * Math.pow(1.015, 3) - 1000;
    expect(result.fee).toBeCloseTo(expected);
  });

  it('produces a larger fee than non-compounding over many months', () => {
    const today = todayOffset('2026-04-01', 120); // ~4 months
    const base: LateFeeInput = {
      amount: 10_000,
      dueDate: '2026-04-01',
      today,
      feePercent: 2,
      graceDays: 0,
    };
    const compounding = computeLateFee({ ...base, compoundingMonthly: true });
    const flat = computeLateFee({ ...base, compoundingMonthly: false });
    expect(compounding.fee).toBeGreaterThan(flat.fee);
  });

  it('defaults to compounding when compoundingMonthly is not provided', () => {
    const today = todayOffset('2026-04-01', 30);
    const result = computeLateFee({
      amount: 500,
      dueDate: '2026-04-01',
      today,
      feePercent: 2,
      graceDays: 0,
    });
    const expected = 500 * Math.pow(1.02, 1) - 500;
    expect(result.fee).toBeCloseTo(expected);
  });
});

describe('computeLateFee - edge cases', () => {
  it('handles zero amount', () => {
    const today = todayOffset('2026-04-01', 30);
    const result = computeLateFee({
      amount: 0,
      dueDate: '2026-04-01',
      today,
      feePercent: 5,
      graceDays: 0,
    });
    expect(result.fee).toBe(0);
    expect(result.totalOwed).toBe(0);
  });

  it('handles zero feePercent', () => {
    const today = todayOffset('2026-04-01', 30);
    const result = computeLateFee({
      amount: 1000,
      dueDate: '2026-04-01',
      today,
      feePercent: 0,
      graceDays: 0,
    });
    expect(result.fee).toBe(0);
    expect(result.totalOwed).toBe(1000);
  });

  it('accepts a Date object for dueDate', () => {
    const dueDate = new Date('2026-04-01');
    const today = new Date('2026-05-01'); // 30 days later
    const result = computeLateFee({
      amount: 1000,
      dueDate,
      today,
      feePercent: 1.5,
      graceDays: 0,
    });
    expect(result.isLate).toBe(true);
    expect(result.fee).toBeCloseTo(15);
  });

  it('uses wall-clock today when today is not provided', () => {
    // Due date far in the past → must be late regardless of wall clock.
    const result = computeLateFee({
      amount: 1000,
      dueDate: '2000-01-01',
      feePercent: 1,
      graceDays: 0,
    });
    expect(result.isLate).toBe(true);
    expect(result.fee).toBeGreaterThan(0);
  });
});
