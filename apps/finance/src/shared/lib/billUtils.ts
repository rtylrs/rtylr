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
  const ref = today ?? new Date().toISOString().slice(0, 10);
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
    const ref = today ?? new Date().toISOString().slice(0, 10);
    if (dueDate < ref) return 'Overdue';
  }
  return 'Unpaid';
}
