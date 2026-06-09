import type { CartItem } from '../../shared/stores/cartStore';
import type { Order } from '../../shared/types/pos';
import type { PaymentPayload } from '../../shared/services/posService';
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
  employeeId?: string | number | null;
  createdAt: string;
}

interface ProcessCheckoutParams {
  order: Order;
  method: CheckoutPaymentMethod;
  createOrder: (order: Order) => Promise<{ id?: number; order_id?: number }>;
  chargeOrder: (orderId: number, payload: PaymentPayload) => Promise<void>;
  enqueueOffline: (order: Order) => void;
  recordCashSale: (amount: number, orderId: string) => void;
}

type ProcessCheckoutResult = { status: 'online'; orderId: number } | { status: 'queued' };

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

export function buildCheckoutOrder({
  id,
  items,
  subtotal,
  tax,
  total,
  tip = 0,
  employeeId,
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
    customerId: String(employeeId ?? ''),
  };
}

export async function processCheckout({
  order,
  method,
  createOrder,
  chargeOrder,
  enqueueOffline,
  recordCashSale,
}: ProcessCheckoutParams): Promise<ProcessCheckoutResult> {
  try {
    const createRes = await createOrder(order);
    const orderId = createRes.id ?? createRes.order_id;
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
      enqueueOffline(order);
      recordCashSale(order.total, order.id);
      return { status: 'queued' };
    }

    throw new Error('Payment failed. Please try again.');
  }
}
