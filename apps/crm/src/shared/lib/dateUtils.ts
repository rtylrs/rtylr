const DAY_MS = 1000 * 60 * 60 * 24;

/** Parses date-only values in the user's local timezone instead of UTC. */
export function parseCrmDate(value: string): Date {
  const trimmed = value.trim();
  const dateOnly = /^(\d{4})-(\d{2})-(\d{2})$/.exec(trimmed);
  if (!dateOnly) return new Date(trimmed);

  const year = Number(dateOnly[1]);
  const month = Number(dateOnly[2]);
  const day = Number(dateOnly[3]);
  const date = new Date(year, month - 1, day);
  if (date.getFullYear() !== year || date.getMonth() !== month - 1 || date.getDate() !== day) {
    return new Date(Number.NaN);
  }
  return date;
}

function localDateKey(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(
    date.getDate()
  ).padStart(2, '0')}`;
}

function calendarDay(date: Date): number {
  return Date.UTC(date.getFullYear(), date.getMonth(), date.getDate()) / DAY_MS;
}

/** Returns true if the given ISO date string is today in the user's timezone. */
export function isToday(dateStr: string): boolean {
  const date = parseCrmDate(dateStr);
  return Number.isFinite(date.getTime()) && localDateKey(date) === localDateKey(new Date());
}

/** Returns true if the ISO date string falls within the current local month. */
export function isThisMonth(dateStr: string): boolean {
  const now = new Date();
  const date = parseCrmDate(dateStr);
  return (
    Number.isFinite(date.getTime()) &&
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth()
  );
}

/** Returns a human-readable relative label: "Today", "Yesterday", or formatted date. */
export function relativeDate(dateStr: string): string {
  const date = parseCrmDate(dateStr);
  if (!Number.isFinite(date.getTime())) return '';
  const now = new Date();
  const diffDays = calendarDay(now) - calendarDay(date);
  if (diffDays === 0) return 'Today';
  if (diffDays === 1) return 'Yesterday';
  if (diffDays < 0) {
    return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
  }
  if (diffDays < 7) return `${diffDays} days ago`;
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

/** Returns the ISO date string for N days ago. */
export function daysAgo(n: number, from = new Date()): string {
  const d = new Date(from);
  d.setDate(d.getDate() - n);
  return localDateKey(d);
}

/** Returns the ISO date string for N days in the future. */
export function daysFromNow(n: number, from = new Date()): string {
  const d = new Date(from);
  d.setDate(d.getDate() + n);
  return localDateKey(d);
}

/** Returns true if date a is before date b. */
export function isBefore(a: string, b: string): boolean {
  const first = parseCrmDate(a);
  const second = parseCrmDate(b);
  return Number.isFinite(first.getTime()) && Number.isFinite(second.getTime()) && first < second;
}

/** Returns true if date a is after date b. */
export function isAfter(a: string, b: string): boolean {
  const first = parseCrmDate(a);
  const second = parseCrmDate(b);
  return Number.isFinite(first.getTime()) && Number.isFinite(second.getTime()) && first > second;
}
