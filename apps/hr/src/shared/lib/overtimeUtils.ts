/**
 * Pure overtime calculation helpers for the HR module.
 * All functions are side-effect free.
 */

export type OvertimeCategory = 'regular' | 'weekend' | 'holiday' | 'night';
export type OvertimeStatus = 'pending' | 'approved' | 'rejected' | 'cancelled';

export const OT_MULTIPLIERS: Record<OvertimeCategory, number> = {
  regular: 1.5,
  weekend: 2.0,
  holiday: 2.5,
  night: 1.25,
};

export function calculateOvertimePay(
  hours: number,
  hourlyRate: number,
  category: OvertimeCategory
): number {
  const multiplier = OT_MULTIPLIERS[category];
  return Math.round(hours * hourlyRate * multiplier * 100) / 100;
}

export function overtimeCategoryLabel(category: OvertimeCategory): string {
  const labels: Record<OvertimeCategory, string> = {
    regular: 'Regular OT',
    weekend: 'Weekend OT',
    holiday: 'Holiday OT',
    night: 'Night Shift OT',
  };
  return labels[category];
}

export function overtimeStatusLabel(status: OvertimeStatus): string {
  const labels: Record<OvertimeStatus, string> = {
    pending: 'Pending',
    approved: 'Approved',
    rejected: 'Rejected',
    cancelled: 'Cancelled',
  };
  return labels[status];
}

export function isMutable(status: OvertimeStatus): boolean {
  return status === 'pending';
}

export function totalOvertimeHours(records: Array<{ hours: number }>): number {
  return records.reduce((sum, r) => sum + r.hours, 0);
}

export function approvedOvertimeRecords<T extends { status: OvertimeStatus }>(records: T[]): T[] {
  return records.filter((r) => r.status === 'approved');
}
