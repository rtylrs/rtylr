/**
 * Date and time helpers for the landing app.
 * Pure functions with no external dependencies, safe for server and browser.
 */

/**
 * Formats an ISO date string for display.
 * formatDisplayDate('2024-03-15') -> 'March 15, 2024'
 */
export function formatDisplayDate(isoDate: string, locale = 'en-US'): string {
  const d = new Date(isoDate);
  if (isNaN(d.getTime())) return '';
  return d.toLocaleDateString(locale, { year: 'numeric', month: 'long', day: 'numeric' });
}

/**
 * Formats an ISO date string as a short date.
 * formatShortDate('2024-03-15') -> 'Mar 15, 2024'
 */
export function formatShortDate(isoDate: string, locale = 'en-US'): string {
  const d = new Date(isoDate);
  if (isNaN(d.getTime())) return '';
  return d.toLocaleDateString(locale, { year: 'numeric', month: 'short', day: 'numeric' });
}

/**
 * Returns a relative time string like "3 days ago" or "in 2 hours".
 * Falls back to a formatted date when more than 30 days away.
 */
export function relativeTime(isoDate: string): string {
  const date = new Date(isoDate);
  if (isNaN(date.getTime())) return '';
  const now = Date.now();
  const diffMs = date.getTime() - now;
  const absSeconds = Math.abs(diffMs) / 1000;
  const past = diffMs < 0;

  if (absSeconds < 60) return past ? 'just now' : 'in a moment';
  if (absSeconds < 3600) {
    const mins = Math.round(absSeconds / 60);
    return past ? `${mins}m ago` : `in ${mins}m`;
  }
  if (absSeconds < 86400) {
    const hours = Math.round(absSeconds / 3600);
    return past ? `${hours}h ago` : `in ${hours}h`;
  }
  if (absSeconds < 86400 * 30) {
    const days = Math.round(absSeconds / 86400);
    return past ? `${days}d ago` : `in ${days}d`;
  }
  return formatShortDate(isoDate);
}

/**
 * Returns true when the given ISO date string falls on a weekend.
 */
export function isWeekend(isoDate: string): boolean {
  const d = new Date(isoDate);
  const day = d.getUTCDay();
  return day === 0 || day === 6;
}

/**
 * Returns the ISO date string for today.
 */
export function today(): string {
  return new Date().toISOString().slice(0, 10);
}

/**
 * Adds a number of days to an ISO date string and returns a new ISO date string.
 */
export function addDays(isoDate: string, days: number): string {
  const d = new Date(isoDate);
  d.setUTCDate(d.getUTCDate() + days);
  return d.toISOString().slice(0, 10);
}

/**
 * Returns the number of calendar days between two ISO date strings.
 */
export function daysBetween(startIso: string, endIso: string): number {
  const msPerDay = 1000 * 60 * 60 * 24;
  return Math.round((new Date(endIso).getTime() - new Date(startIso).getTime()) / msPerDay);
}

/**
 * Returns a human-friendly year range string: '2024' or '2024-2025'.
 */
export function yearRange(startYear: number, endYear?: number): string {
  if (!endYear || endYear === startYear) return String(startYear);
  return `${startYear}-${endYear}`;
}
