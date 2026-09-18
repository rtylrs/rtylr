// Pure helpers for date manipulation used in POS sessions and reports.

/** Returns the current date as a YYYY-MM-DD string in local time. */
export function todayISO(): string {
  const d = new Date();
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

/** Returns true if the given ISO string represents today. */
export function isToday(iso: string): boolean {
  try {
    const d = new Date(iso);
    const today = new Date();
    return (
      d.getFullYear() === today.getFullYear() &&
      d.getMonth() === today.getMonth() &&
      d.getDate() === today.getDate()
    );
  } catch {
    return false;
  }
}

/** Returns the start of day (00:00:00.000) for a given date. */
export function startOfDay(date: Date): Date {
  const d = new Date(date);
  d.setHours(0, 0, 0, 0);
  return d;
}

/** Returns the end of day (23:59:59.999) for a given date. */
export function endOfDay(date: Date): Date {
  const d = new Date(date);
  d.setHours(23, 59, 59, 999);
  return d;
}

/** Returns the number of full minutes between two dates (from - to). */
export function minutesBetween(from: Date, to: Date): number {
  return Math.floor((to.getTime() - from.getTime()) / 60_000);
}

/** Returns the number of full seconds between two dates (from - to). */
export function secondsBetween(from: Date, to: Date): number {
  return Math.floor((to.getTime() - from.getTime()) / 1_000);
}

/** Format a duration in seconds as a human-readable string, e.g. "2 min 30 s". */
export function formatDuration(totalSeconds: number): string {
  if (!Number.isFinite(totalSeconds) || totalSeconds < 0) return '0 s';
  const minutes = Math.floor(totalSeconds / 60);
  const seconds = totalSeconds % 60;
  if (minutes === 0) return `${seconds} s`;
  if (seconds === 0) return `${minutes} min`;
  return `${minutes} min ${seconds} s`;
}
