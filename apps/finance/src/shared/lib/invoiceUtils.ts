import { todayIso } from './date';
/**
 * Pure utility helpers for invoice calculations and display logic.
 */

export type InvoiceStatus = 'Draft' | 'Sent' | 'Paid' | 'Overdue';

export interface InvoiceSummary {
  amount: number;
  paid?: number | null;
  dueDate?: string;
  status?: InvoiceStatus;
}

/**
 * Returns the outstanding (unpaid) amount for an invoice.
 */
export function getOutstandingAmount(invoice: InvoiceSummary): number {
  const paid = invoice.paid ?? 0;
  return Math.max(0, invoice.amount - paid);
}

/**
 * Returns true if the invoice has been fully paid.
 */
export function isFullyPaid(invoice: InvoiceSummary): boolean {
  return getOutstandingAmount(invoice) === 0;
}

/**
 * Returns true if the invoice is overdue based on its due date.
 * Compares against an optional reference date (defaults to today).
 */
export function isOverdue(invoice: InvoiceSummary, today?: string): boolean {
  if (!invoice.dueDate) return false;
  if (invoice.status === 'Paid') return false;
  const ref = today ?? todayIso();
  return invoice.dueDate < ref;
}

export interface InvoiceLineItem {
  quantity?: number | null;
  unitPrice?: number | null;
  amount?: number | null;
}

export interface InvoiceTotals {
  subtotal: number;
  discount: number;
  tax: number;
  total: number;
}

/**
 * Computes subtotal, discount, tax, and total from invoice line items.
 */
export function computeInvoiceTotals(
  items: InvoiceLineItem[],
  discountAmount = 0,
  taxRate = 0
): InvoiceTotals {
  const subtotal = items.reduce((sum, item) => {
    if (item.amount != null) return sum + item.amount;
    const qty = item.quantity ?? 0;
    const price = item.unitPrice ?? 0;
    return sum + qty * price;
  }, 0);

  const discount = Math.min(discountAmount, subtotal);
  const taxable = subtotal - discount;
  const tax = taxRate > 0 ? taxable * (taxRate / 100) : 0;
  const total = taxable + tax;

  return { subtotal, discount, tax, total };
}

/**
 * Returns a human-readable status label for display.
 */
export function getStatusLabel(status: InvoiceStatus): string {
  const labels: Record<InvoiceStatus, string> = {
    Draft: 'Draft',
    Sent: 'Sent',
    Paid: 'Paid',
    Overdue: 'Overdue',
  };
  return labels[status] ?? status;
}

/**
 * Returns a CSS color class name for an invoice status badge.
 */
export function getStatusColor(status: InvoiceStatus): string {
  const colors: Record<InvoiceStatus, string> = {
    Draft: 'gray',
    Sent: 'blue',
    Paid: 'green',
    Overdue: 'red',
  };
  return colors[status] ?? 'gray';
}
