import type { Receipt } from '../../shared/types/pos';
import type { VoidRefundPayload } from '../../shared/services/api/posService';

interface BuildRefundRequestParams {
  orderId: string | undefined;
  fullRefund: boolean;
  amountInput: string;
  reason: string;
  receipt?: Receipt;
}

interface BuiltRefundRequest {
  orderId: number;
  payload: VoidRefundPayload;
  cashRefundAmount: number;
}

interface ProcessRefundParams extends BuildRefundRequestParams {
  refundOrder: (orderId: number, payload: VoidRefundPayload) => Promise<void>;
  recordCashRefund: (amount: number, orderId: string) => void;
}

function parseOrderId(orderId: string | undefined): number {
  const numericOrderId = Number(orderId);
  if (!orderId || !Number.isInteger(numericOrderId) || numericOrderId <= 0) {
    throw new Error('Select a valid order before refunding.');
  }
  return numericOrderId;
}

function parseRefundAmount(
  fullRefund: boolean,
  amountInput: string,
  receipt: Receipt | undefined,
): number {
  if (fullRefund) {
    const total = receipt?.total ?? 0;
    if (total <= 0) throw new Error('Cannot determine order total for full refund.');
    return total;
  }
  const amount = Number(amountInput);
  if (!Number.isFinite(amount) || amount <= 0) {
    throw new Error('Enter a refund amount greater than zero.');
  }
  return amount;
}

function cashPaid(receipt: Receipt | undefined): number {
  return (
    receipt?.payments
      .filter((payment) => payment.method === 'cash')
      .reduce((sum, payment) => sum + payment.amount, 0) ?? 0
  );
}

export function buildRefundRequest({
  orderId,
  fullRefund,
  amountInput,
  reason,
  receipt,
}: BuildRefundRequestParams): BuiltRefundRequest {
  const numericOrderId = parseOrderId(orderId);
  const amount = parseRefundAmount(fullRefund, amountInput, receipt);
  const cashTendered = cashPaid(receipt);

  return {
    orderId: numericOrderId,
    payload: {
      amount,
      reason,
    },
    cashRefundAmount: Math.min(amount, cashTendered),
  };
}

export async function processRefund({
  refundOrder,
  recordCashRefund,
  ...params
}: ProcessRefundParams): Promise<void> {
  const refund = buildRefundRequest(params);
  await refundOrder(refund.orderId, refund.payload);

  if (refund.cashRefundAmount > 0) {
    recordCashRefund(refund.cashRefundAmount, String(refund.orderId));
  }
}
