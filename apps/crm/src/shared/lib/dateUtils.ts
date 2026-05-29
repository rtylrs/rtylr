/** Returns true if the given ISO date string is today. */
export function isToday(dateStr: string): boolean {
  const today = new Date().toISOString().split('T')[0];
  return dateStr.startsWith(today);
}

/** Returns true if the ISO date string falls within the current month. */
export function isThisMonth(dateStr: string): boolean {
  const now = new Date();
  const d = new Date(dateStr);
  return d.getFullYear() === now.getFullYear() && d.getMonth() === now.getMonth();
}

/** Returns a human-readable relative label: "Today", "Yesterday", or formatted date. */
export function relativeDate(dateStr: string): string {
  const date = new Date(dateStr);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffDays = Math.floor(diffMs / (1000 * 60 * 60 * 24));
  if (diffDays === 0) return 'Today';
  if (diffDays === 1) return 'Yesterday';
  if (diffDays < 7) return `${diffDays} days ago`;
  return date.toLocaleDateString('en-US', { month: 'short', day: 'numeric', year: 'numeric' });
}

/** Returns the ISO date string for N days ago. */
export function daysAgo(n: number, from = new Date()): string {
  const d = new Date(from);
  d.setDate(d.getDate() - n);
  return d.toISOString().split('T')[0];
}

/** Returns the ISO date string for N days in the future. */
export function daysFromNow(n: number, from = new Date()): string {
  const d = new Date(from);
  d.setDate(d.getDate() + n);
  return d.toISOString().split('T')[0];
}

/** Returns true if date a is before date b. */
export function isBefore(a: string, b: string): boolean {
  return new Date(a) < new Date(b);
}

/** Returns true if date a is after date b. */
export function isAfter(a: string, b: string): boolean {
  return new Date(a) > new Date(b);
}
