import { todayIso } from './date';
/**
 * Pure utility helpers for bill state and display logic.
 */

export type BillStatus = 'Unpaid' | 'Paid' | 'Overdue' | 'Partial';

/**
 * Returns true when the bill is overdue based on its due date.
 * Compares against an optional reference date (defaults to today).
 */
export function isBillOverdue(
  dueDate: string | null | undefined,
  status: BillStatus,
  today?: string
): boolean {
  if (!dueDate) return false;
  if (status === 'Paid') return false;
  const ref = today ?? todayIso();
  return dueDate < ref;
}

/**
 * Returns the amount still owed on a bill.
 */
export function getBillBalance(amount: number | null | undefined, paid = 0): number {
  const total = amount ?? 0;
  return Math.max(0, total - paid);
}

/**
 * Returns true when the bill has been fully paid.
 */
export function isBillPaid(amount: number | null | undefined, paid = 0): boolean {
  return getBillBalance(amount, paid) === 0;
}

/**
 * Returns a human-readable label for a bill status.
 */
export function getBillStatusLabel(status: BillStatus): string {
  const labels: Record<BillStatus, string> = {
    Unpaid: 'Unpaid',
    Paid: 'Paid',
    Overdue: 'Overdue',
    Partial: 'Partially Paid',
  };
  return labels[status] ?? status;
}

/**
 * Returns the number of days a bill is overdue relative to a reference date.
 * Returns 0 when the bill is not overdue or dueDate is null.
 */
export function daysOverdue(dueDate: string | null | undefined, today?: string): number {
  if (!dueDate) return 0;
  const ref = today ?? todayIso();
  if (dueDate >= ref) return 0;
  const due = new Date(dueDate).getTime();
  const now = new Date(ref).getTime();
  return Math.floor((now - due) / (1000 * 60 * 60 * 24));
}

/**
 * Returns an urgency tier for a bill based on how many days overdue it is.
 * "low" = not overdue, "medium" = 1-7 days, "high" = 8-30 days, "critical" = over 30 days.
 */
export function getBillUrgency(
  dueDate: string | null | undefined,
  today?: string
): 'low' | 'medium' | 'high' | 'critical' {
  const days = daysOverdue(dueDate, today);
  if (days === 0) return 'low';
  if (days <= 7) return 'medium';
  if (days <= 30) return 'high';
  return 'critical';
}

/**
 * Derives the effective bill status based on amount paid and due date.
 */
export function deriveBillStatus(
  amount: number,
  paid: number,
  dueDate: string | null | undefined,
  today?: string
): BillStatus {
  const balance = getBillBalance(amount, paid);
  if (balance === 0) return 'Paid';
  if (paid > 0) return 'Partial';
  if (dueDate) {
    const ref = today ?? todayIso();
    if (dueDate < ref) return 'Overdue';
  }
  return 'Unpaid';
}
