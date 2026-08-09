/**
 * Date utility helpers for the HR module.
 * All functions are side-effect free and operate on ISO date strings.
 */

/**
 * Returns today's date as an ISO date string (YYYY-MM-DD).
 */
export function todayIso(): string {
  return new Date().toISOString().slice(0, 10);
}

/**
 * Formats an ISO date string into a locale-aware display format.
 * Example: "2025-06-15" produces "Jun 15, 2025".
 */
export function formatDate(isoDate: string, locale = 'en-US'): string {
  return new Date(isoDate).toLocaleDateString(locale, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

/**
 * Returns the number of calendar days between two ISO date strings.
 * A positive value means end is after start.
 */
export function daysBetween(startIso: string, endIso: string): number {
  const start = new Date(startIso).getTime();
  const end = new Date(endIso).getTime();
  return Math.round((end - start) / (1000 * 60 * 60 * 24));
}

/**
 * Returns true when the given ISO date is strictly before today.
 */
export function isPast(isoDate: string): boolean {
  return isoDate < todayIso();
}

/**
 * Returns true when the given ISO date matches today.
 */
export function isToday(isoDate: string): boolean {
  return isoDate === todayIso();
}

/**
 * Adds a specified number of calendar days to an ISO date string.
 */
export function addDays(isoDate: string, days: number): string {
  const d = new Date(isoDate);
  d.setUTCDate(d.getUTCDate() + days);
  return d.toISOString().slice(0, 10);
}

/**
 * Returns the ISO week number (1-53) for an ISO date string.
 */
export function isoWeekNumber(isoDate: string): number {
  const d = new Date(isoDate);
  const dayOfWeek = d.getUTCDay() || 7;
  d.setUTCDate(d.getUTCDate() + 4 - dayOfWeek);
  const yearStart = new Date(Date.UTC(d.getUTCFullYear(), 0, 1));
  return Math.ceil(((d.getTime() - yearStart.getTime()) / 86400000 + 1) / 7);
}

/**
 * Returns the Monday of the week containing the given ISO date.
 */
export function weekStart(isoDate: string): string {
  const d = new Date(isoDate);
  const day = d.getUTCDay() || 7;
  d.setUTCDate(d.getUTCDate() - (day - 1));
  return d.toISOString().slice(0, 10);
}
