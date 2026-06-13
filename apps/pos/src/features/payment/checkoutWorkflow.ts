import type { CartItem } from '../../shared/stores/cartStore';
import type { Order } from '../../shared/types/pos';
import type {
  CreateOrderPayload,
  PaymentPayload,
} from '../../shared/services/api/posService';
import { roundCents } from '../../shared/utils/orderCalc';

export type CheckoutPaymentMethod = 'card' | 'cash' | 'qr';

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
  customerId?: string | number | null;
  customerName?: string | null;
  notes?: string;
}

interface ProcessCheckoutParams {
  order: Order;
  payload: CreateOrderPayload;
  method: CheckoutPaymentMethod;
  // The API client camelCases responses, so the backend's order_id arrives as
  // orderId; accept all shapes so neither the runtime contract nor older mocks break.
  createOrder: (payload: CreateOrderPayload) => Promise<{ orderId?: number; id?: number; order_id?: number }>;
  chargeOrder: (orderId: number, payload: PaymentPayload) => Promise<void>;
  enqueueOffline: (payload: CreateOrderPayload) => void;
  recordCashSale: (amount: number, orderId: string) => void;
}

type ProcessCheckoutResult = { status: 'online'; orderId: number } | { status: 'queued' };

function positiveInteger(value: unknown): number | undefined {
  const numeric = typeof value === 'string' ? Number(value.trim()) : Number(value);
  if (!Number.isInteger(numeric) || numeric <= 0) return undefined;
  return numeric;
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
  customerId,
  customerName,
  notes,
}: BuildCreateOrderPayloadParams): CreateOrderPayload {
  return {
    location_id: positiveInteger(locationId),
    table_id: positiveInteger(tableId),
    session_id: positiveInteger(sessionId),
    customer_id: positiveInteger(customerId),
    customer_name: customerName ?? undefined,
    payment_method: method,
    notes,
    items: items.map((item) => ({
      menu_item_id: item.productId,
      name: item.name,
      quantity: item.quantity,
      unit_price: item.unitPrice,
      discount_pct: item.discountPct ?? 0,
      tax_rate: item.taxRate ?? 0,
      notes: item.notes ?? '',
    })),
  };
}

export async function processCheckout({
  order,
  payload,
  method,
  createOrder,
  chargeOrder,
  enqueueOffline,
  recordCashSale,
}: ProcessCheckoutParams): Promise<ProcessCheckoutResult> {
  try {
    const createRes = await createOrder(payload);
    const orderId = createRes.orderId ?? createRes.id ?? createRes.order_id;
    if (!orderId || orderId <= 0) {
      throw new Error('Order was not accepted by the server.');
    }

    await chargeOrder(orderId, {
      payment_method: method,
      amount_paid: order.total,
    });

    if (method === 'cash') {
      recordCashSale(order.total, order.id);
    }

    return { status: 'online', orderId };
  } catch {
    if (method === 'cash') {
      enqueueOffline(payload);
      recordCashSale(order.total, order.id);
      return { status: 'queued' };
    }

    throw new Error('Payment failed. Please try again.');
  }
}
