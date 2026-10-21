import { describe, it, expect } from 'vitest';
import {
  isToday,
  isThisMonth,
  daysAgo,
  daysFromNow,
  isBefore,
  isAfter,
  relativeDate,
  followUpStatus,
  formatDay,
} from './dateUtils';

describe('isToday', () => {
  it('returns true for today', () => {
    // Built from local parts. isToday reports the reader's day, so deriving
    // the fixture through toISOString made this test pass only while the UTC
    // day and the local day agreed: it failed at 00:08 local, three hours
    // before UTC midnight.
    const now = new Date();
    const today = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
    expect(isToday(today)).toBe(true);
  });

  it('returns false for yesterday', () => {
    const d = new Date();
    d.setDate(d.getDate() - 1);
    expect(isToday(d.toISOString().split('T')[0])).toBe(false);
  });
});

describe('isThisMonth', () => {
  it('returns true for a date in the current month', () => {
    const now = new Date();
    const dateStr = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-01`;
    expect(isThisMonth(dateStr)).toBe(true);
  });

  it('returns false for a date in a past month', () => {
    expect(isThisMonth('2020-01-15')).toBe(false);
  });
});

describe('daysAgo', () => {
  it('returns a date exactly N days before the reference', () => {
    const ref = new Date(2024, 4, 10);
    expect(daysAgo(5, ref)).toBe('2024-05-05');
  });

  it('returns today for 0 days ago', () => {
    const ref = new Date(2024, 5, 1);
    expect(daysAgo(0, ref)).toBe('2024-06-01');
  });
});

describe('daysFromNow', () => {
  it('returns a date exactly N days after the reference', () => {
    const ref = new Date(2024, 4, 10);
    expect(daysFromNow(5, ref)).toBe('2024-05-15');
  });

  it('returns today for 0 days from now', () => {
    const ref = new Date(2024, 5, 1);
    expect(daysFromNow(0, ref)).toBe('2024-06-01');
  });
});

describe('isBefore', () => {
  it('returns true when a is before b', () => {
    expect(isBefore('2024-01-01', '2024-06-01')).toBe(true);
  });

  it('returns false when a is after b', () => {
    expect(isBefore('2024-06-01', '2024-01-01')).toBe(false);
  });

  it('returns false for equal dates', () => {
    expect(isBefore('2024-01-01', '2024-01-01')).toBe(false);
  });
});

describe('isAfter', () => {
  it('returns true when a is after b', () => {
    expect(isAfter('2024-06-01', '2024-01-01')).toBe(true);
  });

  it('returns false when a is before b', () => {
    expect(isAfter('2024-01-01', '2024-06-01')).toBe(false);
  });
});

describe('relativeDate', () => {
  it('returns "Yesterday" for yesterday', () => {
    const ref = new Date();
    ref.setDate(ref.getDate() - 1);
    expect(relativeDate(ref.toISOString())).toBe('Yesterday');
  });

  it('returns "X days ago" for days within the past week', () => {
    const ref = new Date();
    ref.setDate(ref.getDate() - 3);
    expect(relativeDate(ref.toISOString())).toBe('3 days ago');
  });

  it('returns a formatted date for dates older than a week', () => {
    const result = relativeDate('2020-01-15T00:00:00.000Z');
    expect(result).toContain('2020');
    expect(result).toContain('Jan');
  });
});

describe('followUpStatus', () => {
  it('compares a YYYY-MM-DD follow-up against today', () => {
    expect(followUpStatus('2026-10-17', '2026-10-18')).toBe('overdue');
    expect(followUpStatus('2026-10-18', '2026-10-18')).toBe('today');
    expect(followUpStatus('2026-10-19', '2026-10-18')).toBe('upcoming');
  });
});

describe('formatDay', () => {
  it('shows the stored calendar day even when it arrives as UTC midnight', () => {
    // A local formatter reads 1990-05-14T00:00:00Z as 13 May west of UTC.
    expect(formatDay('1990-05-14T00:00:00Z', 'en-US')).toBe('May 14, 1990');
    expect(formatDay('1990-05-14', 'en-US')).toBe('May 14, 1990');
  });
});
