import { describe, expect, it, vi } from 'vitest';
import { collapsePayments, finalizeMultiPayment } from './multiPaymentWorkflow';
import type { PaymentEntry } from '../../shared/types/pos';

const payments: PaymentEntry[] = [
  {
    id: 'p1',
    method: 'card',
    amount: 20,
    createdAt: '2026-05-06T00:00:00.000Z',
  },
];

describe('collapsePayments', () => {
  it('collapses single-method payments into one payload', () => {
    expect(collapsePayments('1', payments)).toEqual({
      orderId: '1',
      paymentMethod: 'card',
      amountPaid: 20,
    });
  });

  it('picks dominant method and sums all tendered amounts exactly', () => {
    const split: PaymentEntry[] = [
      { id: 'p1', method: 'cash', amount: 30, createdAt: '2026-05-06T00:00:00.000Z' },
      { id: 'p2', method: 'card', amount: 25, createdAt: '2026-05-06T00:00:00.000Z' },
    ];
    expect(collapsePayments('1', split)).toEqual({
      orderId: '1',
      paymentMethod: 'cash',
      amountPaid: 55,
    });
  });

  it('throws when payments array is empty', () => {
    expect(() => collapsePayments('1', [])).toThrow('Add at least one payment before charging.');
  });

  it('throws when all payment amounts are zero', () => {
    const zeroed: PaymentEntry[] = [
      { id: 'z1', method: 'cash', amount: 0, createdAt: '2026-05-06T00:00:00.000Z' },
    ];
    expect(() => collapsePayments('1', zeroed)).toThrow('Payment amounts must be greater than zero.');
  });
});

describe('finalizeMultiPayment', () => {
  it('rejects invalid order ids before charging', async () => {
    const chargeOrder = vi.fn(async () => undefined);

    await expect(
      finalizeMultiPayment({
        orderId: 'current',
        payments,
        chargeOrder,
      })
    ).rejects.toThrow('Select a valid order before charging.');

    expect(chargeOrder).not.toHaveBeenCalled();
  });

  it('charges a valid order with a collapsed single-method payload', async () => {
    const chargeOrder = vi.fn(async () => undefined);

    await finalizeMultiPayment({
      orderId: '42',
      payments,
      chargeOrder,
    });

    expect(chargeOrder).toHaveBeenCalledWith({
      orderId: '42',
      paymentMethod: 'card',
      amountPaid: 20,
      // Required on this route; the server answers 428 without it.
      idempotencyKey: expect.any(String),
    });
  });
});
