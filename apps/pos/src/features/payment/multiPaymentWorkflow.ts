import type { PaymentEntry } from '../../shared/types/pos';
import type { PaymentPayload } from '../../shared/services/api/posService';

interface FinalizeMultiPaymentParams {
  orderId: string | undefined;
  payments: PaymentEntry[];
  tip: number;
  total: number;
  chargeOrder: (orderId: number, payload: PaymentPayload) => Promise<void>;
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
// method that covered the largest share.
export function collapsePayments(payments: PaymentEntry[], tip: number): PaymentPayload {
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
    payment_method: dominantMethod,
    amount_paid: amountPaid + (Number.isFinite(tip) && tip > 0 ? tip : 0),
  };
}

export async function finalizeMultiPayment({
  orderId,
  payments,
  tip,
  chargeOrder,
}: FinalizeMultiPaymentParams): Promise<void> {
  await chargeOrder(parseOrderId(orderId), collapsePayments(payments, tip));
}
