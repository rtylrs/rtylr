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
}: ProcessCheckoutParams): Promise<ProcessCheckoutResult> {
  const persistedOrderId = positiveInteger(existingOrderId);
  try {
    let orderId = persistedOrderId;
    if (!orderId) {
      const createRes = await createOrder(payload);
      orderId = positiveInteger(createRes.orderId);
      if (!orderId || orderId <= 0) {
        throw new Error('Order was not accepted by the server.');
      }
    }

    await chargeOrder({
      orderId: String(orderId),
      paymentMethod: method,
      amountPaid: order.total,
    });

    if (method === 'cash') {
      recordCashSale(order.total, order.id);
    }

    return { status: 'online', orderId };
  } catch {
    if (method === 'cash' && !persistedOrderId) {
      enqueueOffline(payload);
      recordCashSale(order.total, order.id);
      return { status: 'queued' };
    }

    throw new Error('Payment failed. Please try again.');
  }
}
