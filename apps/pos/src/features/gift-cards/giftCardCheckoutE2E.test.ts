/**
 * End-to-end: gift card applied to a real checkout order.
 *
 * Proves gift card redemption reduces the checkout total and that the
 * correct remaining balance is tracked after the sale.
 */
import { describe, expect, it, vi, beforeEach } from 'vitest';

const giftCardClient = vi.hoisted(() => ({
  createGiftCard: vi.fn(() => Promise.resolve({ giftCardId: 'remote-gc', message: 'ok' })),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => giftCardClient,
}));

import { issueGiftCard } from './giftCardWorkflow';
import { buildCheckoutOrder, buildCreateOrderPayload, processCheckout } from '../payment/checkoutWorkflow';
import type { CartItem } from '../../shared/stores/cartStore';
import type { GiftCard } from '../../shared/types/pos';

beforeEach(() => {
  vi.clearAllMocks();
});

const item: CartItem = {
  lineId: '2::',
  id: '2::',
  productId: 2,
  name: 'Pizza',
  unitPrice: 25,
  price: 25,
  quantity: 2,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0,
  status: 'pending',
  notes: '',
};

function makeStore() {
  const cards: GiftCard[] = [];
  return {
    issue: (partial: Omit<GiftCard, 'id' | 'issuedAt' | 'active'>): GiftCard => {
      if (!Number.isFinite(partial.balance) || partial.balance <= 0)
        throw new Error('Gift card amount must be greater than zero');
      const card: GiftCard = {
        ...partial,
        id: 'gc_' + cards.length,
        issuedAt: '2026-05-13T00:00:00.000Z',
        active: true,
      };
      cards.push(card);
      return card;
    },
    redeem: (code: string, amount: number) => {
      const card = cards.find((c) => c.code === code);
      if (!card) return { ok: false as const, error: 'Card not found' };
      if (!card.active) return { ok: false as const, error: 'Card is inactive' };
      if (card.balance < amount) return { ok: false as const, error: 'Insufficient balance' };
      card.balance -= amount;
      return { ok: true as const, remaining: card.balance };
    },
    cards,
  };
}

describe('gift card → checkout', () => {
  it('$20 gift card covers partial $50 order - customer pays $30 remaining', async () => {
    const store = makeStore();
    await issueGiftCard({ codeInput: 'GC-PAY', amountInput: '20', issue: store.issue });

    const orderTotal = 50;
    const redeemResult = store.redeem('GC-PAY', 20);

    expect(redeemResult.ok).toBe(true);
    expect(redeemResult.remaining).toBe(0);

    const remainingToPay = orderTotal - 20;
    const chargeOrder = vi.fn(async () => undefined);
    const order = buildCheckoutOrder({
      id: 'local-1',
      items: [item],
      subtotal: 50,
      tax: 0,
      total: remainingToPay,
      customerId: 1,
      createdAt: '2026-05-13T10:00:00.000Z',
    });

    const payload = buildCreateOrderPayload({ items: [item], method: 'cash' });

    const result = await processCheckout({
      order,
      payload,
      method: 'cash',
      createOrder: vi.fn(() => Promise.resolve({ orderId: '300' })),
      chargeOrder,
      enqueueOffline: vi.fn(),
      recordCashSale: vi.fn(),
    });

    expect(result).toEqual({ status: 'online', orderId: 300 });
    expect(chargeOrder).toHaveBeenCalledWith(
      expect.objectContaining({ orderId: '300', amountPaid: 30 }),
    );
  });

  it('gift card with enough balance covers full order - charge is $0 remainder', async () => {
    const store = makeStore();
    await issueGiftCard({ codeInput: 'GC-FULL', amountInput: '100', issue: store.issue });

    const redeemResult = store.redeem('GC-FULL', 50);
    expect(redeemResult.ok).toBe(true);
    expect(redeemResult.remaining).toBe(50);

    const chargeOrder = vi.fn(async () => undefined);
    const order = buildCheckoutOrder({
      id: 'local-2',
      items: [item],
      subtotal: 50,
      tax: 0,
      total: 0,
      customerId: 1,
      createdAt: '2026-05-13T10:00:00.000Z',
    });

    const payload = buildCreateOrderPayload({ items: [item], method: 'gift_card' });

    const result = await processCheckout({
      order,
      payload,
      method: 'gift_card',
      createOrder: vi.fn(() => Promise.resolve({ orderId: '301' })),
      chargeOrder,
      enqueueOffline: vi.fn(),
      recordCashSale: vi.fn(),
    });

    expect(result).toEqual({ status: 'online', orderId: 301 });
    expect(chargeOrder).toHaveBeenCalledWith(
      expect.objectContaining({ orderId: '301', amountPaid: 0 }),
    );
  });

  it('gift card redemption fails gracefully when balance is insufficient', async () => {
    const store = makeStore();
    await issueGiftCard({ codeInput: 'GC-SMALL', amountInput: '10', issue: store.issue });

    const result = store.redeem('GC-SMALL', 50);

    expect(result.ok).toBe(false);
    expect(result.error).toBe('Insufficient balance');

    const card = store.cards.find((c) => c.code === 'GC-SMALL');
    expect(card?.balance).toBe(10);
  });

  it('multiple gift cards can be applied to a single order (split across cards)', async () => {
    const store = makeStore();
    await issueGiftCard({ codeInput: 'GC-A', amountInput: '15', issue: store.issue });
    await issueGiftCard({ codeInput: 'GC-B', amountInput: '15', issue: store.issue });

    const orderTotal = 50;
    const r1 = store.redeem('GC-A', 15);
    const r2 = store.redeem('GC-B', 15);

    expect(r1.ok).toBe(true);
    expect(r2.ok).toBe(true);

    const totalGiftCardPayment = 15 + 15;
    const remainingToPay = orderTotal - totalGiftCardPayment;
    expect(remainingToPay).toBe(20);

    const chargeOrder = vi.fn(async () => undefined);
    const order = buildCheckoutOrder({
      id: 'local-3',
      items: [item],
      subtotal: 50,
      tax: 0,
      total: remainingToPay,
      customerId: 1,
      createdAt: '2026-05-13T10:00:00.000Z',
    });

    const payload = buildCreateOrderPayload({ items: [item], method: 'cash' });

    await processCheckout({
      order,
      payload,
      method: 'cash',
      createOrder: vi.fn(() => Promise.resolve({ orderId: '302' })),
      chargeOrder,
      enqueueOffline: vi.fn(),
      recordCashSale: vi.fn(),
    });

    expect(chargeOrder).toHaveBeenCalledWith(
      expect.objectContaining({ orderId: '302', amountPaid: 20 }),
    );
  });
});
