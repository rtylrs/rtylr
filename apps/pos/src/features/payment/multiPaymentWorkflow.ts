import type { PaymentEntry } from '../../shared/types/pos';
import type { PaymentPayload } from '../../shared/services/posService';

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

export async function finalizeMultiPayment({
  orderId,
  payments,
  tip,
  total,
  chargeOrder,
}: FinalizeMultiPaymentParams): Promise<void> {
  await chargeOrder(parseOrderId(orderId), { payments, tip, total });
}
