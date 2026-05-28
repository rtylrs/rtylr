/**
 * Order status and calculation utilities for the Dash app.
 */

export type OrderStatus = 'pending' | 'processing' | 'completed' | 'cancelled' | 'refunded';

export interface OrderItem {
  price: number;
  quantity: number;
  discount?: number;
}

/** Compute the subtotal for a list of order items. */
export function orderSubtotal(items: OrderItem[]): number {
  return items.reduce((sum, item) => {
    const lineDiscount = item.discount ?? 0;
    return sum + item.price * item.quantity - lineDiscount;
  }, 0);
}

/** Compute the total including tax. */
export function orderTotal(subtotal: number, taxRate: number): number {
  return subtotal * (1 + taxRate / 100);
}

/** Return whether an order status is a terminal state. */
export function isTerminalStatus(status: OrderStatus): boolean {
  return status === 'completed' || status === 'cancelled' || status === 'refunded';
}

/** Return whether an order status is active (in-progress). */
export function isActiveStatus(status: OrderStatus): boolean {
  return status === 'pending' || status === 'processing';
}

/** Return a display label for an order status. */
export function statusLabel(status: OrderStatus): string {
  const labels: Record<OrderStatus, string> = {
    pending: 'Pending',
    processing: 'Processing',
    completed: 'Completed',
    cancelled: 'Cancelled',
    refunded: 'Refunded',
  };
  return labels[status] ?? status;
}

/** Sort orders by date, newest first. */
export function sortOrdersByDate<T extends { createdAt: string }>(orders: T[]): T[] {
  return [...orders].sort((a, b) => b.createdAt.localeCompare(a.createdAt));
}

/** Filter orders by status. */
export function filterByStatus<T extends { status: OrderStatus }>(
  orders: T[],
  status: OrderStatus
): T[] {
  return orders.filter((o) => o.status === status);
}
