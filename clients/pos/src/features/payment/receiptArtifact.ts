import type { CheckoutPaymentMethod } from './checkoutWorkflow';
import type { Order, PaymentMethod, Receipt } from '../../shared/types/pos';

const RECEIPT_PAYMENT_METHOD: Record<CheckoutPaymentMethod, PaymentMethod> = {
  cash: 'cash',
  card: 'card',
  gift_card: 'gift_card',
  qr: 'other',
};

interface BuildReceiptFromOrderParams {
  order: Order;
  method: CheckoutPaymentMethod;
  printedAt: string;
}

export function buildReceiptFromOrder({
  order,
  method,
  printedAt,
}: BuildReceiptFromOrderParams): Receipt {
  return {
    id: `receipt-${order.id}`,
    orderId: order.id,
    total: order.total,
    printedAt,
    items: order.items.map((item) => ({
      name: item.name,
      qty: item.quantity,
      price: item.price,
    })),
    payments: [
      {
        id: `payment-${order.id}`,
        method: RECEIPT_PAYMENT_METHOD[method],
        amount: order.total,
        createdAt: printedAt,
      },
    ],
  };
}
