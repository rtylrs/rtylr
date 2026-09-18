// Pure helpers for receipt data assembly and validation.

import type { Receipt, PaymentEntry } from '../types/pos';

/** Return true if the receipt has been emailed to the customer. */
export function isEmailed(receipt: Receipt): boolean {
  return receipt.emailed === true;
}

/** Return true if the receipt contains at least one item. */
export function hasItems(receipt: Receipt): boolean {
  return receipt.items.length > 0;
}

/** Return the total number of items (summing qty) across all receipt lines. */
export function receiptItemCount(receipt: Receipt): number {
  return receipt.items.reduce((sum, i) => sum + i.qty, 0);
}

/** Return true if the receipt total matches the sum of payments within tolerance. */
export function isBalanced(receipt: Receipt, tolerance = 0.01): boolean {
  const paid = receipt.payments.reduce(
    (sum, p: PaymentEntry) => sum + (Number.isFinite(p.amount) ? p.amount : 0),
    0,
  );
  return Math.abs(paid - receipt.total) <= tolerance;
}

/** Return the payment method labels used in a receipt. */
export function paymentMethodLabels(receipt: Receipt): string[] {
  const seen = new Set<string>();
  const labels: string[] = [];
  for (const p of receipt.payments) {
    if (!seen.has(p.method)) {
      seen.add(p.method);
      labels.push(p.method);
    }
  }
  return labels;
}

/** Return a greeting line for the receipt based on optional customer name. */
export function receiptGreeting(customerName?: string): string {
  if (!customerName || !customerName.trim()) return 'Thank you!';
  return `Thank you, ${customerName.trim()}!`;
}

/** Return a formatted receipt date string from the printedAt ISO timestamp. */
export function receiptDateLabel(receipt: Receipt): string {
  const date = new Date(receipt.printedAt);
  if (!Number.isFinite(date.getTime())) return '';
  return date.toLocaleDateString();
}
