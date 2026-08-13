/** Returns today's date as an ISO string (YYYY-MM-DD). */
export function todayIso(): string {
  return new Date().toISOString().slice(0, 10);
}

/** Returns true if the date string is before today. */
export function isPast(date: string): boolean {
  return date < todayIso();
}

/** Returns true if the date string is today. */
export function isToday(date: string): boolean {
  return date === todayIso();
}

/** Returns true if the date string is after today. */
export function isFuture(date: string): boolean {
  return date > todayIso();
}

/** Returns the number of calendar days between two ISO date strings. */
export function daysBetween(a: string, b: string): number {
  const msPerDay = 86_400_000;
  return Math.round((new Date(b).getTime() - new Date(a).getTime()) / msPerDay);
}

/** Returns the number of days until a future date (negative if past). */
export function daysUntil(date: string, from?: string): number {
  const base = from ?? todayIso();
  return daysBetween(base, date);
}

/** Returns a human-readable relative label: "Today", "Tomorrow", "Yesterday", or "N days". */
export function relativeDayLabel(date: string, from?: string): string {
  const diff = daysUntil(date, from);
  if (diff === 0) return 'Today';
  if (diff === 1) return 'Tomorrow';
  if (diff === -1) return 'Yesterday';
  if (diff > 0) return `In ${diff} days`;
  return `${Math.abs(diff)} days ago`;
}

/** Returns true if the date falls within the given number of days from now. */
export function isWithinDays(date: string, days: number, from?: string): boolean {
  const diff = daysUntil(date, from);
  return diff >= 0 && diff <= days;
}

/** Formats an ISO date string as a short readable label, e.g. "Jan 5, 2026". */
export function formatDateShort(date: string): string {
  return new Date(date + 'T00:00:00').toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });
}

/** Adds the given number of days to an ISO date string, returning a new ISO string. */
export function addDays(date: string, days: number): string {
  const [y, m, d] = date.split('-').map(Number);
  const dt = new Date(Date.UTC(y, m - 1, d + days));
  return dt.toISOString().slice(0, 10);
}

/** Returns the ISO start-of-week (Monday) for a given date string. */
export function startOfWeek(date: string): string {
  const [y, m, d] = date.split('-').map(Number);
  const dt = new Date(Date.UTC(y, m - 1, d));
  const day = dt.getUTCDay(); // 0 = Sunday
  const diff = day === 0 ? -6 : 1 - day;
  dt.setUTCDate(dt.getUTCDate() + diff);
  return dt.toISOString().slice(0, 10);
}

/** Returns the ISO end-of-week (Sunday) for a given date string. */
export function endOfWeek(date: string): string {
  return addDays(startOfWeek(date), 6);
}

/**
 * Formats a full ISO timestamp as `Tue 11 Aug`, for the provenance line under a
 * time entry.
 *
 * Distinct from `formatDateShort`, which takes a date-only string and appends
 * `T00:00:00` - handing it a timestamp produces an invalid date.
 *
 * Month and weekday names follow the active language; the digits do not. The
 * `nu-latn` extension pins Latin numerals, because a duration or a date read
 * back to a payroll import must not arrive in Eastern Arabic digits.
 */
export function formatEntryTimestamp(iso: string, locale = 'en-US'): string {
  const dt = new Date(iso);
  if (Number.isNaN(dt.getTime())) return '';
  return new Intl.DateTimeFormat(`${locale}-u-nu-latn`, {
    weekday: 'short',
    day: 'numeric',
    month: 'short',
    timeZone: 'UTC',
  }).format(dt);
}
