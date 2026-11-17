/**
 * The calendar day a Date falls on where the reader is standing, YYYY-MM-DD.
 *
 * Local, not UTC. `toISOString()` gives the UTC day, which at UTC+3 is the
 * previous day for the first three hours of every local day, and the whole day
 * for a Date built from local parts.
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

const RELATIVE_UNITS: ReadonlyArray<[Intl.RelativeTimeFormatUnit, number]> = [
  ['year', 31_536_000],
  ['month', 2_592_000],
  ['week', 604_800],
  ['day', 86_400],
  ['hour', 3_600],
  ['minute', 60],
];

/** "3 minutes ago" in the reader's locale, or '' when iso does not parse. */
export function formatRelativeTime(iso: string, locale: string, now = Date.now()): string {
  const then = Date.parse(iso);
  if (Number.isNaN(then)) return '';
  const seconds = Math.round((then - now) / 1000);
  const format = new Intl.RelativeTimeFormat(locale, { numeric: 'auto' });
  for (const [unit, size] of RELATIVE_UNITS) {
    if (Math.abs(seconds) >= size) return format.format(Math.round(seconds / size), unit);
  }
  return format.format(0, 'minute');
}
