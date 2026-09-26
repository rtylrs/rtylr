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
  return toLocalIso(new Date());
}

/**
 * The calendar day a Date falls on where the reader is standing, YYYY-MM-DD.
 *
 * Use this instead of `toISOString().slice(0, 10)` for any Date built from
 * local parts. `new Date(2026, 8, 5)` is local midnight, which at UTC+3 is
 * 21:00 on the 4th in UTC, so `toISOString()` reports the wrong day for the
 * entire day rather than only near midnight. Clicking 5 September in the
 * calendar used to create a task due 4 September for every user east of
 * Greenwich.
 *
 * The helpers below that go through `Date.UTC(...)` and read back with
 * `getUTCDay`/`setUTCDate` are a different thing and are already correct: they
 * are UTC in and UTC out, so they never touch the local offset. Do not route
 * those through here.
 */
export function toLocalIso(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/**
 * Where a due date sits relative to the reader's own today.
 *
 * `'late'` is strictly before today, `'today'` is today, `'future'` is after,
 * and `'none'` is an absent or empty due date.
 *
 * Compared as strings, because a due date is a calendar day and not an instant.
 * Three components computed this as `new Date(task.dueDate) < new Date()`, which
 * parses a bare `YYYY-MM-DD` as **UTC midnight** and compares it to the current
 * instant. Measured at UTC+3 for a task due today:
 *
 *   checked 00:30 local -> not overdue
 *   checked 08:00 local -> OVERDUE
 *   checked 23:00 local -> OVERDUE
 *
 * So a task due today was painted as late from 00:00 UTC onward, which is the
 * whole working day in the product's home market (rtylr#1077). Further east it
 * starts earlier still.
 */
export type DueState = 'late' | 'today' | 'future' | 'none';

export function dueState(dueDate: string | null | undefined): DueState {
  const due = dueDate ? dueDate.split('T')[0] : '';
  if (due === '') return 'none';

  const today = todayIso();
  if (due < today) return 'late';
  if (due === today) return 'today';
  return 'future';
}

/**
 * Whether a task should read as overdue: past its day and not finished.
 *
 * Kept beside `dueState` rather than left to each caller, because the three
 * that hand-rolled it also each decided separately what "finished" meant.
 */
export function isOverdue(dueDate: string | null | undefined, status?: string): boolean {
  if (status === 'done' || status === 'cancelled') return false;
  return dueState(dueDate) === 'late';
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
