// Date formatting and parsing utilities for the insights app.
// These helpers produce consistent display strings for axis labels,
// tooltips, and table cells without requiring a date library.

/**
 * toISODate converts a Date to 'YYYY-MM-DD' using UTC to avoid timezone
 * drift when the app is used in non-UTC locales.
 */
export function toISODate(date: Date): string {
  const y = date.getUTCFullYear();
  const m = String(date.getUTCMonth() + 1).padStart(2, '0');
  const d = String(date.getUTCDate()).padStart(2, '0');
  return `${y}-${m}-${d}`;
}

/**
 * addDays returns a new Date n days after (or before, for negative n)
 * the given date, computed in UTC.
 */
export function addDays(date: Date, n: number): Date {
  const ms = date.getTime() + n * 24 * 60 * 60 * 1000;
  return new Date(ms);
}

/**
 * diffDays returns the number of calendar days between two ISO date
 * strings, always positive. Useful for "this period vs last period"
 * length calculations.
 */
export function diffDays(isoA: string, isoB: string): number {
  const a = new Date(isoA + 'T00:00:00Z').getTime();
  const b = new Date(isoB + 'T00:00:00Z').getTime();
  return Math.abs(Math.round((b - a) / (1000 * 60 * 60 * 24)));
}

/**
 * formatAxisDate formats an ISO date string for a chart x-axis label.
 * Uses short month names to save space. e.g. '2026-05-15' -> 'May 15'
 */
export function formatAxisDate(iso: string): string {
  const [, month, day] = iso.split('-').map(Number);
  const months = [
    'Jan',
    'Feb',
    'Mar',
    'Apr',
    'May',
    'Jun',
    'Jul',
    'Aug',
    'Sep',
    'Oct',
    'Nov',
    'Dec',
  ];
  return `${months[month - 1]} ${day}`;
}

/**
 * formatFullDate formats an ISO date string with year for tooltip and
 * table display. e.g. '2026-05-15' -> 'May 15, 2026'
 */
export function formatFullDate(iso: string): string {
  const [year, month, day] = iso.split('-').map(Number);
  const months = [
    'January',
    'February',
    'March',
    'April',
    'May',
    'June',
    'July',
    'August',
    'September',
    'October',
    'November',
    'December',
  ];
  return `${months[month - 1]} ${day}, ${year}`;
}

/**
 * isSameMonth returns true when two ISO date strings share the same
 * year and month.
 */
export function isSameMonth(isoA: string, isoB: string): boolean {
  return isoA.slice(0, 7) === isoB.slice(0, 7);
}

/**
 * startOfMonth returns the first day of the month for a given ISO date
 * string as 'YYYY-MM-01'.
 */
export function startOfMonth(iso: string): string {
  return iso.slice(0, 7) + '-01';
}
