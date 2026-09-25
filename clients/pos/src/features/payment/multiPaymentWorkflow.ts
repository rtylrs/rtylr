import type { PaymentEntry } from '../../shared/types/pos';
import type { ProcessPaymentRequest } from '@voxire/voxie/pos/orders/payments/v1/types';

interface FinalizeMultiPaymentParams {
  orderId: string | undefined;
  payments: PaymentEntry[];
  chargeOrder: (payload: ProcessPaymentRequest) => Promise<void>;
}

function parseOrderId(orderId: string | undefined): number {
  const numericOrderId = Number(orderId);
  if (!orderId || !Number.isInteger(numericOrderId) || numericOrderId <= 0) {
    throw new Error('Select a valid order before charging.');
  }
  return numericOrderId;
}

// The backend payment endpoint accepts a single { payment_method, amount_paid }.
// A split tender is recorded as one payment for the combined amount under the
// method that covered the largest share. The tip is already collected as part
// of the tender amounts (the screen requires payments to cover the tip-inclusive
// total), so it must NOT be added again here.
export function collapsePayments(
  orderId: string,
  payments: PaymentEntry[]
): ProcessPaymentRequest {
  if (payments.length === 0) {
    throw new Error('Add at least one payment before charging.');
  }

  const byMethod = new Map<PaymentEntry['method'], number>();
  let amountPaid = 0;
  for (const entry of payments) {
    const amount = Number.isFinite(entry.amount) && entry.amount > 0 ? entry.amount : 0;
    amountPaid += amount;
    byMethod.set(entry.method, (byMethod.get(entry.method) ?? 0) + amount);
  }

  if (amountPaid <= 0) {
    throw new Error('Payment amounts must be greater than zero.');
  }

  let dominantMethod: PaymentEntry['method'] = payments[0].method;
  let dominantAmount = -1;
  for (const [method, amount] of byMethod) {
    if (amount > dominantAmount) {
      dominantMethod = method;
      dominantAmount = amount;
    }
  }

  return {
    orderId,
    paymentMethod: dominantMethod,
    amountPaid,
  };
}

export async function finalizeMultiPayment({
  orderId,
  payments,
  chargeOrder,
}: FinalizeMultiPaymentParams): Promise<void> {
  const parsedOrderId = parseOrderId(orderId);
  await chargeOrder(collapsePayments(String(parsedOrderId), payments));
}
