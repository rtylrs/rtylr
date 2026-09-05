import { describe, it, expect } from 'vitest';
import {
  formatDate,
  formatDateShort,
  daysBetween,
  addDays,
  monthKey,
  monthLabel,
  fiscalYearStartFromDate,
  quarterOf,
} from './date';

describe('formatDate', () => {
  it('formats an ISO date string to a readable date', () => {
    // "2026-04-15" → "Apr 15, 2026" in en-US
    const result = formatDate('2026-04-15');
    expect(result).toMatch(/Apr/);
    expect(result).toMatch(/2026/);
    expect(result).toMatch(/15/);
  });

  it('formats a Date object', () => {
    const d = new Date(2026, 0, 1);
    const result = formatDate(d);
    expect(result).toMatch(/2026/);
  });

  it('returns "-" for an invalid date string', () => {
    expect(formatDate('not-a-date')).toBe('-');
  });

  it('accepts custom Intl options', () => {
    const result = formatDate('2026-06-05', { month: 'long', year: 'numeric' });
    expect(result).toMatch(/June/);
    expect(result).toMatch(/2026/);
  });
});

describe('formatDateShort', () => {
  it('omits the year', () => {
    const result = formatDateShort('2026-04-15');
    expect(result).not.toMatch(/2026/);
    expect(result).toMatch(/Apr/);
    expect(result).toMatch(/15/);
  });
});

describe('daysBetween', () => {
  it('returns 0 for the same date', () => {
    expect(daysBetween('2026-04-01', '2026-04-01')).toBe(0);
  });

  it('returns positive when b is after a', () => {
    expect(daysBetween('2026-04-01', '2026-04-11')).toBe(10);
  });

  it('returns negative when b is before a', () => {
    expect(daysBetween('2026-04-11', '2026-04-01')).toBe(-10);
  });

  it('works with Date objects', () => {
    const a = new Date(2026, 0, 1);
    const b = new Date(2026, 0, 31);
    expect(daysBetween(a, b)).toBe(30);
  });
});

describe('addDays', () => {
  it('adds positive days', () => {
    const result = addDays('2026-04-01', 10);
    expect(result.getDate()).toBe(11);
    expect(result.getMonth()).toBe(3); // April = 3 (0-indexed)
  });

  it('adds zero days (identity)', () => {
    const result = addDays('2026-04-01', 0);
    expect(result.getDate()).toBe(1);
    expect(result.getMonth()).toBe(3);
  });

  it('subtracts days with negative count', () => {
    const result = addDays('2026-04-10', -5);
    expect(result.getDate()).toBe(5);
  });

  it('rolls over month boundary', () => {
    const result = addDays('2026-01-30', 5);
    expect(result.getMonth()).toBe(1); // February
  });

  it('accepts a Date object', () => {
    const d = new Date(2026, 5, 1);
    const result = addDays(d, 1);
    expect(result.getDate()).toBe(2);
  });
});

describe('monthKey', () => {
  it('returns YYYY-MM for January', () => {
    expect(monthKey(new Date(2026, 0, 15))).toBe('2026-01');
  });

  it('pads single-digit months with a leading zero', () => {
    expect(monthKey(new Date(2026, 8, 1))).toBe('2026-09');
  });

  it('handles December', () => {
    expect(monthKey(new Date(2026, 11, 31))).toBe('2026-12');
  });
});

describe('monthLabel', () => {
  it('returns short month and year', () => {
    const result = monthLabel(new Date(2026, 3, 1));
    expect(result).toMatch(/Apr/);
    expect(result).toMatch(/2026/);
  });

  it('accepts a string input', () => {
    const result = monthLabel('2026-07-01');
    expect(result).toMatch(/Jul/);
  });
});

describe('fiscalYearStartFromDate', () => {
  it('returns the start of the fiscal year when the date is after the fiscal start month', () => {
    // Default fiscal start = January. Date in April → FY started in January of same year.
    const result = fiscalYearStartFromDate(new Date(2026, 3, 15));
    expect(result.getFullYear()).toBe(2026);
    expect(result.getMonth()).toBe(0); // January
    expect(result.getDate()).toBe(1);
  });

  it('returns previous year when date is before fiscal start month', () => {
    // Fiscal year starts in July. Date in March → FY started in July of the prior year.
    const result = fiscalYearStartFromDate(new Date(2026, 2, 1), 7);
    expect(result.getFullYear()).toBe(2025);
    expect(result.getMonth()).toBe(6); // July = 6
  });

  it('handles fiscal start month = 1 (January)', () => {
    const result = fiscalYearStartFromDate(new Date(2026, 0, 1), 1);
    expect(result.getFullYear()).toBe(2026);
    expect(result.getMonth()).toBe(0);
  });
});

describe('quarterOf', () => {
  it('returns Q1 for January', () => {
    expect(quarterOf(new Date(2026, 0, 1))).toBe(1);
  });

  it('returns Q1 for March', () => {
    expect(quarterOf(new Date(2026, 2, 31))).toBe(1);
  });

  it('returns Q2 for April', () => {
    expect(quarterOf(new Date(2026, 3, 1))).toBe(2);
  });

  it('returns Q2 for June', () => {
    expect(quarterOf(new Date(2026, 5, 30))).toBe(2);
  });

  it('returns Q3 for July', () => {
    expect(quarterOf(new Date(2026, 6, 1))).toBe(3);
  });

  it('returns Q4 for October', () => {
    expect(quarterOf(new Date(2026, 9, 15))).toBe(4);
  });

  it('returns Q4 for December', () => {
    expect(quarterOf(new Date(2026, 11, 31))).toBe(4);
  });
});
