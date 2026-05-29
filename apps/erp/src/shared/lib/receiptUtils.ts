/**
 * Purchase order receipt and receiving helpers for the ERP app.
 * All functions are pure with no side-effects.
 */

import type { PurchaseOrder } from '../services/erpApi';

export type ReceiveStatus = 'not-started' | 'partial' | 'complete' | 'over-received';

export interface ReceiveLine {
  productId: number;
  sku: string;
  name: string;
  ordered: number;
  received: number;
  unit?: string;
}

export interface ReceiveSummary {
  totalOrdered: number;
  totalReceived: number;
  totalPending: number;
  status: ReceiveStatus;
  completionPct: number;
}

/** Compute the receive status for a single line item. */
export function lineReceiveStatus(line: ReceiveLine): ReceiveStatus {
  if (line.received === 0) return 'not-started';
  if (line.received > line.ordered) return 'over-received';
  if (line.received === line.ordered) return 'complete';
  return 'partial';
}

/** Compute an overall receiving summary for a list of lines. */
export function receiveSummary(lines: ReceiveLine[]): ReceiveSummary {
  const totalOrdered = lines.reduce((s, l) => s + l.ordered, 0);
  const totalReceived = lines.reduce((s, l) => s + l.received, 0);
  const totalPending = Math.max(0, totalOrdered - totalReceived);
  const completionPct = totalOrdered === 0 ? 0 : Math.round((totalReceived / totalOrdered) * 100);

  let status: ReceiveStatus;
  if (totalReceived === 0) {
    status = 'not-started';
  } else if (totalReceived > totalOrdered) {
    status = 'over-received';
  } else if (totalReceived === totalOrdered) {
    status = 'complete';
  } else {
    status = 'partial';
  }

  return { totalOrdered, totalReceived, totalPending, status, completionPct };
}

/** Return lines that have not been fully received. */
export function pendingLines(lines: ReceiveLine[]): ReceiveLine[] {
  return lines.filter((l) => l.received < l.ordered);
}

/** Return lines that have been over-received. */
export function overReceivedLines(lines: ReceiveLine[]): ReceiveLine[] {
  return lines.filter((l) => l.received > l.ordered);
}

/** Check if a purchase order can be marked as received. */
export function canMarkReceived(order: PurchaseOrder): boolean {
  return order.status === 'Sent' || order.status === 'Approved';
}

/** Check if a purchase order can be sent to the supplier. */
export function canSendOrder(order: PurchaseOrder): boolean {
  return order.status === 'Draft' || order.status === 'Pending Approval';
}

/** Return a human-readable label for a PO status. */
export function poStatusLabel(status: PurchaseOrder['status']): string {
  switch (status) {
    case 'Draft':
      return 'Draft';
    case 'Sent':
      return 'Sent to Supplier';
    case 'Received':
      return 'Received';
    case 'Pending Approval':
      return 'Awaiting Approval';
    case 'Approved':
      return 'Approved';
  }
}
