import { describe, it, expect } from 'vitest';
import {
  formatDate,
  daysBetween,
  addDays,
  startOfMonth,
  endOfMonth,
  isWeekend,
  isToday,
  isFutureDate,
  isPastDate,
  yearsOld,
} from './hrDateUtils';

describe('formatDate', () => {
  it('formats short date by default', () => {
    const result = formatDate('2024-06-15');
    expect(result).toContain('2024');
  });
  it('returns empty string for invalid date', () => {
    expect(formatDate('not-a-date')).toBe('');
  });
  it('formats numeric date', () => {
    const result = formatDate('2024-06-15', 'numeric');
    expect(result).toBeTruthy();
  });
});

describe('daysBetween', () => {
  it('calculates days between two dates', () => {
    expect(daysBetween('2024-01-01', '2024-01-11')).toBe(10);
  });
  it('returns 0 for same date', () => {
    expect(daysBetween('2024-06-01', '2024-06-01')).toBe(0);
  });
  it('returns negative for reversed range', () => {
    expect(daysBetween('2024-01-11', '2024-01-01')).toBe(-10);
  });
});

describe('addDays', () => {
  it('adds positive days', () => {
    expect(addDays('2024-01-01', 10)).toBe('2024-01-11');
  });
  it('adds across month boundary', () => {
    expect(addDays('2024-01-31', 1)).toBe('2024-02-01');
  });
  it('subtracts days with negative value', () => {
    expect(addDays('2024-01-11', -10)).toBe('2024-01-01');
  });
});

describe('startOfMonth', () => {
  it('returns first day of month', () => {
    expect(startOfMonth('2024-06-15')).toBe('2024-06-01');
  });
  it('works for first day already', () => {
    expect(startOfMonth('2024-03-01')).toBe('2024-03-01');
  });
});

describe('endOfMonth', () => {
  it('returns last day of June', () => {
    expect(endOfMonth('2024-06-15')).toBe('2024-06-30');
  });
  it('handles February', () => {
    const result = endOfMonth('2024-02-10');
    expect(['2024-02-28', '2024-02-29']).toContain(result);
  });
});

describe('isWeekend', () => {
  it('returns true for Sunday', () => {
    expect(isWeekend('2024-06-16')).toBe(true);
  });
  it('returns true for Saturday', () => {
    expect(isWeekend('2024-06-15')).toBe(true);
  });
  it('returns false for Monday', () => {
    expect(isWeekend('2024-06-17')).toBe(false);
  });
});

describe('isToday', () => {
  it('returns false for a past date', () => {
    expect(isToday('2000-01-01')).toBe(false);
  });
});

describe('isFutureDate', () => {
  it('returns true for far future date', () => {
    expect(isFutureDate('2099-12-31')).toBe(true);
  });
  it('returns false for past date', () => {
    expect(isFutureDate('2000-01-01')).toBe(false);
  });
});

describe('isPastDate', () => {
  it('returns true for past date', () => {
    expect(isPastDate('2000-01-01')).toBe(true);
  });
  it('returns false for future date', () => {
    expect(isPastDate('2099-12-31')).toBe(false);
  });
});

describe('yearsOld', () => {
  it('calculates age from birth date', () => {
    const age = yearsOld('1990-01-01');
    expect(age).toBeGreaterThan(30);
  });
  it('returns 0 for today', () => {
    // Local parts, for the same reason as the CRM fixture: the UTC day and
    // the reader's day differ for the first hours of every local day east of
    // Greenwich, and a fixture must not be the one deciding which.
    const now = new Date();
    const today = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
    expect(yearsOld(today)).toBe(0);
  });
});
