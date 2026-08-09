/**
 * Pure shift / scheduling utility helpers for the HR scheduling module.
 * No external dependencies; all functions operate on plain values.
 */

export type ShiftStatus = 'Scheduled' | 'In Progress' | 'Completed';

/**
 * Parses a "HH:MM" time string into total minutes since midnight.
 */
export function timeToMinutes(time: string): number {
  const [hourStr, minuteStr] = time.split(':');
  return parseInt(hourStr, 10) * 60 + parseInt(minuteStr ?? '0', 10);
}

/**
 * Calculates the duration of a shift in minutes.
 * Handles overnight shifts where end time is earlier than start time.
 */
export function shiftDurationMinutes(start: string, end: string): number {
  const startMin = timeToMinutes(start);
  const endMin = timeToMinutes(end);
  if (endMin >= startMin) {
    return endMin - startMin;
  }
  return 24 * 60 - startMin + endMin;
}

/**
 * Returns true if a shift spans midnight (overnight shift).
 */
export function isOvernightShift(start: string, end: string): boolean {
  return timeToMinutes(end) < timeToMinutes(start);
}

/**
 * Returns true if two shifts overlap in time on the same day.
 */
export function shiftsOverlap(start1: string, end1: string, start2: string, end2: string): boolean {
  const s1 = timeToMinutes(start1);
  const e1 = timeToMinutes(end1);
  const s2 = timeToMinutes(start2);
  const e2 = timeToMinutes(end2);
  return s1 < e2 && e1 > s2;
}

/**
 * Returns the human-readable badge label for a shift status.
 */
export function shiftStatusLabel(status: ShiftStatus): string {
  const labels: Record<ShiftStatus, string> = {
    Scheduled: 'Scheduled',
    'In Progress': 'In Progress',
    Completed: 'Completed',
  };
  return labels[status] ?? status;
}

/**
 * Returns true if a shift is currently active (In Progress).
 */
export function isShiftActive(status: ShiftStatus): boolean {
  return status === 'In Progress';
}

/**
 * Formats a shift time range as a display string.
 * e.g. "09:00" and "17:00" -> "09:00 - 17:00"
 */
export function formatShiftRange(start: string, end: string): string {
  return `${start} - ${end}`;
}
