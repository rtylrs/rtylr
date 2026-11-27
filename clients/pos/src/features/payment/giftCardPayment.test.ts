import { describe, expect, it, vi } from 'vitest';
import { finalizeMultiPayment } from './multiPaymentWorkflow';
import { buildCheckoutOrder, buildCreateOrderPayload, processCheckout } from './checkoutWorkflow';

describe('gift card tender redemption in POS payment flows', () => {
  it('redeems gift card entries on the server before charging in finalizeMultiPayment', async () => {
    const events: string[] = [];
    const redeemGiftCard = vi.fn(async () => {
      events.push('redeem');
    });
    const chargeOrder = vi.fn(async () => {
      events.push('charge');
    });

    await finalizeMultiPayment({
      orderId: '88',
      payments: [
        { id: 'p1', method: 'gift_card', amount: 15, reference: 'GC-SPLIT-1', createdAt: '' },
        { id: 'p2', method: 'cash', amount: 10, createdAt: '' },
      ],
      redeemGiftCard,
      chargeOrder,
    });

    expect(redeemGiftCard).toHaveBeenCalledWith({
      code: 'GC-SPLIT-1',
      orderId: '88',
      amount: 15,
    });
    expect(events).toEqual(['redeem', 'charge']);
  });

  it('aborts checkout before marking the order paid if gift card redemption fails', async () => {
    const chargeOrder = vi.fn(async () => undefined);
    const item = {
      lineId: '1::',
      id: '1::',
      productId: 1,
      name: 'Espresso',
      unitPrice: 10,
      price: 10,
      quantity: 1,
      modifiers: [],
      discountPct: 0,
      discountAmount: 0,
      taxRate: 0,
      status: 'pending' as const,
      notes: '',
    };

    await expect(
      processCheckout({
        order: buildCheckoutOrder({
          id: 'local-1',
          items: [item],
          subtotal: 10,
          tax: 0,
          total: 10,
          createdAt: '2026-11-27T12:00:00Z',
        }),
        payload: buildCreateOrderPayload({ items: [item], method: 'gift_card' }),
        method: 'gift_card',
        createOrder: vi.fn(async () => ({ orderId: '501' })),
        redeemGiftCard: vi.fn(async () => {
          throw new Error('Insufficient gift card balance');
        }),
        chargeOrder,
        enqueueOffline: vi.fn(),
        recordCashSale: vi.fn(),
      })
    ).rejects.toThrow('Insufficient gift card balance');

    expect(chargeOrder).not.toHaveBeenCalled();
  });
});
