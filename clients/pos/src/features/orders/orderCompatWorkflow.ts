import { createRestRouteClient, unwrap } from "@voxire/api";
import type { CreateOrderResponse } from "@voxire/voxie/pos/orders/v1/types";
import type { CartItem, CartProductInput } from "../../shared/stores/cartStore";

export interface OrderCompatItem {
  menuItemId?: number;
  comboId?: number;
  modifierIds?: number[];
  name: string;
  sku: string;
  quantity: number;
  unitPrice: number;
  discountPct: number;
  discountAmount: number;
  taxRate: number;
  notes: string;
}

export interface OrderCompatPayload {
  sessionId: number;
  tableId: number;
  locationId: number;
  type: string;
  customerId: number;
  customerName: string;
  paymentMethod: string;
  notes: string;
  kitchenNotes: string;
  items: OrderCompatItem[];
}

export interface OrderCompatResponse {
  orderId?: number | string;
  orderNumber?: string;
  message?: string;
  success?: boolean;
}

// The generated order contract still lacks combo, modifier, and flat-discount
// fields. Keep the fallback in this adapter until Voxie exposes them.
const restClient = createRestRouteClient("pos");

function positiveInteger(value: unknown): number | undefined {
  const numeric =
    typeof value === "string" ? Number(value.trim()) : Number(value);
  return Number.isInteger(numeric) && numeric > 0 ? numeric : undefined;
}

function idNumber(value: unknown): number {
  return positiveInteger(value) ?? 0;
}

function finiteNonNegative(value: unknown): number {
  const numeric = Number(value);
  return Number.isFinite(numeric) && numeric >= 0 ? numeric : 0;
}

function compatItem(item: CartProductInput | CartItem): OrderCompatItem {
  const comboId = positiveInteger(item.comboId);
  return {
    ...(comboId ? { comboId } : { menuItemId: idNumber(item.productId) }),
    modifierIds: (item.modifiers ?? []).map((modifier) =>
      idNumber(modifier.id),
    ),
    name: item.name,
    sku: item.sku ?? "",
    quantity: positiveInteger(item.quantity) ?? 1,
    unitPrice: finiteNonNegative(item.unitPrice),
    discountPct: Math.min(100, finiteNonNegative(item.discountPct)),
    discountAmount: finiteNonNegative(item.discountAmount),
    taxRate: finiteNonNegative(item.taxRate),
    notes: item.notes ?? "",
  };
}

export function hasComboItems(
  items: Array<CartProductInput | CartItem>,
): boolean {
  return items.some((item) => positiveInteger(item.comboId) != null);
}

export function requiresCompatOrder(
  items: Array<CartProductInput | CartItem>,
): boolean {
  return items.some(
    (item) =>
      hasComboItems([item]) ||
      (item.modifiers?.length ?? 0) > 0 ||
      finiteNonNegative(item.discountAmount) > 0,
  );
}

function collapseProducts(
  items: Array<CartProductInput | CartItem>,
): Array<CartProductInput | CartItem> {
  const collapsed = new Map<string, CartProductInput | CartItem>();
  for (const item of items) {
    const key =
      (typeof item.lineId === "string" && item.lineId.trim()) ||
      `${item.productId}:${item.comboId ?? ""}:${item.variantName ?? ""}:${item.notes ?? ""}`;
    const existing = collapsed.get(key);
    if (existing) {
      collapsed.set(key, {
        ...existing,
        quantity:
          (Number(existing.quantity) || 1) + (Number(item.quantity) || 1),
      });
    } else {
      collapsed.set(key, { ...item });
    }
  }
  return Array.from(collapsed.values());
}

export function buildOrderCompatPayload(input: {
  items: Array<CartProductInput | CartItem>;
  sessionId?: number | null;
  locationId?: number | null;
  tableId?: number | null;
  orderType?: string;
  customerId?: number | string | null;
  customerName?: string | null;
  paymentMethod?: string;
  notes?: string;
  kitchenNotes?: string;
}): OrderCompatPayload {
  return {
    sessionId: idNumber(input.sessionId),
    tableId: idNumber(input.tableId),
    locationId: idNumber(input.locationId),
    type: input.orderType ?? "",
    customerId: idNumber(input.customerId),
    customerName: input.customerName ?? "",
    paymentMethod: input.paymentMethod ?? "",
    notes: input.notes ?? "",
    kitchenNotes: input.kitchenNotes ?? "",
    items: collapseProducts(input.items).map(compatItem),
  };
}

export function buildResumeOrderPayload(input: {
  orderId: number;
  items: Array<CartProductInput | CartItem>;
  locationId?: number | null;
  tableId?: number | null;
  orderType?: string;
  customerId?: number | string | null;
  customerName?: string | null;
  notes?: string;
  kitchenNotes?: string;
}): OrderCompatPayload & { orderId: number } {
  return {
    ...buildOrderCompatPayload(input),
    orderId: idNumber(input.orderId),
  };
}

export async function createOrderCompat(
  payload: OrderCompatPayload,
): Promise<CreateOrderResponse> {
  const response = await unwrap(
    restClient.post<OrderCompatResponse>("/pos/orders/compat", payload),
  );
  return {
    orderId: response.orderId == null ? undefined : String(response.orderId),
    orderNumber: response.orderNumber,
    message: response.message,
  };
}

export async function persistResumedOrder(
  payload: OrderCompatPayload & { orderId: number },
): Promise<void> {
  const response = await unwrap(
    restClient.put<OrderCompatResponse>(
      `/pos/orders/${encodeURIComponent(payload.orderId)}/resume`,
      payload,
    ),
  );
  if (response.success !== true) {
    throw new Error("The server did not confirm the resumed order update.");
  }
}
