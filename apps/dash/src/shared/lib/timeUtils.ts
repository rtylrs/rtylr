/**
 * Date and time helpers for the Dash app.
 * All functions are pure with no side-effects.
 */

/**
 * Formats a duration in seconds into a human-readable string.
 * Omits zero parts. e.g. 3690 -> '1h 1m 30s'
 */
export function formatDuration(seconds: number): string {
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  const s = Math.floor(seconds % 60);
  const parts: string[] = [];
  if (h > 0) parts.push(`${h}h`);
  if (m > 0) parts.push(`${m}m`);
  if (s > 0) parts.push(`${s}s`);
  return parts.length > 0 ? parts.join(' ') : '0s';
}

/**
 * Returns a relative time string from a timestamp.
 * e.g. 'just now', '2 minutes ago', 'in 3 hours'
 */
export function relativeTime(dateMs: number, nowMs: number = Date.now()): string {
  const diffMs = dateMs - nowMs;
  const absDiff = Math.abs(diffMs);
  const past = diffMs < 0;

  if (absDiff < 30_000) return 'just now';

  const minutes = Math.round(absDiff / 60_000);
  if (minutes < 60) return past ? `${minutes} minutes ago` : `in ${minutes} minutes`;

  const hours = Math.round(absDiff / 3_600_000);
  if (hours < 24) return past ? `${hours} hours ago` : `in ${hours} hours`;

  const days = Math.round(absDiff / 86_400_000);
  return past ? `${days} days ago` : `in ${days} days`;
}

/** Returns a new Date set to midnight (start) of the given date. */
export function startOfDay(date: Date): Date {
  const d = new Date(date);
  d.setHours(0, 0, 0, 0);
  return d;
}

/** Returns a new Date set to 23:59:59.999 (end) of the given date. */
export function endOfDay(date: Date): Date {
  const d = new Date(date);
  d.setHours(23, 59, 59, 999);
  return d;
}

/** Returns true if two dates fall on the same calendar day. */
export function isSameDay(a: Date, b: Date): boolean {
  return (
    a.getFullYear() === b.getFullYear() &&
    a.getMonth() === b.getMonth() &&
    a.getDate() === b.getDate()
  );
}

/** Returns a new Date with the given number of days added (negative to subtract). */
export function addDays(date: Date, days: number): Date {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d;
}

/** Returns the absolute number of calendar days between two dates. */
export function daysBetween(a: Date, b: Date): number {
  const msPerDay = 86_400_000;
  return Math.round(Math.abs(startOfDay(a).getTime() - startOfDay(b).getTime()) / msPerDay);
}
