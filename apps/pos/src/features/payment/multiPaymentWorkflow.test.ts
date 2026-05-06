import { describe, expect, it, vi } from 'vitest';
import { finalizeMultiPayment } from './multiPaymentWorkflow';
import type { PaymentEntry } from '../../shared/types/pos';

const payments: PaymentEntry[] = [
  {
    id: 'p1',
    method: 'card',
    amount: 20,
    createdAt: '2026-05-06T00:00:00.000Z',
  },
];

describe('finalizeMultiPayment', () => {
  it('rejects invalid order ids before charging', async () => {
    const chargeOrder = vi.fn(async () => undefined);

    await expect(
      finalizeMultiPayment({
        orderId: 'current',
        payments,
        tip: 0,
        total: 20,
        chargeOrder,
      })
    ).rejects.toThrow('Select a valid order before charging.');

    expect(chargeOrder).not.toHaveBeenCalled();
  });

  it('charges a valid order with the full multi-payment payload', async () => {
    const chargeOrder = vi.fn(async () => undefined);

    await finalizeMultiPayment({
      orderId: '42',
      payments,
      tip: 2,
      total: 22,
      chargeOrder,
    });

    expect(chargeOrder).toHaveBeenCalledWith(42, { payments, tip: 2, total: 22 });
  });
});
