export type OrderStatusValue = 'received' | 'preparing' | 'ready' | 'done';

export interface StatusStep {
  key: OrderStatusValue;
  label: string;
}

export const ORDER_STATUS_STEPS: StatusStep[] = [
  { key: 'received', label: 'Received' },
  { key: 'preparing', label: 'Preparing' },
  { key: 'ready', label: 'Ready' },
  { key: 'done', label: 'Done' },
];

/**
 * Returns the zero-based index of a status within the ORDER_STATUS_STEPS array.
 * Returns -1 when the status is not found.
 */
export function statusIndex(status: OrderStatusValue): number {
  return ORDER_STATUS_STEPS.findIndex((step) => step.key === status);
}

/**
 * Returns the progress percentage (0-100) for a given status.
 */
export function statusProgress(status: OrderStatusValue): number {
  const idx = statusIndex(status);
  if (idx < 0) return 0;
  return ((idx + 1) / ORDER_STATUS_STEPS.length) * 100;
}

/**
 * Returns true when the order has reached the "done" state.
 */
export function isOrderDone(status: OrderStatusValue): boolean {
  return status === 'done';
}

/**
 * Returns a human-readable status message for the given order status.
 */
export function statusMessage(status: OrderStatusValue): string {
  switch (status) {
    case 'received':
      return 'Order received!';
    case 'preparing':
      return 'Being prepared...';
    case 'ready':
      return 'Ready for pickup!';
    case 'done':
      return 'Order complete!';
  }
}
