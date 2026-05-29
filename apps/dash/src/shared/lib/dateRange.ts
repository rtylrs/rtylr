/**
 * Date range helpers for the Dash analytics filters.
 * All functions are pure with no side-effects.
 */

export type DateRangePreset = '7d' | '30d' | '90d' | '1y' | 'mtd' | 'ytd' | 'custom';

export interface DateRange {
  start: Date;
  end: Date;
}

/** Add days to a date, returning a new Date. */
export function addDays(date: Date, days: number): Date {
  const result = new Date(date.getTime());
  result.setDate(result.getDate() + days);
  return result;
}

/** Return true if a date falls within the given range (inclusive). */
export function isWithinRange(date: Date, range: DateRange): boolean {
  return date >= range.start && date <= range.end;
}

/** Return the number of whole days between two dates (absolute value). */
export function daysBetween(a: Date, b: Date): number {
  const msPerDay = 1000 * 60 * 60 * 24;
  return Math.round(Math.abs(b.getTime() - a.getTime()) / msPerDay);
}

/** Convert a preset string to a DateRange relative to now. */
export function presetToRange(preset: DateRangePreset, now = new Date()): DateRange {
  const end = new Date(now);
  switch (preset) {
    case '7d':
      return { start: addDays(now, -7), end };
    case '30d':
      return { start: addDays(now, -30), end };
    case '90d':
      return { start: addDays(now, -90), end };
    case '1y':
      return { start: addDays(now, -365), end };
    case 'mtd': {
      const start = new Date(now.getFullYear(), now.getMonth(), 1);
      return { start, end };
    }
    case 'ytd': {
      const start = new Date(now.getFullYear(), 0, 1);
      return { start, end };
    }
    case 'custom':
      return { start: now, end };
    default:
      return { start: now, end };
  }
}

/** Format a DateRange as a human-readable string. */
export function formatDateRange(range: DateRange, locale = 'en-US'): string {
  const opts: Intl.DateTimeFormatOptions = { month: 'short', day: 'numeric', year: 'numeric' };
  const fmt = new Intl.DateTimeFormat(locale, opts);
  return `${fmt.format(range.start)} - ${fmt.format(range.end)}`;
}
