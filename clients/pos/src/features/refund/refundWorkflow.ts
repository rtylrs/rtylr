import type { ProcessRefundRequest } from '@voxire/voxie/pos/orders/refunds/v1/types';
import type { Receipt } from '../../shared/types/pos';

interface BuildRefundRequestParams {
  orderId: string | undefined;
  fullRefund: boolean;
  amountInput: string;
  reason: string;
  /** Order total from server order data; preferred over the local receipt. */
  orderTotal?: number;
  receipt?: Receipt;
  /** Tender the cashier confirmed in the Refund UI ('cash' moves the drawer). */
  refundMethod?: string;
}

interface BuiltRefundRequest {
  payload: ProcessRefundRequest;
  cashRefundAmount: number;
}

interface ProcessRefundParams extends BuildRefundRequestParams {
  refundOrder: (payload: ProcessRefundRequest) => Promise<void>;
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
  orderTotal: number | undefined,
  receipt: Receipt | undefined,
): number {
  if (fullRefund) {
    // Server order total first; local receipts are keyed by local refs and
    // are usually absent for server orders.
    const total = orderTotal && orderTotal > 0 ? orderTotal : (receipt?.total ?? 0);
    if (total <= 0) throw new Error('Cannot determine order total for full refund.');
    return total;
  }
  const amount = Number(amountInput);
  if (!Number.isFinite(amount) || amount <= 0) {
    throw new Error('Enter a refund amount greater than zero.');
  }
  return amount;
}

// Cash handed back is recorded from what the cashier confirmed (amount +
// method), not from receipt lookups — stored receipts are keyed by local
// refs and never match server order ids.
function confirmedCashPortion(amount: number, refundMethod: string | undefined): number {
  return refundMethod === 'cash' ? amount : 0;
}

export function buildRefundRequest({
  orderId,
  fullRefund,
  amountInput,
  reason,
  orderTotal,
  receipt,
  refundMethod,
}: BuildRefundRequestParams): BuiltRefundRequest {
  const numericOrderId = parseOrderId(orderId);
  const amount = parseRefundAmount(fullRefund, amountInput, orderTotal, receipt);

  return {
    payload: {
      orderId: String(numericOrderId),
      amount,
      reason,
    },
    cashRefundAmount: confirmedCashPortion(amount, refundMethod),
  };
}

export async function processRefund({
  refundOrder,
  recordCashRefund,
  ...params
}: ProcessRefundParams): Promise<void> {
  const refund = buildRefundRequest(params);
  await refundOrder(refund.payload);

  if (refund.cashRefundAmount > 0) {
    recordCashRefund(refund.cashRefundAmount, refund.payload.orderId ?? '');
  }
}
