import { describe, it, expect } from 'vitest';
import {
  formatLongDate,
  formatShortDate,
  toIsoDate,
  daysBetween,
  isBeforeToday,
  relativeLabel,
} from '../date';

describe('formatLongDate', () => {
  it('formats a date object', () => {
    expect(formatLongDate(new Date(2024, 0, 15))).toBe('January 15, 2024');
  });

  it('accepts an ISO string', () => {
    const result = formatLongDate('2024-03-01');
    expect(result).toContain('2024');
  });
});

describe('formatShortDate', () => {
  it('formats a date with short month', () => {
    expect(formatShortDate(new Date(2024, 0, 15))).toBe('Jan 15, 2024');
  });
});

describe('toIsoDate', () => {
  it('returns YYYY-MM-DD format', () => {
    const result = toIsoDate(new Date('2024-06-20T12:00:00Z'));
    expect(result).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});

describe('daysBetween', () => {
  it('returns 0 for same day', () => {
    const d = new Date('2024-01-01');
    expect(daysBetween(d, d)).toBe(0);
  });

  it('returns correct days', () => {
    expect(daysBetween('2024-01-01', '2024-01-08')).toBe(7);
  });

  it('is order-independent', () => {
    expect(daysBetween('2024-01-08', '2024-01-01')).toBe(7);
  });
});

describe('isBeforeToday', () => {
  it('returns true for a past date', () => {
    expect(isBeforeToday('2000-01-01')).toBe(true);
  });

  it('returns false for a future date', () => {
    expect(isBeforeToday('2099-01-01')).toBe(false);
  });
});

describe('relativeLabel', () => {
  it('returns Today for now', () => {
    expect(relativeLabel(new Date())).toBe('Today');
  });

  it('returns Yesterday for one day ago', () => {
    const d = new Date();
    d.setDate(d.getDate() - 1);
    expect(relativeLabel(d)).toBe('Yesterday');
  });

  it('returns days ago for recent dates', () => {
    const d = new Date();
    d.setDate(d.getDate() - 3);
    expect(relativeLabel(d)).toBe('3 days ago');
  });

  it('falls back to short date for old dates', () => {
    const result = relativeLabel('2000-01-01');
    expect(result).toContain('2000');
  });
});
