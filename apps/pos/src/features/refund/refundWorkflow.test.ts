import { describe, expect, it, vi } from 'vitest';
import { buildRefundRequest, processRefund } from './refundWorkflow';
import type { Receipt } from '../../shared/types/pos';

const receipt: Receipt = {
  id: 'receipt-order-1',
  orderId: 'order-1',
  total: 42,
  printedAt: '2026-05-06T10:00:00.000Z',
  items: [],
  payments: [
    {
      id: 'payment-order-1',
      method: 'cash',
      amount: 42,
      createdAt: '2026-05-06T10:00:00.000Z',
    },
  ],
};

describe('buildRefundRequest', () => {
  it('rejects non-numeric order ids before calling the API', () => {
    expect(() =>
      buildRefundRequest({
        orderId: 'current',
        fullRefund: true,
        amountInput: '',
        reason: 'Customer complaint',
        receipt,
      })
    ).toThrow('Select a valid order before refunding.');
  });

  it('builds a partial refund request and caps local cash drawer adjustment to cash paid', () => {
    const result = buildRefundRequest({
      orderId: '7',
      fullRefund: false,
      amountInput: '15.50',
      reason: 'Wrong item',
      receipt,
    });

    expect(result).toEqual({
      payload: {
        orderId: '7',
        amount: 15.5,
        reason: 'Wrong item',
      },
      cashRefundAmount: 15.5,
    });
  });
});

describe('processRefund', () => {
  it('records a cash drawer refund only after the API refund succeeds', async () => {
    const refundOrder = vi.fn(async () => undefined);
    const recordCashRefund = vi.fn();

    await processRefund({
      orderId: '7',
      fullRefund: true,
      amountInput: '',
      reason: 'Customer complaint',
      receipt,
      refundOrder,
      recordCashRefund,
    });

    expect(refundOrder).toHaveBeenCalledWith({
      orderId: '7',
      amount: 42,
      reason: 'Customer complaint',
    });
    expect(recordCashRefund).toHaveBeenCalledWith(42, '7');
  });
});
