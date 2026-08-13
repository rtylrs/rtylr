import { describe, it, expect } from 'vitest';

import {
  isPast,
  isToday,
  isFuture,
  daysBetween,
  daysUntil,
  relativeDayLabel,
  isWithinDays,
  formatDateShort,
  formatEntryTimestamp,
  addDays,
  startOfWeek,
  endOfWeek,
} from './dateUtils';

const TODAY = '2026-06-15';

describe('isPast', () => {
  it('returns true for a date before today', () => {
    expect(isPast('2026-01-01')).toBe(true);
  });
});

describe('isToday', () => {
  it('matches today exactly', () => {
    const today = new Date().toISOString().slice(0, 10);
    expect(isToday(today)).toBe(true);
  });
});

describe('isFuture', () => {
  it('returns true for a far future date', () => {
    expect(isFuture('2099-12-31')).toBe(true);
  });
});

describe('daysBetween', () => {
  it('counts days between two dates', () => {
    expect(daysBetween('2026-01-01', '2026-01-08')).toBe(7);
  });

  it('returns 0 for same date', () => {
    expect(daysBetween('2026-06-01', '2026-06-01')).toBe(0);
  });

  it('returns negative for reversed order', () => {
    expect(daysBetween('2026-01-08', '2026-01-01')).toBe(-7);
  });
});

describe('daysUntil', () => {
  it('returns positive for a future date', () => {
    expect(daysUntil('2026-06-20', TODAY)).toBe(5);
  });

  it('returns 0 for today', () => {
    expect(daysUntil(TODAY, TODAY)).toBe(0);
  });

  it('returns negative for a past date', () => {
    expect(daysUntil('2026-06-10', TODAY)).toBe(-5);
  });
});

describe('relativeDayLabel', () => {
  it('returns Today for same date', () => {
    expect(relativeDayLabel(TODAY, TODAY)).toBe('Today');
  });

  it('returns Tomorrow for next day', () => {
    expect(relativeDayLabel('2026-06-16', TODAY)).toBe('Tomorrow');
  });

  it('returns Yesterday for prior day', () => {
    expect(relativeDayLabel('2026-06-14', TODAY)).toBe('Yesterday');
  });

  it('returns In N days for future', () => {
    expect(relativeDayLabel('2026-06-20', TODAY)).toBe('In 5 days');
  });

  it('returns N days ago for past', () => {
    expect(relativeDayLabel('2026-06-10', TODAY)).toBe('5 days ago');
  });
});

describe('isWithinDays', () => {
  it('returns true when within range', () => {
    expect(isWithinDays('2026-06-17', 3, TODAY)).toBe(true);
  });

  it('returns false when beyond range', () => {
    expect(isWithinDays('2026-06-25', 3, TODAY)).toBe(false);
  });

  it('returns false for past dates', () => {
    expect(isWithinDays('2026-06-10', 5, TODAY)).toBe(false);
  });
});

describe('formatDateShort', () => {
  it('formats a date as short label', () => {
    const label = formatDateShort('2026-01-05');
    expect(label).toContain('Jan');
    expect(label).toContain('5');
    expect(label).toContain('2026');
  });
});

describe('addDays', () => {
  it('adds days to a date', () => {
    expect(addDays('2026-01-01', 7)).toBe('2026-01-08');
  });

  it('handles month boundaries', () => {
    expect(addDays('2026-01-28', 5)).toBe('2026-02-02');
  });
});

describe('startOfWeek', () => {
  it('returns Monday for a Wednesday', () => {
    // 2026-06-15 is a Monday
    expect(startOfWeek('2026-06-17')).toBe('2026-06-15');
  });

  it('returns same day for a Monday', () => {
    expect(startOfWeek('2026-06-15')).toBe('2026-06-15');
  });
});

describe('endOfWeek', () => {
  it('returns Sunday six days after Monday', () => {
    expect(endOfWeek('2026-06-15')).toBe('2026-06-21');
  });
});

describe('formatEntryTimestamp', () => {
  it('renders a weekday, day and month for a full ISO timestamp', () => {
    expect(formatEntryTimestamp('2026-08-11T09:00:00Z', 'en-US')).toBe('Tue, Aug 11');
  });

  it('follows each locale’s own day/month order rather than pinning one', () => {
    expect(formatEntryTimestamp('2026-08-11T09:00:00Z', 'en-GB')).toBe('Tue 11 Aug');
  });

  it('localises the month name but keeps Latin numerals in Arabic', () => {
    const formatted = formatEntryTimestamp('2026-08-11T09:00:00Z', 'ar');

    expect(formatted).toMatch(/11/);
    expect(formatted).not.toMatch(/[٠-٩]/);
  });

  it('returns an empty string for an unparseable value rather than "Invalid Date"', () => {
    expect(formatEntryTimestamp('', 'en-US')).toBe('');
    expect(formatEntryTimestamp('not a date', 'en-US')).toBe('');
  });
});
