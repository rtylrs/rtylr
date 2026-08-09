/**
 * Date utility functions used across the ERP application.
 */

/** Return the start of the day (midnight) for a given date. */
export function startOfDay(date: Date): Date {
  const d = new Date(date);
  d.setHours(0, 0, 0, 0);
  return d;
}

/** Return the end of the day (23:59:59.999) for a given date. */
export function endOfDay(date: Date): Date {
  const d = new Date(date);
  d.setHours(23, 59, 59, 999);
  return d;
}

/** Return a date N days from now. */
export function addDays(date: Date, days: number): Date {
  const d = new Date(date);
  d.setDate(d.getDate() + days);
  return d;
}

/** Return true when a date falls within a range (inclusive). */
export function isWithinRange(date: Date, start: Date, end: Date): boolean {
  return date >= start && date <= end;
}

/** Return true when a date is in the past. */
export function isPast(date: Date, now = new Date()): boolean {
  return date < now;
}

/** Return true when a date is in the future. */
export function isFuture(date: Date, now = new Date()): boolean {
  return date > now;
}

/** Return the ISO date string (YYYY-MM-DD) for a Date object. */
export function toIsoDate(date: Date): string {
  return date.toISOString().split('T')[0];
}

/** Return the number of calendar days between two dates. */
export function daysBetween(start: Date, end: Date): number {
  const ms = end.getTime() - start.getTime();
  return Math.round(ms / (1000 * 60 * 60 * 24));
}

/** Return a Date from an ISO date string without timezone offset issues. */
export function parseIsoDate(dateString: string): Date {
  const [year, month, day] = dateString.split('-').map(Number);
  return new Date(year, month - 1, day);
}
