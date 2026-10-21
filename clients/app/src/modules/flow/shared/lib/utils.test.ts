import path from 'node:path';
import i18next from 'i18next';
import { initTestI18nFrom } from '../../../../../test/i18nTestInstance';
import { describe, it, expect, beforeEach, vi, afterEach, beforeAll } from 'vitest';

import { cn, getInitials, formatDueDate, formatShortDate } from './utils';

// ---------------------------------------------------------------------------
// cn - className merger
// ---------------------------------------------------------------------------

describe('cn', () => {
  it('joins two truthy classes', () => {
    expect(cn('foo', 'bar')).toBe('foo bar');
  });

  it('strips falsy values', () => {
    expect(cn('a', false, null, undefined, 0, 'b')).toBe('a b');
  });

  it('returns empty string when all values are falsy', () => {
    expect(cn(false, null, undefined)).toBe('');
  });

  it('single class passes through', () => {
    expect(cn('only')).toBe('only');
  });
});

// ---------------------------------------------------------------------------
// getInitials
// ---------------------------------------------------------------------------

describe('getInitials', () => {
  it('returns ? for null', () => {
    expect(getInitials(null)).toBe('?');
  });

  it('returns ? for undefined', () => {
    expect(getInitials(undefined)).toBe('?');
  });

  it('returns ? for empty string', () => {
    expect(getInitials('')).toBe('?');
  });

  it('one word → first letter uppercased', () => {
    expect(getInitials('alice')).toBe('A');
  });

  it('two words → both first letters uppercased', () => {
    expect(getInitials('Alice Brown')).toBe('AB');
  });

  it('caps at two characters for long names', () => {
    expect(getInitials('Alice Brown Charlie')).toBe('AB');
  });
});

// ---------------------------------------------------------------------------
// formatShortDate and formatDueDate (rtylr#1221)
// ---------------------------------------------------------------------------

beforeAll(() => initTestI18nFrom(path.resolve(import.meta.dirname, '../../locales')));
const t = i18next.t.bind(i18next);

describe('formatShortDate', () => {
  it('returns a dash for an empty date', () => {
    expect(formatShortDate('')).toBe('-');
  });

  it('formats a date as short month and day, never as an overdue label', () => {
    // A cycle that began last week is running, not overdue.
    expect(formatShortDate('2026-05-15')).toMatch(/May\s*15/i);
  });
});

describe('formatDueDate', () => {
  beforeEach(() => {
    // Pin "now" to 2026-05-02T12:00:00Z so relative labels are stable.
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T12:00:00.000Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns "-" for an empty string', () => {
    expect(formatDueDate('', t)).toBe('-');
  });

  it('returns "-" for a whitespace-only string', () => {
    expect(formatDueDate('   ', t)).toBe('-');
  });

  it('returns the translated today label', () => {
    expect(formatDueDate('2026-05-02', t)).toBe('Today');
  });

  it('returns the translated tomorrow label', () => {
    expect(formatDueDate('2026-05-03', t)).toBe('Tomorrow');
  });

  it('returns the translated yesterday label', () => {
    expect(formatDueDate('2026-05-01', t)).toBe('Yesterday');
  });

  it('returns the translated overdue label with the day count', () => {
    expect(formatDueDate('2026-04-29', t)).toBe('3d overdue');
  });

  it('returns a short date for something further out', () => {
    expect(formatDueDate('2026-05-15', t)).toMatch(/May\s*15/i);
  });
});

describe('formatDueDate counts calendar days, not hours (#1097 D-020)', () => {
  const ORIGINAL_TZ = process.env.TZ;

  afterEach(() => {
    vi.useRealTimers();
    if (ORIGINAL_TZ === undefined) delete process.env.TZ;
    else process.env.TZ = ORIGINAL_TZ;
  });

  const at = (localIso: string): void => {
    process.env.TZ = 'Asia/Beirut';
    vi.useFakeTimers();
    vi.setSystemTime(new Date(localIso));
  };

  it('calls a task due today "Today" at 02:00, not "Tomorrow"', () => {
    // The old sum subtracted a local now from a UTC-parsed date, so in the
    // hours before UTC midnight a task due today came out one day ahead. Only
    // this hour distinguishes the correct answer from the broken one.
    at('2026-09-07T02:00:00');
    expect(formatDueDate('2026-09-07', t)).toBe('Today');
  });

  it('gives the same answer at every hour of that day', () => {
    for (const hour of ['00:30', '02:00', '09:00', '14:00', '23:45']) {
      at(`2026-09-07T${hour}:00`);
      expect(formatDueDate('2026-09-07', t), `at ${hour}`).toBe('Today');
      vi.useRealTimers();
    }
  });

  it('still labels the neighbouring days', () => {
    at('2026-09-07T02:00:00');
    expect(formatDueDate('2026-09-08', t)).toBe('Tomorrow');
    expect(formatDueDate('2026-09-06', t)).toBe('Yesterday');
  });

  it('returns a dash for an absent date rather than Invalid Date', () => {
    at('2026-09-07T09:00:00');
    expect(formatDueDate('', t)).toBe('-');
  });
});
