import type { CartItem } from '../../shared/stores/cartStore';
import type { Order } from '../../shared/types/pos';
import type { CreateOrderRequest, CreateOrderResponse } from '@voxire/voxie/pos/orders/v1/types';
import type { ProcessPaymentRequest } from '@voxire/voxie/pos/orders/payments/v1/types';
import { roundCents } from '../../shared/utils/orderCalc';

export type CheckoutPaymentMethod = 'card' | 'cash' | 'qr' | 'gift_card';

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
}

interface ProcessCheckoutParams {
  order: Order;
  payload: CreateOrderRequest;
  method: CheckoutPaymentMethod;
  existingOrderId?: number | null;
  createOrder: (payload: CreateOrderRequest) => Promise<CreateOrderResponse>;
  chargeOrder: (payload: ProcessPaymentRequest) => Promise<void>;
  enqueueOffline: (payload: CreateOrderRequest) => void;
  recordCashSale: (amount: number, orderId: string) => void;
  onOrderCreated?: (orderId: number) => void;
}

type ProcessCheckoutResult = { status: 'online'; orderId: number } | { status: 'queued' };

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
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
  if (error && typeof error === 'object' && 'statusCode' in error) {
    const status = (error as { statusCode?: unknown }).statusCode;
    if (typeof status === 'number') {
      return status >= 500 || status === 408 || status === 429;
    }
  }
  return true;
}

export function buildCashTender(amountInput: string, totalDue: number): CashTender {
  const received = Number(amountInput);
  const normalizedReceived = Number.isFinite(received) && received > 0 ? received : 0;
  const normalizedTotalDue = Number.isFinite(totalDue) && totalDue > 0 ? totalDue : 0;
  const canCharge = normalizedReceived >= normalizedTotalDue && normalizedTotalDue > 0;

  return {
    received: normalizedReceived,
    change: canCharge ? roundCents(normalizedReceived - normalizedTotalDue) : null,
    canCharge,
  };
}

// Local Order object used for receipts and the receipt history screen.
export function buildCheckoutOrder({
  id,
  items,
  subtotal,
  tax,
  total,
  tip = 0,
  customerId,
  createdAt,
}: BuildCheckoutOrderParams): Order {
  return {
    id,
    status: 'paid',
    items: items.map((item) => ({
      id: item.id,
      name: item.name,
      price: item.price,
      quantity: item.quantity,
      notes: item.notes ?? '',
    })),
    subtotal,
    tax,
    tip,
    discount: 0,
    total,
    createdAt,
    paidAt: createdAt,
    customerId: String(customerId ?? ''),
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
}: BuildCreateOrderPayloadParams): CreateOrderRequest {
  return {
    locationId: idString(locationId),
    tableId: idString(tableId),
    sessionId: idString(sessionId),
    type: orderType ?? '',
    customerId: idString(customerId),
    customerName: customerName ?? '',
    paymentMethod: method,
    notes: notes ?? '',
    kitchenNotes: '',
    items: items.map((item) => ({
      menuItemId: String(item.productId),
      name: item.name,
      sku: item.sku ?? '',
      quantity: item.quantity,
      unitPrice: item.unitPrice,
      discountPct: item.discountPct ?? 0,
      taxRate: item.taxRate ?? 0,
      notes: item.notes ?? '',
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
  enqueueOffline,
  recordCashSale,
  onOrderCreated,
}: ProcessCheckoutParams): Promise<ProcessCheckoutResult> {
  const persistedOrderId = positiveInteger(existingOrderId);
  let orderId = persistedOrderId;

  if (!orderId) {
    try {
      const createRes = await createOrder(payload);
      orderId = positiveInteger(createRes.orderId);
    } catch (error) {
      if (method === 'cash' && isRetryableCheckoutFailure(error)) {
        enqueueOffline(payload);
        recordCashSale(order.total, order.id);
        return { status: 'queued' };
      }
      throw new Error('Payment failed. Please try again.');
    }
    if (!orderId || orderId <= 0) {
      throw new Error('Payment failed. Please try again.');
    }
    onOrderCreated?.(orderId);
  }

  // The order now exists server-side: a charge failure must never enqueue the
  // create payload again, or the offline replay would duplicate the order.
  try {
    await chargeOrder({
      orderId: String(orderId),
      paymentMethod: method,
      amountPaid: order.total,
    });
  } catch {
    throw new Error('Payment failed. Please try again.');
  }

  if (method === 'cash') {
    recordCashSale(order.total, order.id);
  }

  return { status: 'online', orderId };
}
