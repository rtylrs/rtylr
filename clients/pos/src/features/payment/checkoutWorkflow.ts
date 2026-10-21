import type { CartItem } from "../../shared/stores/cartStore";
import type { Order, OfflineTender } from "../../shared/types/pos";
import type {
  CreateOrderRequest,
  CreateOrderResponse,
} from "@voxire/voxie/pos/orders/v1/types";
import type { ProcessPaymentRequest } from "@voxire/voxie/pos/orders/payments/v1/types";
import { roundCents } from "../../shared/utils/orderCalc";

export type CheckoutPaymentMethod = "card" | "cash" | "qr" | "gift_card";

/**
 * One charge, identified by the user action that started it.
 *
 * `Idempotency-Key` is a required header on the POS payment route (the server
 * answers 428 without it) and is also what dedups a retry, so it belongs to the
 * action rather than to the request: every attempt of this checkout sends the
 * same key, and a later, different sale gets its own. It travels in the same
 * object as the request so no transport can drop it on the floor.
 */
export type ChargeOrderRequest = ProcessPaymentRequest & { idempotencyKey: string };

/**
 * The money a queued offline sale already took, captured here because
 * CreateOrderRequest carries no amount - and the drain has nothing else to
 * charge. Taken from the same figures `recordCashSale` books, so the replay
 * settles exactly what the drawer was credited with.
 */
export function buildOfflineTender(
  method: CheckoutPaymentMethod,
  order: Order,
): OfflineTender {
  return { paymentMethod: method, amountPaid: order.total, localOrderId: order.id };
}

interface CashTender {
  received: number;
  change: number | null;
  canCharge: boolean;
}

interface BuildCheckoutOrderParams {
  id: string;
  items: CartItem[];
  subtotal: number;
  tax: number;
  discount?: number;
  total: number;
  tip?: number;
  customerId?: string | number | null;
  createdAt: string;
}

interface BuildCreateOrderPayloadParams {
  items: CartItem[];
  method: CheckoutPaymentMethod;
  locationId?: number | null;
  tableId?: number | null;
  sessionId?: number | null;
  orderType?: string;
  customerId?: string | number | null;
  customerName?: string | null;
  notes?: string;
  kitchenNotes?: string;
}

interface ProcessCheckoutParams {
  order: Order;
  payload: CreateOrderRequest;
  method: CheckoutPaymentMethod;
  existingOrderId?: number | null;
  createOrder: (payload: CreateOrderRequest) => Promise<CreateOrderResponse>;
  chargeOrder: (request: ChargeOrderRequest) => Promise<void>;
  persistExistingOrder?: () => Promise<void>;
  enqueueOffline: (payload: CreateOrderRequest) => void;
  recordCashSale: (amount: number, orderId: string) => void;
  onOrderCreated?: (orderId: number) => void;
}

type ProcessCheckoutResult =
  { status: "online"; orderId: number } | { status: "queued" };

function positiveInteger(value: unknown): number | undefined {
  const numeric =
    typeof value === "string" ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
}

function idString(value: unknown): string {
  return String(positiveInteger(value) ?? 0);
}

// True when a failure looks transport-related (no HTTP response at all, or a
// 5xx/408/429 from the server). Definitive 4xx business rejections must not be
// replayed offline - they would fail forever as poison queue entries.
export function isRetryableCheckoutFailure(error: unknown): boolean {
  if (error && typeof error === "object" && "statusCode" in error) {
    const status = (error as { statusCode?: unknown }).statusCode;
    if (typeof status === "number") {
      return status >= 500 || status === 408 || status === 429;
    }
  }
  return true;
}

export function buildCashTender(
  amountInput: string,
  totalDue: number,
): CashTender {
  const received = Number(amountInput);
  const normalizedReceived =
    Number.isFinite(received) && received > 0 ? received : 0;
  const normalizedTotalDue =
    Number.isFinite(totalDue) && totalDue > 0 ? totalDue : 0;
  const canCharge =
    normalizedReceived >= normalizedTotalDue && normalizedTotalDue > 0;

  return {
    received: normalizedReceived,
    change: canCharge
      ? roundCents(normalizedReceived - normalizedTotalDue)
      : null,
    canCharge,
  };
}

// Local Order object used for receipts and the receipt history screen.
export function buildCheckoutOrder({
  id,
  items,
  subtotal,
  tax,
  discount = 0,
  total,
  tip = 0,
  customerId,
  createdAt,
}: BuildCheckoutOrderParams): Order {
  return {
    id,
    status: "paid",
    items: items.map((item) => ({
      id: item.id,
      name: item.name,
      price: item.price,
      quantity: item.quantity,
      notes: item.notes ?? "",
    })),
    subtotal,
    tax,
    tip,
    discount: Math.max(0, discount),
    total,
    createdAt,
    paidAt: createdAt,
    customerId: String(customerId ?? ""),
  };
}

// API payload matching the backend CreateOrderRequest contract. The backend
// computes totals from items, so unit_price/tax_rate must be per-line.
export function buildCreateOrderPayload({
  items,
  method,
  locationId,
  tableId,
  sessionId,
  orderType,
  customerId,
  customerName,
  notes,
  kitchenNotes,
}: BuildCreateOrderPayloadParams): CreateOrderRequest {
  return {
    locationId: idString(locationId),
    tableId: idString(tableId),
    sessionId: idString(sessionId),
    type: orderType ?? "",
    customerId: idString(customerId),
    customerName: customerName ?? "",
    paymentMethod: method,
    notes: notes ?? "",
    kitchenNotes: kitchenNotes ?? "",
    items: items.map((item) => ({
      menuItemId: String(item.productId),
      name: item.name,
      sku: item.sku ?? "",
      quantity: item.quantity,
      unitPrice: item.unitPrice,
      discountPct: item.discountPct ?? 0,
      taxRate: item.taxRate ?? 0,
      notes: item.notes ?? "",
    })),
  };
}

export async function processCheckout({
  order,
  payload,
  method,
  existingOrderId,
  createOrder,
  chargeOrder,
  persistExistingOrder,
  enqueueOffline,
  recordCashSale,
  onOrderCreated,
}: ProcessCheckoutParams): Promise<ProcessCheckoutResult> {
  // Minted ONCE, above every retry. Generating it per attempt would give a
  // double-tap or a retried drain two keys, so both would execute and the
  // server's fingerprint check could not tell them apart from two real sales.
  const idempotencyKey = crypto.randomUUID();

  const persistedOrderId = positiveInteger(existingOrderId);
  let orderId = persistedOrderId;

  if (!orderId) {
    try {
      const createRes = await createOrder(payload);
      orderId = positiveInteger(createRes.orderId);
    } catch (error) {
      if (method === "cash" && isRetryableCheckoutFailure(error)) {
        // The caller stores the tender with the entry, in one write, so the
        // drain can never find a queued sale it cannot charge.
        enqueueOffline(payload);
        recordCashSale(order.total, order.id);
        return { status: "queued" };
      }
      throw new Error("Payment failed. Please try again.");
    }
    if (!orderId || orderId <= 0) {
      throw new Error("Payment failed. Please try again.");
    }
    onOrderCreated?.(orderId);
  }

  // A resumed order already has a server identity. Persist the current cart
  // before charging it; otherwise edits made after resume are silently ignored
  // while the old lines are charged.
  if (orderId && persistExistingOrder) {
    try {
      await persistExistingOrder();
    } catch {
      throw new Error("Could not save the resumed order. Please try again.");
    }
  }

  // The order now exists server-side: a charge failure must never enqueue the
  // create payload again, or the offline replay would duplicate the order.
  try {
    await chargeOrder({
      orderId: String(orderId),
      paymentMethod: method,
      amountPaid: order.total,
      idempotencyKey,
    });
  } catch {
    throw new Error("Payment failed. Please try again.");
  }

  if (method === "cash") {
    recordCashSale(order.total, order.id);
  }

  return { status: "online", orderId };
}
