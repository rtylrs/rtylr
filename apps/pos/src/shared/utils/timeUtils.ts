// Pure helpers for time-of-day checks and business hour logic.

/** Return the current hour (0-23) in local time. */
export function currentHour(): number {
  return new Date().getHours();
}

/** Return true if the given hour is within a business-hour window [open, close). */
export function isBusinessHour(hour: number, openHour: number, closeHour: number): boolean {
  if (!Number.isFinite(hour) || !Number.isFinite(openHour) || !Number.isFinite(closeHour)) {
    return false;
  }
  if (closeHour > openHour) {
    return hour >= openHour && hour < closeHour;
  }
  // Handles overnight hours (e.g., open at 22, close at 2)
  return hour >= openHour || hour < closeHour;
}

/** Return the part of day based on hour. */
export function partOfDay(hour: number): 'morning' | 'afternoon' | 'evening' | 'night' {
  if (!Number.isFinite(hour)) return 'night';
  if (hour >= 5 && hour < 12) return 'morning';
  if (hour >= 12 && hour < 17) return 'afternoon';
  if (hour >= 17 && hour < 21) return 'evening';
  return 'night';
}

/** Format seconds as "MM:SS" for display. */
export function formatMmSs(totalSeconds: number): string {
  if (!Number.isFinite(totalSeconds) || totalSeconds < 0) return '00:00';
  const s = Math.floor(totalSeconds);
  const minutes = Math.floor(s / 60);
  const seconds = s % 60;
  return `${String(minutes).padStart(2, '0')}:${String(seconds).padStart(2, '0')}`;
}

/** Return the number of full minutes in a given seconds count. */
export function secondsToMinutes(seconds: number): number {
  if (!Number.isFinite(seconds)) return 0;
  return Math.floor(seconds / 60);
}

/** Return true if a given ISO timestamp is today (in local time). */
export function isToday(iso: string): boolean {
  const date = new Date(iso);
  if (!Number.isFinite(date.getTime())) return false;
  const now = new Date();
  return (
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth() &&
    date.getDate() === now.getDate()
  );
}
