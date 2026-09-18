// Date helpers used across the finance app.

/**
 * The calendar day a Date falls on where the reader is standing, YYYY-MM-DD.
 *
 * Every other helper in this file reads a Date through local getters, so an
 * ISO day has to be produced the same way. `toISOString().slice(0, 10)` gives
 * the UTC day instead: at UTC+3 that is the previous day for the first three
 * hours of every local day, and for the whole day when the Date was built from
 * local parts.
 */
export function toLocalIso(date: Date): string {
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}-${month}-${day}`;
}

/** Today as YYYY-MM-DD in the reader's own timezone. */
export function todayIso(): string {
  return toLocalIso(new Date());
}

/**
 * Parses a date value the way the rest of this file reads one: locally.
 *
 * `new Date('2024-01-15')` is UTC midnight, but every formatter here uses
 * local getters, so west of Greenwich that renders as 14 January. A bare
 * YYYY-MM-DD is a calendar day with no time and no zone, so it is built from
 * its parts as local midnight. Anything carrying a time is a real instant and
 * is parsed unchanged.
 */
function parseLocal(d: Date | string): Date {
  if (typeof d !== 'string') return d;
  const dateOnly = /^(\d{4})-(\d{2})-(\d{2})$/.exec(d);
  if (!dateOnly) return new Date(d);
  return new Date(Number(dateOnly[1]), Number(dateOnly[2]) - 1, Number(dateOnly[3]));
}

export function formatDate(d: Date | string, opts?: Intl.DateTimeFormatOptions): string {
  const date = parseLocal(d);
  if (isNaN(date.getTime())) return '-';
  return date.toLocaleDateString(
    'en-US',
    opts ?? { month: 'short', day: 'numeric', year: 'numeric' }
  );
}

export function formatDateShort(d: Date | string): string {
  return formatDate(d, { month: 'short', day: 'numeric' });
}

export function daysBetween(a: Date | string, b: Date | string): number {
  // Counted in calendar days, not in 24 hour blocks. A local day is 23 or 25
  // hours long when the clocks change, so dividing raw milliseconds and
  // flooring made a ten day gap measure eleven across a DST transition. The
  // one caller is the late fee calculation, which charged the extra day.
  const d1 = parseLocal(a);
  const d2 = parseLocal(b);
  const start = Date.UTC(d1.getFullYear(), d1.getMonth(), d1.getDate());
  const end = Date.UTC(d2.getFullYear(), d2.getMonth(), d2.getDate());
  return Math.round((end - start) / (1000 * 60 * 60 * 24));
}

export function addDays(d: Date | string, days: number): Date {
  const date = new Date(parseLocal(d));
  date.setDate(date.getDate() + days);
  return date;
}

export function monthKey(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
}

export function monthLabel(d: Date | string): string {
  const date = parseLocal(d);
  return date.toLocaleDateString('en-US', { month: 'short', year: 'numeric' });
}

export function fiscalYearStartFromDate(d: Date, fiscalStartMonth = 1): Date {
  const year = d.getMonth() + 1 >= fiscalStartMonth ? d.getFullYear() : d.getFullYear() - 1;
  return new Date(year, fiscalStartMonth - 1, 1);
}

export function quarterOf(d: Date): number {
  return Math.floor(d.getMonth() / 3) + 1;
}
