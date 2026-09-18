import { describe, expect, it } from 'vitest';

import {
  addDays,
  daysBetween,
  formatDisplayDate,
  formatShortDate,
  isWeekend,
  relativeTime,
  today,
  yearRange,
} from './dateUtils';

describe('formatDisplayDate', () => {
  it('formats a date as long month name', () => {
    const result = formatDisplayDate('2024-03-15');
    expect(result).toContain('March');
    expect(result).toContain('2024');
  });

  it('returns empty string for an invalid date', () => {
    expect(formatDisplayDate('not-a-date')).toBe('');
  });
});

describe('formatShortDate', () => {
  it('formats a date with abbreviated month', () => {
    const result = formatShortDate('2024-03-15');
    expect(result).toContain('Mar');
    expect(result).toContain('2024');
  });

  it('returns empty string for an invalid date', () => {
    expect(formatShortDate('invalid')).toBe('');
  });
});

describe('relativeTime', () => {
  it('returns just now for very recent dates', () => {
    const recent = new Date(Date.now() - 5000).toISOString();
    expect(relativeTime(recent)).toBe('just now');
  });

  it('returns days ago for dates a few days back', () => {
    const pastDate = new Date(Date.now() - 2 * 86400 * 1000).toISOString();
    const result = relativeTime(pastDate);
    expect(result).toContain('d ago');
  });

  it('returns empty string for invalid dates', () => {
    expect(relativeTime('bad-date')).toBe('');
  });

  it('returns formatted date for dates over 30 days', () => {
    const old = '2020-01-01';
    const result = relativeTime(old);
    expect(result).toContain('2020');
  });
});

describe('isWeekend', () => {
  it('returns true for a Sunday', () => {
    expect(isWeekend('2024-03-17')).toBe(true);
  });

  it('returns true for a Saturday', () => {
    expect(isWeekend('2024-03-16')).toBe(true);
  });

  it('returns false for a weekday', () => {
    expect(isWeekend('2024-03-15')).toBe(false);
  });
});

describe('today', () => {
  it('returns a string matching YYYY-MM-DD format', () => {
    expect(today()).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});

describe('addDays', () => {
  it('adds positive days', () => {
    expect(addDays('2024-01-15', 10)).toBe('2024-01-25');
  });

  it('adds negative days', () => {
    expect(addDays('2024-01-25', -5)).toBe('2024-01-20');
  });

  it('crosses month boundary', () => {
    expect(addDays('2024-01-31', 1)).toBe('2024-02-01');
  });
});

describe('daysBetween', () => {
  it('calculates days between two dates', () => {
    expect(daysBetween('2024-01-01', '2024-01-11')).toBe(10);
  });

  it('returns 0 for the same date', () => {
    expect(daysBetween('2024-03-15', '2024-03-15')).toBe(0);
  });

  it('returns negative when end is before start', () => {
    expect(daysBetween('2024-01-15', '2024-01-01')).toBe(-14);
  });
});

describe('yearRange', () => {
  it('returns a single year when no end year', () => {
    expect(yearRange(2024)).toBe('2024');
  });

  it('returns a single year when start equals end', () => {
    expect(yearRange(2024, 2024)).toBe('2024');
  });

  it('returns a range when years differ', () => {
    expect(yearRange(2024, 2026)).toBe('2024-2026');
  });
});
