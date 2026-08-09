/**
 * Pure timesheet utility helpers for the HR timesheet module.
 * All functions are side-effect free.
 */

export type TimesheetStatus = 'draft' | 'submitted' | 'approved' | 'rejected';

export interface TimeEntry {
  date: string;
  startTime: string;
  endTime: string;
  breakMinutes: number;
}

export function parseTimeToMinutes(time: string): number {
  const [hours, minutes] = time.split(':').map(Number);
  return hours * 60 + minutes;
}

export function netWorkingMinutes(entry: TimeEntry): number {
  const start = parseTimeToMinutes(entry.startTime);
  const end = parseTimeToMinutes(entry.endTime);
  const raw = end - start;
  return Math.max(0, raw - entry.breakMinutes);
}

export function minutesToDecimalHours(minutes: number): number {
  return Math.round((minutes / 60) * 100) / 100;
}

export function totalHoursForEntries(entries: TimeEntry[]): number {
  const totalMinutes = entries.reduce((sum, e) => sum + netWorkingMinutes(e), 0);
  return minutesToDecimalHours(totalMinutes);
}

export function isEditable(status: TimesheetStatus): boolean {
  return status === 'draft' || status === 'rejected';
}

export function isPendingReview(status: TimesheetStatus): boolean {
  return status === 'submitted';
}

export function timesheetStatusLabel(status: TimesheetStatus): string {
  const labels: Record<TimesheetStatus, string> = {
    draft: 'Draft',
    submitted: 'Pending Review',
    approved: 'Approved',
    rejected: 'Rejected',
  };
  return labels[status];
}

export function groupEntriesByDate(entries: TimeEntry[]): Record<string, TimeEntry[]> {
  return entries.reduce<Record<string, TimeEntry[]>>((acc, entry) => {
    if (!acc[entry.date]) acc[entry.date] = [];
    acc[entry.date].push(entry);
    return acc;
  }, {});
}
