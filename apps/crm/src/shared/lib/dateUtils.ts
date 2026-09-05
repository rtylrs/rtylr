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

/**
 * Parses a value the way the helpers here read one: locally.
 *
 * A bare YYYY-MM-DD is a calendar day with no time and no zone, but
 * `new Date('2026-01-01')` is UTC midnight, so a local getter reads it as
 * 31 December west of Greenwich. Values carrying a time are real instants and
 * are parsed unchanged.
 */
function parseLocal(dateStr: string): Date {
  const dateOnly = /^(\d{4})-(\d{2})-(\d{2})$/.exec(dateStr);
  if (!dateOnly) return new Date(dateStr);
  return new Date(Number(dateOnly[1]), Number(dateOnly[2]) - 1, Number(dateOnly[3]));
}

/** Returns true if the given ISO date string is today. */
export function isToday(dateStr: string): boolean {
  return dateStr.startsWith(todayIso());
}

/** Returns true if the ISO date string falls within the current month. */
export function isThisMonth(dateStr: string): boolean {
  const now = new Date();
  const d = parseLocal(dateStr);
  return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth();
}

/** Returns a human-readable relative label: "Today", "Yesterday", or formatted date. */
export function relativeDate(dateStr: string): string {
  // Compared as calendar days, not as elapsed hours. Dividing the raw gap
  // labelled a record from 23:00 yesterday "Today" when read at 01:00, and a
  // local day is 23 or 25 hours long when the clocks change.
  const date = parseLocal(dateStr);
  const now = new Date();
  const startOfDate = Date.UTC(date.getFullYear(), date.getMonth(), date.getDate());
  const startOfNow = Date.UTC(now.getFullYear(), now.getMonth(), now.getDate());
  const diffDays = Math.round((startOfNow - startOfDate) / (1000 * 60 * 60 * 24));
  if (diffDays === 0) return 'Today';
  if (diffDays === 1) return 'Yesterday';
  if (diffDays < 7) return `${diffDays} days ago`;
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

/** Returns the ISO date string for N days ago. */
export function daysAgo(n: number, from = new Date()): string {
  const d = new Date(from);
  d.setDate(d.getDate() - n);
  return toLocalIso(d);
}

/** Returns the ISO date string for N days in the future. */
export function daysFromNow(n: number, from = new Date()): string {
  const d = new Date(from);
  d.setDate(d.getDate() + n);
  return toLocalIso(d);
}

/** Returns true if date a is before date b. */
export function isBefore(a: string, b: string): boolean {
  return new Date(a) < new Date(b);
}

/** Returns true if date a is after date b. */
export function isAfter(a: string, b: string): boolean {
  return new Date(a) > new Date(b);
}
