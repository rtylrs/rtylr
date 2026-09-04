/**
 * Today as YYYY-MM-DD in the reader's own timezone.
 *
 * Local, not UTC. A due date is a calendar day where the person is standing,
 * and `toISOString()` gives the UTC day: the two differ between local
 * midnight and UTC midnight, which is three hours a night at UTC+3 and the
 * whole evening further east. In that window `isToday` called a task due
 * yesterday today, `isPast` refused to call an overdue task overdue, and this
 * function disagreed with the identically named one in lifeRhythm.ts, so the
 * Today screen and a task row could label the same task differently on the
 * same render.
 *
 * lifeRhythm.ts now re-exports this one, so there is a single definition.
 */
export function todayIso(): string {
  const now = new Date();
  const year = now.getFullYear();
  const month = String(now.getMonth() + 1).padStart(2, '0');
  const day = String(now.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/** Short day-of-week key, Sunday first to match `Date#getUTCDay()`'s own indexing. */
export type DayKey = 'sun' | 'mon' | 'tue' | 'wed' | 'thu' | 'fri' | 'sat';

export const DAY_KEYS: readonly DayKey[] = ['sun', 'mon', 'tue', 'wed', 'thu', 'fri', 'sat'];

/**
 * Returns the day-of-week key for an ISO date string, anchored to UTC like
 * every other date helper here - parsing a bare "YYYY-MM-DD" without an
 * explicit UTC anchor lets the runtime's local offset shift midnight-ish
 * dates onto the wrong calendar day.
 */
export function dayKeyForDate(date: string): DayKey {
  const dt = new Date(date + 'T00:00:00Z');
  return DAY_KEYS[dt.getUTCDay()];
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

/**
 * Returns the ISO start-of-week for a given date string, Monday by default.
 * Pass 'sunday' for a caller (the Time week strip) whose workspace has
 * configured that week start; every other caller keeps the Monday default it
 * already relied on.
 */
export function startOfWeek(date: string, weekStart: 'monday' | 'sunday' = 'monday'): string {
  const [y, m, d] = date.split('-').map(Number);
  const dt = new Date(Date.UTC(y, m - 1, d));
  const day = dt.getUTCDay(); // 0 = Sunday
  const diff = weekStart === 'sunday' ? -day : day === 0 ? -6 : 1 - day;
  dt.setUTCDate(dt.getUTCDate() + diff);
  return dt.toISOString().slice(0, 10);
}

/** Returns the ISO end-of-week for a given date string, six days after startOfWeek. */
export function endOfWeek(date: string, weekStart: 'monday' | 'sunday' = 'monday'): string {
  return addDays(startOfWeek(date, weekStart), 6);
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

/**
 * Formats an ISO date string as a short weekday and day number, for the week
 * strip. Digits are pinned to Latin the same way formatEntryTimestamp does -
 * a week-strip button is still a date a person reads, not a duration, but the
 * reasoning is identical: nothing here should ever render in a locale's own
 * numeral system.
 */
export function formatWeekDayShort(date: string, locale = 'en-US'): string {
  const dt = new Date(date + 'T00:00:00Z');
  return new Intl.DateTimeFormat(`${locale}-u-nu-latn`, {
    weekday: 'short',
    day: 'numeric',
    timeZone: 'UTC',
  }).format(dt);
}

/**
 * Formats an ISO date string as a full weekday, day and month, e.g.
 * "Wednesday 12 November" - for a screen that names the day it is writing
 * against, rather than assuming the reader already has it in view. Digits
 * stay Latin for the same reason every other formatter here does: a date a
 * person reads is not exempt just because it isn't a duration.
 */
export function formatFullDayLabel(date: string, locale = 'en-US'): string {
  const dt = new Date(date + 'T00:00:00Z');
  return new Intl.DateTimeFormat(`${locale}-u-nu-latn`, {
    weekday: 'long',
    day: 'numeric',
    month: 'long',
    timeZone: 'UTC',
  }).format(dt);
}
