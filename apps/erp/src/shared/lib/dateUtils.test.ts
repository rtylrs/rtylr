import { describe, it, expect } from 'vitest';
import {
  startOfDay,
  endOfDay,
  addDays,
  isWithinRange,
  isPast,
  isFuture,
  toIsoDate,
  daysBetween,
  parseIsoDate,
} from './dateUtils';

const baseDate = new Date('2024-06-15T12:00:00Z');

describe('startOfDay', () => {
  it('returns midnight for a given date', () => {
    const d = startOfDay(new Date('2024-06-15T15:30:00'));
    expect(d.getHours()).toBe(0);
    expect(d.getMinutes()).toBe(0);
    expect(d.getSeconds()).toBe(0);
  });

  it('does not mutate the original date', () => {
    const original = new Date('2024-06-15T15:30:00');
    startOfDay(original);
    expect(original.getHours()).toBe(15);
  });
});

describe('endOfDay', () => {
  it('returns 23:59:59 for a given date', () => {
    const d = endOfDay(new Date('2024-06-15T12:00:00'));
    expect(d.getHours()).toBe(23);
    expect(d.getMinutes()).toBe(59);
    expect(d.getSeconds()).toBe(59);
  });
});

describe('addDays', () => {
  it('adds positive days', () => {
    const result = addDays(new Date('2024-01-01'), 10);
    expect(result.getDate()).toBe(11);
  });

  it('subtracts negative days', () => {
    const result = addDays(new Date('2024-01-10'), -5);
    expect(result.getDate()).toBe(5);
  });

  it('does not mutate the original date', () => {
    const original = new Date('2024-01-01');
    addDays(original, 10);
    expect(original.getDate()).toBe(1);
  });
});

describe('isWithinRange', () => {
  const start = new Date('2024-01-01');
  const end = new Date('2024-12-31');

  it('returns true for date within range', () => {
    expect(isWithinRange(new Date('2024-06-15'), start, end)).toBe(true);
  });

  it('returns true for date at boundaries', () => {
    expect(isWithinRange(start, start, end)).toBe(true);
    expect(isWithinRange(end, start, end)).toBe(true);
  });

  it('returns false for date outside range', () => {
    expect(isWithinRange(new Date('2025-01-01'), start, end)).toBe(false);
  });
});

describe('isPast', () => {
  it('returns true for a date in the past', () => {
    expect(isPast(new Date('2020-01-01'))).toBe(true);
  });

  it('returns false for a future date', () => {
    expect(isPast(new Date('2099-01-01'))).toBe(false);
  });
});

describe('isFuture', () => {
  it('returns true for a future date', () => {
    expect(isFuture(new Date('2099-01-01'))).toBe(true);
  });

  it('returns false for a past date', () => {
    expect(isFuture(new Date('2020-01-01'))).toBe(false);
  });
});

describe('toIsoDate', () => {
  it('returns YYYY-MM-DD format', () => {
    expect(toIsoDate(new Date('2024-06-15T12:00:00Z'))).toBe('2024-06-15');
  });
});

describe('daysBetween', () => {
  it('returns positive days when end is after start', () => {
    const start = new Date('2024-01-01');
    const end = new Date('2024-01-11');
    expect(daysBetween(start, end)).toBe(10);
  });

  it('returns negative days when end is before start', () => {
    const start = new Date('2024-01-11');
    const end = new Date('2024-01-01');
    expect(daysBetween(start, end)).toBe(-10);
  });

  it('returns 0 for same date', () => {
    const d = new Date('2024-01-01');
    expect(daysBetween(d, d)).toBe(0);
  });
});

describe('parseIsoDate', () => {
  it('parses ISO date string correctly', () => {
    const d = parseIsoDate('2024-06-15');
    expect(d.getFullYear()).toBe(2024);
    expect(d.getMonth()).toBe(5);
    expect(d.getDate()).toBe(15);
  });
});
