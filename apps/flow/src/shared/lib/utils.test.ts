import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';

import { cn, getInitials, formatDate } from './utils';

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
// formatDate
// ---------------------------------------------------------------------------

describe('formatDate', () => {
  beforeEach(() => {
    // Pin "now" to 2026-05-02T12:00:00Z so relative labels are stable.
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T12:00:00.000Z'));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns "-" for an empty string', () => {
    expect(formatDate('')).toBe('-');
  });

  it('returns "-" for a whitespace-only string', () => {
    expect(formatDate('   ')).toBe('-');
  });

  it('returns "Today" for today\'s date', () => {
    expect(formatDate('2026-05-02')).toBe('Today');
  });

  it('returns "Tomorrow" for tomorrow', () => {
    expect(formatDate('2026-05-03')).toBe('Tomorrow');
  });

  it('returns "Yesterday" for yesterday', () => {
    expect(formatDate('2026-05-01')).toBe('Yesterday');
  });

  it('returns overdue label for a past date beyond yesterday', () => {
    const result = formatDate('2026-04-29'); // 3 days ago
    expect(result).toMatch(/overdue/i);
    expect(result).toContain('3');
  });

  it('returns a month + day string for a future date more than a day out', () => {
    const result = formatDate('2026-05-15');
    // en-US short month + day: "May 15"
    expect(result).toMatch(/May\s*15/i);
  });
});
