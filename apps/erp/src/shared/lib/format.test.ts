import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import {
  formatCurrency,
  formatNumber,
  formatDate,
  formatDateTime,
  timeAgo,
  daysUntil,
} from './format';

// ---------------------------------------------------------------------------
// formatCurrency
// ---------------------------------------------------------------------------

describe('formatCurrency', () => {
  it('formats a whole dollar amount', () => {
    expect(formatCurrency(10)).toBe('$10.00');
  });

  it('formats cents correctly', () => {
    expect(formatCurrency(3.5)).toBe('$3.50');
  });

  it('formats zero as $0.00', () => {
    expect(formatCurrency(0)).toBe('$0.00');
  });

  it('formats large amounts with comma separators', () => {
    // en-US locale adds commas for thousands.
    expect(formatCurrency(1000000)).toMatch(/1,000,000\.00/);
  });

  it('respects a custom currency code', () => {
    const result = formatCurrency(25, 'EUR');
    expect(result).toMatch(/25\.00/);
    // The exact EUR symbol position is locale-dependent, but the digits must be there.
  });
});

// ---------------------------------------------------------------------------
// formatNumber
// ---------------------------------------------------------------------------

describe('formatNumber', () => {
  it('formats a plain integer', () => {
    expect(formatNumber(42)).toBe('42');
  });

  it('adds comma separators for thousands', () => {
    expect(formatNumber(1500)).toBe('1,500');
  });

  it('formats zero', () => {
    expect(formatNumber(0)).toBe('0');
  });
});

// ---------------------------------------------------------------------------
// formatDate - uses real Date so we just check format, not relative labels
// ---------------------------------------------------------------------------

describe('formatDate', () => {
  it('returns a string for a valid ISO date', () => {
    const result = formatDate('2026-01-15');
    // en-US: "Jan 15, 2026"
    expect(result).toMatch(/Jan\s*15,\s*2026/i);
  });

  it('returns "Invalid Date" string when given an unparseable date string', () => {
    // toLocaleDateString on an invalid Date returns the string "Invalid Date"
    // rather than throwing, so the try-block succeeds with that value.
    expect(formatDate('not-a-date')).toBe('Invalid Date');
  });

  it('handles an empty string gracefully (returns "Invalid Date" fallback)', () => {
    // new Date('') produces an Invalid Date; toLocaleDateString returns the
    // string "Invalid Date" rather than throwing, so the function returns it.
    const result = formatDate('');
    expect(result).toBe('Invalid Date');
  });
});

// ---------------------------------------------------------------------------
// formatDateTime
// ---------------------------------------------------------------------------

describe('formatDateTime', () => {
  it('returns a string containing the month and day', () => {
    // A timestamp is an instant, so it correctly renders in the reader's own
    // zone: 14:30Z is already 5 March in Sydney. The zone is pinned here
    // rather than inherited so the assertion means something either way.
    const original = process.env.TZ;
    process.env.TZ = 'UTC';
    try {
      expect(formatDateTime('2026-03-04T14:30:00Z')).toMatch(/Mar\s*4/i);
    } finally {
      if (original === undefined) delete process.env.TZ;
      else process.env.TZ = original;
    }
  });

  it('returns "Invalid Date" string when given an unparseable date string', () => {
    // Same as formatDate: toLocaleString on an Invalid Date returns "Invalid Date".
    expect(formatDateTime('bad')).toBe('Invalid Date');
  });
});

// ---------------------------------------------------------------------------
// timeAgo - pin "now" with fake timers
// ---------------------------------------------------------------------------

describe('timeAgo', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T12:00:00.000Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns "just now" for a timestamp 30 seconds ago', () => {
    expect(timeAgo('2026-05-02T11:59:30.000Z')).toBe('just now');
  });

  it('returns "Xm ago" for a timestamp 5 minutes ago', () => {
    expect(timeAgo('2026-05-02T11:55:00.000Z')).toBe('5m ago');
  });

  it('returns "Xh ago" for a timestamp 3 hours ago', () => {
    expect(timeAgo('2026-05-02T09:00:00.000Z')).toBe('3h ago');
  });

  it('returns "Xd ago" for a timestamp 2 days ago', () => {
    expect(timeAgo('2026-04-30T12:00:00.000Z')).toBe('2d ago');
  });

  it('falls back to formatDate for dates more than a week ago', () => {
    // The fallback renders an instant in the reader's own zone, where 12:00Z
    // on 1 April is already the 2nd at UTC+14. Pinned so the assertion holds
    // wherever this runs.
    const original = process.env.TZ;
    process.env.TZ = 'UTC';
    try {
      const result = timeAgo('2026-04-01T12:00:00.000Z');
      // Should be a formatted date, not a relative label.
      expect(result).toMatch(/Apr\s*1,\s*2026/i);
    } finally {
      if (original === undefined) delete process.env.TZ;
      else process.env.TZ = original;
    }
  });

  it('returns "Invalid Date" string for an unparseable date', () => {
    // new Date('invalid') → NaN diff → falls into the >week branch →
    // calls formatDate which returns the "Invalid Date" string.
    expect(timeAgo('invalid')).toBe('Invalid Date');
  });
});

// ---------------------------------------------------------------------------
// daysUntil
// ---------------------------------------------------------------------------

describe('daysUntil', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T00:00:00.000Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns 0 for today', () => {
    expect(daysUntil('2026-05-02')).toBe(0);
  });

  it('returns 1 for tomorrow', () => {
    expect(daysUntil('2026-05-03')).toBe(1);
  });

  it('returns 7 for one week from now', () => {
    expect(daysUntil('2026-05-09')).toBe(7);
  });

  it('returns a negative number for a past date', () => {
    expect(daysUntil('2026-05-01')).toBe(-1);
  });
});
