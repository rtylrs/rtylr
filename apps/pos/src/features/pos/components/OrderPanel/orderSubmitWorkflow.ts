import type { CreateOrderPayload } from '../../../../shared/services/api/posService';
import type { CartItem } from '../../../../shared/stores/cartStore';

export const ORDER_SUBMIT_ID_ERROR = 'Order was saved, but no server order ID was returned.';

interface BuildOrderSubmitPayloadInput {
  items: CartItem[];
  locationId?: number | null;
  tableId?: number | null;
  customerId?: string | number | null;
  customerName?: string | null;
  orderType?: string;
  notes?: string;
}

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

// Builds a backend CreateOrderRequest-shaped payload. The backend computes
// totals from the per-line unit_price/tax_rate, so no totals are sent.
export function buildOrderSubmitPayload(input: BuildOrderSubmitPayloadInput): CreateOrderPayload {
  return {
    location_id: positiveInteger(input.locationId),
    table_id: positiveInteger(input.tableId),
    customer_id: positiveInteger(input.customerId),
    customer_name: input.customerName ?? undefined,
    type: input.orderType,
    notes: input.notes,
    items: input.items.map((item) => ({
      menu_item_id: item.productId,
      name: item.name,
      quantity: positiveInteger(item.quantity) ?? 1,
      unit_price: Number.isFinite(item.unitPrice) && item.unitPrice >= 0 ? item.unitPrice : 0,
      discount_pct: item.discountPct ?? 0,
      tax_rate: item.taxRate ?? 0,
      notes: item.notes ?? '',
    })),
  };
}

export function extractCreatedOrderId(created: unknown): number {
  const response = created as { id?: unknown; order_id?: unknown; orderId?: unknown };
  const orderId =
    positiveInteger(response.id) ??
    positiveInteger(response.order_id) ??
    positiveInteger(response.orderId);
  if (orderId == null) {
    throw new Error(ORDER_SUBMIT_ID_ERROR);
  }

  return orderId;
}
