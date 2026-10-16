import { todayIso } from './dateUtils';

export function formatDate(
  dateIso: string,
  format: 'short' | 'long' | 'numeric' = 'short'
): string {
  const d = new Date(dateIso);
  if (isNaN(d.getTime())) return '';
  if (format === 'numeric') return d.toLocaleDateString('en-US', { timeZone: 'UTC' });
  if (format === 'long')
    return d.toLocaleDateString('en-US', {
      timeZone: 'UTC',
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  return d.toLocaleDateString('en-US', {
    timeZone: 'UTC',
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

export function daysBetween(startIso: string, endIso: string): number {
  const msPerDay = 1000 * 60 * 60 * 24;
  return Math.round((new Date(endIso).getTime() - new Date(startIso).getTime()) / msPerDay);
}

export function addDays(dateIso: string, days: number): string {
  // UTC setters, not local ones. A bare YYYY-MM-DD parses as UTC midnight, so
  // reading it back with a local getter and writing with a local setter loses
  // a day whenever the clocks go forward: addDays('2026-03-28', 1) returned
  // '2026-03-28' in Beirut, and the equivalent day in every other DST zone.
  const d = new Date(dateIso);
  d.setUTCDate(d.getUTCDate() + days);
  return d.toISOString().slice(0, 10);
}

export function startOfMonth(dateIso: string): string {
  const d = new Date(dateIso);
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth(), 1)).toISOString().slice(0, 10);
}

export function endOfMonth(dateIso: string): string {
  const d = new Date(dateIso);
  return new Date(Date.UTC(d.getUTCFullYear(), d.getUTCMonth() + 1, 0)).toISOString().slice(0, 10);
}

export function isWeekend(dateIso: string): boolean {
  // getUTCDay, to match the UTC midnight a bare YYYY-MM-DD parses to. With the
  // local getter, 2026-09-05 is a Saturday in Beirut and a Friday in New York,
  // so leave day counts and rosters disagreed by a day west of Greenwich.
  const day = new Date(dateIso).getUTCDay();
  return day === 0 || day === 6;
}

export function isToday(dateIso: string): boolean {
  return dateIso === todayIso();
}

/**
 * Whether `dateIso` is after today, where the reader is standing.
 *
 * Compared as strings. `new Date('2026-09-08')` is **UTC midnight**, and
 * comparing that to the current instant answers a different question: west of
 * Greenwich it called tomorrow "not future" for the last hours of the evening,
 * and `isPastDate` called **today** past from 00:00 UTC onward, which at UTC+3
 * is the whole working day (rtylr#1077, the same defect as Flow's).
 *
 * `YYYY-MM-DD` sorts lexicographically in date order, so `<` and `>` are the
 * whole comparison and no parsing is needed.
 */
export function isFutureDate(dateIso: string): boolean {
  return dateIso.split('T')[0] > todayIso();
}

/** Whether `dateIso` is before today. Today itself is not past. */
export function isPastDate(dateIso: string): boolean {
  return dateIso.split('T')[0] < todayIso();
}

export function yearsOld(birthDateIso: string): number {
  const today = new Date();
  const birth = new Date(birthDateIso);
  let years = today.getFullYear() - birth.getFullYear();
  const monthDiff = today.getMonth() - birth.getMonth();
  if (monthDiff < 0 || (monthDiff === 0 && today.getDate() < birth.getDate())) {
    years--;
  }
  return years;
}
