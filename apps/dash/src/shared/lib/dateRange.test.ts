import { describe, it, expect } from 'vitest';
import { addDays, isWithinRange, daysBetween, presetToRange } from './dateRange';

describe('addDays', () => {
  it('adds positive days', () => {
    const base = new Date('2025-01-01');
    expect(addDays(base, 7).toISOString().slice(0, 10)).toBe('2025-01-08');
  });

  it('subtracts days for negative input', () => {
    const base = new Date('2025-01-10');
    expect(addDays(base, -3).toISOString().slice(0, 10)).toBe('2025-01-07');
  });

  it('does not mutate the input date', () => {
    const base = new Date('2025-01-01');
    addDays(base, 5);
    expect(base.toISOString().slice(0, 10)).toBe('2025-01-01');
  });
});

describe('isWithinRange', () => {
  const range = { start: new Date('2025-01-01'), end: new Date('2025-01-31') };

  it('returns true for a date inside the range', () => {
    expect(isWithinRange(new Date('2025-01-15'), range)).toBe(true);
  });

  it('returns true for a date at the start boundary', () => {
    expect(isWithinRange(new Date('2025-01-01'), range)).toBe(true);
  });

  it('returns false for a date before the range', () => {
    expect(isWithinRange(new Date('2024-12-31'), range)).toBe(false);
  });

  it('returns false for a date after the range', () => {
    expect(isWithinRange(new Date('2025-02-01'), range)).toBe(false);
  });
});

describe('daysBetween', () => {
  it('returns the number of days between two dates', () => {
    expect(daysBetween(new Date('2025-01-01'), new Date('2025-01-08'))).toBe(7);
  });

  it('returns 0 for the same date', () => {
    expect(daysBetween(new Date('2025-06-01'), new Date('2025-06-01'))).toBe(0);
  });

  it('is commutative', () => {
    const a = new Date('2025-01-01');
    const b = new Date('2025-03-01');
    expect(daysBetween(a, b)).toBe(daysBetween(b, a));
  });
});

describe('presetToRange', () => {
  const now = new Date('2025-06-15');

  it('returns a 7-day range ending today', () => {
    const range = presetToRange('7d', now);
    expect(daysBetween(range.start, range.end)).toBe(7);
  });

  it('returns a 30-day range', () => {
    const range = presetToRange('30d', now);
    expect(daysBetween(range.start, range.end)).toBe(30);
  });

  it('returns MTD starting on the first of month', () => {
    const range = presetToRange('mtd', now);
    expect(range.start.getDate()).toBe(1);
    expect(range.start.getMonth()).toBe(now.getMonth());
  });

  it('returns YTD starting on Jan 1', () => {
    const range = presetToRange('ytd', now);
    expect(range.start.getMonth()).toBe(0);
    expect(range.start.getDate()).toBe(1);
  });
});
