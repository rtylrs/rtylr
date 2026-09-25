import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
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
  /**
   * The clock is pinned and the dates are adjacent to it, because the previous
   * version of these tests used `2000-01-01` and `2099-12-31`. Ninety-nine
   * years apart, they passed under the broken implementation and the correct
   * one alike, so they said nothing about the bug they were meant to cover
   * (rtylr#1077).
   *
   * The interesting cases are all within a day of today: that is where a UTC
   * midnight parsed from a local day lands on the wrong side of "now".
   */
  beforeEach(() => {
    vi.useFakeTimers();
    // 08:00 in Beirut, so the local day and the UTC day agree and these
    // assertions are about the rule rather than the offset.
    vi.setSystemTime(new Date('2026-05-02T05:00:00.000Z'));
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('is false for today', () => {
    expect(isFutureDate('2026-05-02')).toBe(false);
  });

  it('is true for tomorrow', () => {
    expect(isFutureDate('2026-05-03')).toBe(true);
  });

  it('is false for yesterday', () => {
    expect(isFutureDate('2026-05-01')).toBe(false);
  });

  it('reads a timestamped value as the day it names', () => {
    expect(isFutureDate('2026-05-03T00:00:00Z')).toBe(true);
    expect(isFutureDate('2026-05-02T23:59:00Z')).toBe(false);
  });
});

describe('isPastDate', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T05:00:00.000Z'));
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('is false for today, which is not past', () => {
    // The defect. `new Date('2026-05-02') < new Date()` is true at this
    // instant, so today read as past for the whole working day.
    expect(new Date('2026-05-02') < new Date()).toBe(true);
    expect(isPastDate('2026-05-02')).toBe(false);
  });

  it('is true for yesterday', () => {
    expect(isPastDate('2026-05-01')).toBe(true);
  });

  it('is false for tomorrow', () => {
    expect(isPastDate('2026-05-03')).toBe(false);
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
