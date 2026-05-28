// Pure helpers for working with order status values.

import type { OrderStatus } from '../types/pos';

const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
  open: 'Open',
  sent: 'Sent to Kitchen',
  ready: 'Ready',
  paid: 'Paid',
  voided: 'Voided',
  refunded: 'Refunded',
};

/** Return a human-readable label for an order status. */
export function orderStatusLabel(status: OrderStatus): string {
  return ORDER_STATUS_LABELS[status] ?? status;
}

/** Return true if the order can still be modified (not yet finalized). */
export function isOrderEditable(status: OrderStatus): boolean {
  return status === 'open' || status === 'sent';
}

/** Return true if the order has been fully settled (no further action needed). */
export function isOrderFinalized(status: OrderStatus): boolean {
  return status === 'paid' || status === 'voided' || status === 'refunded';
}

/** Return true if the order is eligible for a refund. */
export function isRefundable(status: OrderStatus): boolean {
  return status === 'paid';
}

/** Return true if the order is eligible for a void. */
export function isVoidable(status: OrderStatus): boolean {
  return status === 'open' || status === 'sent';
}

/** Return the CSS class suffix representing the order status colour. */
export function orderStatusColor(
  status: OrderStatus,
): 'green' | 'yellow' | 'blue' | 'red' | 'gray' {
  switch (status) {
    case 'paid':
      return 'green';
    case 'open':
      return 'blue';
    case 'sent':
    case 'ready':
      return 'yellow';
    case 'voided':
    case 'refunded':
      return 'red';
    default:
      return 'gray';
  }
}
