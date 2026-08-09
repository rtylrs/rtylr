import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';

import {
  todayIso,
  isDueToday,
  mondayStartMs,
  isDueThisCalendarWeek,
  parseTimeToMinutes,
  isQuietNow,
} from './lifeRhythm';

// ---------------------------------------------------------------------------
// todayIso
// ---------------------------------------------------------------------------

describe('todayIso', () => {
  it('returns a YYYY-MM-DD string', () => {
    expect(todayIso()).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });

  it('matches todays local date', () => {
    const t = new Date();
    const y = t.getFullYear();
    const m = String(t.getMonth() + 1).padStart(2, '0');
    const d = String(t.getDate()).padStart(2, '0');
    expect(todayIso()).toBe(`${y}-${m}-${d}`);
  });
});

// ---------------------------------------------------------------------------
// isDueToday - pin system clock
// ---------------------------------------------------------------------------

describe('isDueToday', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T12:00:00.000Z'));
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it("returns true for today's ISO date", () => {
    expect(isDueToday('2026-05-02')).toBe(true);
  });

  it('returns false for yesterday', () => {
    expect(isDueToday('2026-05-01')).toBe(false);
  });

  it('returns false for tomorrow', () => {
    expect(isDueToday('2026-05-03')).toBe(false);
  });

  it('returns false for undefined', () => {
    expect(isDueToday(undefined)).toBe(false);
  });

  it('returns false for empty string', () => {
    expect(isDueToday('')).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// mondayStartMs - pinned to a known Saturday
// ---------------------------------------------------------------------------

describe('mondayStartMs', () => {
  it('returns the Monday before a given Saturday', () => {
    // 2026-05-02 is a Saturday; Monday of that week is 2026-04-27.
    const sat = new Date('2026-05-02T12:00:00.000Z');
    const ms = mondayStartMs(sat);
    const monday = new Date(ms);
    expect(monday.getDay()).toBe(1); // 1 = Monday
    expect(monday.getHours()).toBe(0);
    expect(monday.getMinutes()).toBe(0);
  });

  it('called on a Monday returns that Monday at midnight', () => {
    const mon = new Date('2026-04-27T14:00:00.000Z');
    const ms = mondayStartMs(mon);
    const result = new Date(ms);
    expect(result.getDay()).toBe(1);
    expect(result.getHours()).toBe(0);
  });
});

// ---------------------------------------------------------------------------
// isDueThisCalendarWeek
// ---------------------------------------------------------------------------

describe('isDueThisCalendarWeek', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    // 2026-05-02 is Saturday; this week = Mon 2026-04-27 .. Sun 2026-05-03
    vi.setSystemTime(new Date('2026-05-02T12:00:00.000Z'));
  });
  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns true for a date within the current Mon–Sun window', () => {
    expect(isDueThisCalendarWeek('2026-04-28')).toBe(true);
    expect(isDueThisCalendarWeek('2026-05-01')).toBe(true);
    expect(isDueThisCalendarWeek('2026-05-02')).toBe(true);
  });

  it('returns false for a date outside the current week', () => {
    expect(isDueThisCalendarWeek('2026-05-04')).toBe(false); // next week Monday
    expect(isDueThisCalendarWeek('2026-04-26')).toBe(false); // previous Sunday
  });

  it('returns false for undefined', () => {
    expect(isDueThisCalendarWeek(undefined)).toBe(false);
  });
});

// ---------------------------------------------------------------------------
// parseTimeToMinutes
// ---------------------------------------------------------------------------

describe('parseTimeToMinutes', () => {
  it('converts HH:MM to minutes', () => {
    expect(parseTimeToMinutes('09:30')).toBe(570);
    expect(parseTimeToMinutes('00:00')).toBe(0);
    expect(parseTimeToMinutes('23:59')).toBe(1439);
  });

  it('handles whole-hour values', () => {
    expect(parseTimeToMinutes('08:00')).toBe(480);
  });

  it('returns 0 for invalid input', () => {
    expect(parseTimeToMinutes('notATime')).toBe(0);
  });
});

// ---------------------------------------------------------------------------
// isQuietNow
// ---------------------------------------------------------------------------

describe('isQuietNow', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns true when now is inside a normal (non-overnight) quiet window', () => {
    // Quiet 22:00–07:00; pin time to 23:00 (1380 mins)
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T23:00:00'));
    expect(isQuietNow('22:00', '23:59')).toBe(true);
  });

  it('returns false when now is outside the quiet window', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T14:00:00')); // 14:00 local
    expect(isQuietNow('22:00', '07:00')).toBe(false);
  });

  it('handles overnight quiet window - now after start', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T23:30:00')); // 23:30
    expect(isQuietNow('22:00', '07:00')).toBe(true);
  });

  it('handles overnight quiet window - now before end', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-05-02T06:00:00')); // 06:00
    expect(isQuietNow('22:00', '07:00')).toBe(true);
  });
});
