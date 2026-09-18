import { describe, it, expect } from 'vitest';
import {
  toISODate,
  addDays,
  diffDays,
  formatAxisDate,
  formatFullDate,
  isSameMonth,
  startOfMonth,
} from './date';

describe('toISODate', () => {
  it('formats a Date as YYYY-MM-DD using UTC', () => {
    const d = new Date('2026-05-15T00:00:00Z');
    expect(toISODate(d)).toBe('2026-05-15');
  });

  it('pads single-digit month and day', () => {
    const d = new Date('2026-01-05T00:00:00Z');
    expect(toISODate(d)).toBe('2026-01-05');
  });
});

describe('addDays', () => {
  it('adds a positive number of days', () => {
    const start = new Date('2026-01-01T00:00:00Z');
    const result = addDays(start, 10);
    expect(toISODate(result)).toBe('2026-01-11');
  });

  it('subtracts days for a negative n', () => {
    const start = new Date('2026-01-15T00:00:00Z');
    const result = addDays(start, -5);
    expect(toISODate(result)).toBe('2026-01-10');
  });

  it('does not mutate the original date', () => {
    const start = new Date('2026-01-01T00:00:00Z');
    addDays(start, 5);
    expect(toISODate(start)).toBe('2026-01-01');
  });
});

describe('diffDays', () => {
  it('returns the correct number of days between two dates', () => {
    expect(diffDays('2026-01-01', '2026-01-31')).toBe(30);
  });

  it('returns 0 for the same date', () => {
    expect(diffDays('2026-05-01', '2026-05-01')).toBe(0);
  });

  it('returns a positive value regardless of order', () => {
    expect(diffDays('2026-01-31', '2026-01-01')).toBe(30);
  });
});

describe('formatAxisDate', () => {
  it('formats a date as "Mon DD"', () => {
    expect(formatAxisDate('2026-05-15')).toBe('May 15');
  });

  it('handles January', () => {
    expect(formatAxisDate('2026-01-01')).toBe('Jan 1');
  });

  it('handles December', () => {
    expect(formatAxisDate('2026-12-31')).toBe('Dec 31');
  });
});

describe('formatFullDate', () => {
  it('formats a date as "Month DD, YYYY"', () => {
    expect(formatFullDate('2026-05-15')).toBe('May 15, 2026');
  });

  it('uses full month names', () => {
    expect(formatFullDate('2026-02-28')).toContain('February');
  });
});

describe('isSameMonth', () => {
  it('returns true for dates in the same month', () => {
    expect(isSameMonth('2026-05-01', '2026-05-31')).toBe(true);
  });

  it('returns false for dates in different months', () => {
    expect(isSameMonth('2026-05-31', '2026-06-01')).toBe(false);
  });

  it('returns false for dates in different years', () => {
    expect(isSameMonth('2025-05-15', '2026-05-15')).toBe(false);
  });
});

describe('startOfMonth', () => {
  it('returns the first day of the month', () => {
    expect(startOfMonth('2026-05-15')).toBe('2026-05-01');
  });

  it('returns the same date when already the first', () => {
    expect(startOfMonth('2026-03-01')).toBe('2026-03-01');
  });

  it('handles end of month dates', () => {
    expect(startOfMonth('2026-01-31')).toBe('2026-01-01');
  });
});
