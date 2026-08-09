/**
 * Pure attendance utility helpers for the HR attendance module.
 * No external dependencies; all functions operate on plain values.
 */

export type AttendanceStatus = 'Present' | 'Absent' | 'Late' | 'Half Day' | 'Remote';

/**
 * Returns the display label for an attendance status badge.
 */
export function attendanceStatusLabel(status: AttendanceStatus): string {
  const labels: Record<AttendanceStatus, string> = {
    Present: 'Present',
    Absent: 'Absent',
    Late: 'Late',
    'Half Day': 'Half Day',
    Remote: 'Remote',
  };
  return labels[status] ?? status;
}

/**
 * Returns the color token for an attendance status badge.
 */
export function attendanceStatusColor(status: AttendanceStatus): string {
  const colors: Record<AttendanceStatus, string> = {
    Present: 'green',
    Absent: 'red',
    Late: 'yellow',
    'Half Day': 'orange',
    Remote: 'blue',
  };
  return colors[status] ?? 'gray';
}

/**
 * Returns true when attendance is considered billable/worked.
 */
export function isWorkedDay(status: AttendanceStatus): boolean {
  return status === 'Present' || status === 'Remote' || status === 'Late' || status === 'Half Day';
}

/**
 * Returns the fraction of a day counted as worked for a given status.
 * Half Day = 0.5, Absent = 0, all others = 1.
 */
export function workedDayFraction(status: AttendanceStatus): number {
  if (status === 'Half Day') return 0.5;
  if (status === 'Absent') return 0;
  return 1;
}

/**
 * Calculates the attendance rate as a percentage given present and total days.
 * Returns 0 when totalDays is 0.
 */
export function attendanceRate(presentDays: number, totalDays: number): number {
  if (totalDays === 0) return 0;
  return Math.round((presentDays / totalDays) * 100 * 10) / 10;
}

/**
 * Calculates total effective working days from an array of attendance records.
 * Each record contributes according to its workedDayFraction.
 */
export function totalEffectiveDays(records: Array<{ status: AttendanceStatus }>): number {
  return records.reduce((sum, r) => sum + workedDayFraction(r.status), 0);
}

/**
 * Returns true when the employee was late on the given record.
 */
export function isLate(status: AttendanceStatus): boolean {
  return status === 'Late';
}

/**
 * Counts the number of absences in an array of attendance records.
 */
export function countAbsences(records: Array<{ status: AttendanceStatus }>): number {
  return records.filter((r) => r.status === 'Absent').length;
}

/**
 * Determines whether an attendance record is within a given date range (inclusive).
 */
export function isInRange(dateIso: string, fromIso: string, toIso: string): boolean {
  return dateIso >= fromIso && dateIso <= toIso;
}
