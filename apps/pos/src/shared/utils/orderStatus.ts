// Pure helpers for working with order status values.

import type { OrderStatus } from '../types/pos';

const ORDER_STATUS_LABELS: Record<OrderStatus, string> = {
  pending: 'Pending',
  confirmed: 'Confirmed',
  preparing: 'Preparing',
  open: 'Open',
  sent: 'Submitted',
  ready: 'Ready',
  served: 'Served',
  paid: 'Paid',
  cancelled: 'Cancelled',
  voided: 'Voided',
  refunded: 'Refunded',
  held: 'Held',
};

/** Return a human-readable label for an order status. */
export function orderStatusLabel(status: OrderStatus): string {
  return ORDER_STATUS_LABELS[status] ?? status;
}

/** Return true if the order can still be modified (not yet finalized). */
export function isOrderEditable(status: OrderStatus): boolean {
  return ['pending', 'confirmed', 'preparing', 'ready', 'served', 'held', 'open', 'sent'].includes(status);
}

/** Return true if the order has been fully settled (no further action needed). */
export function isOrderFinalized(status: OrderStatus): boolean {
  return status === 'paid' || status === 'voided' || status === 'refunded' || status === 'cancelled';
}

/** Return true if the order is eligible for a refund. */
export function isRefundable(status: OrderStatus): boolean {
  return status === 'paid';
}

/** Return true if the order is eligible for a void. */
export function isVoidable(status: OrderStatus): boolean {
  return isOrderEditable(status);
}

/** Return the CSS class suffix representing the order status colour. */
export function orderStatusColor(
  status: OrderStatus,
): 'green' | 'yellow' | 'blue' | 'red' | 'gray' {
  switch (status) {
    case 'paid':
      return 'green';
    case 'open':
    case 'pending':
    case 'confirmed':
      return 'blue';
    case 'sent':
    case 'preparing':
    case 'ready':
    case 'served':
    case 'held':
      return 'yellow';
    case 'voided':
    case 'refunded':
    case 'cancelled':
      return 'red';
    default:
      return 'gray';
  }
}
