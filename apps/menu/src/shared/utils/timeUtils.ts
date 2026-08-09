/**
 * Time and schedule utilities for the Menu app.
 */

export interface TimeRange {
  open: number;
  close: number;
}

/** Return true if the given hour is within the time range (open inclusive, close exclusive). */
export function isWithinHours(hour: number, range: TimeRange): boolean {
  return hour >= range.open && hour < range.close;
}

/** Return true if the restaurant is currently open based on the current local hour. */
export function isOpenNow(range: TimeRange): boolean {
  return isWithinHours(new Date().getHours(), range);
}

/** Format a 24h hour as a 12h display string (e.g. 13 -> "1:00 PM"). */
export function formatHour(hour: number): string {
  const period = hour < 12 ? 'AM' : 'PM';
  const h = hour % 12 || 12;
  return `${h}:00 ${period}`;
}

/** Return the number of minutes until the given target hour (may be 0 if already past). */
export function minutesUntilHour(targetHour: number): number {
  const now = new Date();
  const nowMinutes = now.getHours() * 60 + now.getMinutes();
  const targetMinutes = targetHour * 60;
  const diff = targetMinutes - nowMinutes;
  return Math.max(0, diff);
}

/** Format a duration in minutes as a human-readable string (e.g. 90 -> "1h 30m"). */
export function formatDuration(minutes: number): string {
  if (minutes < 60) return `${minutes}m`;
  const h = Math.floor(minutes / 60);
  const m = minutes % 60;
  return m === 0 ? `${h}h` : `${h}h ${m}m`;
}

/** Return an estimated wait-time label based on minutes. */
export function waitLabel(minutes: number): string {
  if (minutes <= 0) return 'Ready now';
  if (minutes < 60) return `${minutes} min`;
  return `About ${Math.ceil(minutes / 60)}h`;
}
