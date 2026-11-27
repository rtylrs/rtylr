import type { Order } from '@voxire/voxie/pos/orders/v1/types';
import type { PaymentMethod, Receipt } from '../../shared/types/pos';

function normalizePaymentMethod(method: string | undefined): PaymentMethod {
  switch (method) {
    case 'cash':
    case 'card':
    case 'gift_card':
      return method;
    default:
      return 'other';
  }
}

export function buildReceiptFromOrder(order: Order): Receipt {
  const id = String(order.id ?? '');
  const timestamp = order.paidAt || order.createdAt || new Date().toISOString();
  const total = order.total ?? 0;
  const items = (order.items ?? []).map((item) => {
    const qty = item.quantity ?? 1;
    const unitPrice =
      item.unitPrice != null && item.unitPrice > 0
        ? item.unitPrice
        : qty > 0 && item.total != null
          ? item.total / qty
          : 0;
    return {
      name: item.name || 'Item',
      qty,
      price: unitPrice,
    };
  });

  const amount =
    order.paymentStatus === 'paid' && (order.amountPaid ?? 0) > 0 ? order.amountPaid! : total;

  return {
    id,
    orderId: order.orderNumber || id,
    total,
    customerName: order.customerName || undefined,
    printedAt: timestamp,
    items,
    payments: [
      {
        id: `payment-${id}`,
        method: normalizePaymentMethod(order.paymentMethod),
        amount,
        createdAt: timestamp,
      },
    ],
  };
}
