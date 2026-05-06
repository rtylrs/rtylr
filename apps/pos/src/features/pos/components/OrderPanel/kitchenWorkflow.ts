import type { OrderPayload } from '../../../../shared/services/posService';
import type { CartItem } from '../../../../shared/stores/cartStore';

export const KITCHEN_ORDER_ID_ERROR = 'Order must sync before it can be sent to kitchen.';

interface BuildKitchenOrderPayloadInput {
  items: CartItem[];
  subtotal: number;
  tax: number;
  total: number;
  employeeId?: number | string;
  now?: () => string;
  localId?: () => string;
}

function nonNegativeNumber(value: number): number {
  if (!Number.isFinite(value) || value < 0) return 0;
  return value;
}

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

export function buildKitchenOrderPayload(input: BuildKitchenOrderPayloadInput): OrderPayload {
  return {
    id: input.localId?.() ?? `local-${Date.now()}`,
    status: 'open',
    items: input.items.map((item) => ({
      id: item.id,
      name: item.name,
      price: nonNegativeNumber(item.price),
      quantity: positiveInteger(item.quantity) ?? 1,
      notes: item.notes ?? '',
    })),
    subtotal: nonNegativeNumber(input.subtotal),
    tax: nonNegativeNumber(input.tax),
    tip: 0,
    discount: 0,
    total: nonNegativeNumber(input.total),
    createdAt: input.now?.() ?? new Date().toISOString(),
    customerId: String(input.employeeId ?? ''),
  };
}

export function extractCreatedOrderId(created: unknown): number {
  const response = created as { id?: unknown; order_id?: unknown; orderId?: unknown };
  const orderId =
    positiveInteger(response.id) ??
    positiveInteger(response.order_id) ??
    positiveInteger(response.orderId);
  if (orderId == null) {
    throw new Error(KITCHEN_ORDER_ID_ERROR);
  }

  return orderId;
}
