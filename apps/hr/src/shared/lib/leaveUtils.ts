/**
 * Pure leave / time-off calculation helpers.
 * No external dependencies; all functions take plain values and return plain values.
 */

export type LeaveStatus = 'Pending' | 'Approved' | 'Rejected';

/**
 * Returns the label color token for a leave request status.
 * Values map to Tailwind CSS colour classes used in the UI.
 */
export function leaveStatusColor(status: LeaveStatus): string {
  const map: Record<LeaveStatus, string> = {
    Pending: 'yellow',
    Approved: 'green',
    Rejected: 'red',
  };
  return map[status] ?? 'gray';
}

/**
 * Calculates the number of working days (Mon-Fri) between two ISO date strings
 * inclusive of both start and end dates.
 */
export function workingDaysBetween(startIso: string, endIso: string): number {
  const start = new Date(startIso);
  const end = new Date(endIso);
  let count = 0;
  const current = new Date(start);
  while (current <= end) {
    const dayOfWeek = current.getUTCDay();
    if (dayOfWeek !== 0 && dayOfWeek !== 6) {
      count++;
    }
    current.setUTCDate(current.getUTCDate() + 1);
  }
  return count;
}

/**
 * Returns true when a leave request is approvable (i.e. still Pending).
 */
export function isApprovable(status: LeaveStatus): boolean {
  return status === 'Pending';
}

/**
 * Returns the remaining balance after subtracting pending and approved days.
 * Never returns a negative number.
 */
export function remainingBalance(totalDays: number, usedDays: number, pendingDays: number): number {
  return Math.max(0, totalDays - usedDays - pendingDays);
}

/**
 * Returns a human-friendly label for the leave duration.
 * e.g. 1 -> "1 day", 5 -> "5 days"
 */
export function formatLeaveDuration(days: number): string {
  if (days <= 0) return '0 days';
  return `${days} day${days === 1 ? '' : 's'}`;
}

/**
 * Checks whether a leave request period overlaps with an existing period.
 * Returns true if there is any overlap.
 */
export function hasOverlap(
  newStart: string,
  newEnd: string,
  existingStart: string,
  existingEnd: string
): boolean {
  const ns = new Date(newStart).getTime();
  const ne = new Date(newEnd).getTime();
  const es = new Date(existingStart).getTime();
  const ee = new Date(existingEnd).getTime();
  return ns <= ee && ne >= es;
}
