import type { CreateOrderRequest, CreateOrderResponse } from '@voxire/voxie/pos/orders/v1/types';
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

function idString(value: unknown): string {
  return String(positiveInteger(value) ?? 0);
}

// Builds the generated Voxie CreateOrderRequest. The backend computes totals
// from the per-line unitPrice/taxRate, so no totals are sent.
export function buildOrderSubmitPayload(input: BuildOrderSubmitPayloadInput): CreateOrderRequest {
  return {
    sessionId: '0',
    locationId: idString(input.locationId),
    tableId: idString(input.tableId),
    customerId: idString(input.customerId),
    customerName: input.customerName ?? '',
    type: input.orderType ?? '',
    paymentMethod: '',
    notes: input.notes ?? '',
    kitchenNotes: '',
    items: input.items.map((item) => ({
      menuItemId: String(item.productId),
      name: item.name,
      sku: item.sku ?? '',
      quantity: positiveInteger(item.quantity) ?? 1,
      unitPrice: Number.isFinite(item.unitPrice) && item.unitPrice >= 0 ? item.unitPrice : 0,
      discountPct: item.discountPct ?? 0,
      taxRate: item.taxRate ?? 0,
      notes: item.notes ?? '',
    })),
  };
}

export function extractCreatedOrderId(created: CreateOrderResponse): number {
  const orderId = positiveInteger(created.orderId);
  if (orderId == null) {
    throw new Error(ORDER_SUBMIT_ID_ERROR);
  }

  return orderId;
}
