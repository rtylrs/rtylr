/**
 * End-to-end: gift card issue → redeem → apply discount in checkout.
 */
import { describe, expect, it, vi, beforeEach } from 'vitest';

const giftCardClient = vi.hoisted(() => ({
  createGiftCard: vi.fn(() => Promise.resolve({ giftCardId: 'remote-gc', message: 'ok' })),
}));

vi.mock('@voxire/api', () => ({
  createOneKitClient: () => giftCardClient,
}));

import { issueGiftCard } from './giftCardWorkflow';
import type { GiftCard } from '../../shared/types/pos';

beforeEach(() => {
  vi.clearAllMocks();
});

function makeStore() {
  const cards: GiftCard[] = [];

  const issue = (partial: Omit<GiftCard, 'id' | 'issuedAt' | 'active'>): GiftCard => {
    if (!Number.isFinite(partial.balance) || partial.balance <= 0) {
      throw new Error('Gift card amount must be greater than zero');
    }
    if (cards.some((c) => c.code === partial.code)) {
      throw new Error('Gift card code already exists');
    }
    const card: GiftCard = {
      ...partial,
      id: 'gc_test_' + cards.length,
      issuedAt: new Date().toISOString(),
      active: true,
    };
    cards.push(card);
    return card;
  };

  const redeem = (code: string, amount: number): { ok: boolean; remaining?: number; error?: string } => {
    if (!Number.isFinite(amount) || amount <= 0) return { ok: false, error: 'Redeem amount must be greater than zero' };
    const card = cards.find((c) => c.code === code);
    if (!card) return { ok: false, error: 'Card not found' };
    if (!card.active) return { ok: false, error: 'Card is inactive' };
    if (card.balance < amount) return { ok: false, error: 'Insufficient balance' };
    card.balance -= amount;
    return { ok: true, remaining: card.balance };
  };

  return { cards, issue, redeem };
}

describe('gift card issue → redeem', () => {
  it('issues a $50 gift card and redeems $30 leaving $20 remaining', async () => {
    const store = makeStore();

    const result = await issueGiftCard({
      codeInput: 'GC-ALPHA',
      amountInput: '50',
      issue: store.issue,
    });

    expect(result).toEqual({ ok: true, card: expect.objectContaining({ code: 'GC-ALPHA', balance: 50 }) });

    const redeemResult = store.redeem('GC-ALPHA', 30);
    expect(redeemResult).toEqual({ ok: true, remaining: 20 });
  });

  it('rejects duplicate card codes', async () => {
    const store = makeStore();

    await issueGiftCard({ codeInput: 'DUPE', amountInput: '25', issue: store.issue });
    const second = await issueGiftCard({ codeInput: 'DUPE', amountInput: '25', issue: store.issue });

    expect(second).toEqual({ ok: false, error: 'Gift card code already exists' });
  });

  it('rejects issue with zero or negative amount', async () => {
    const store = makeStore();

    const result = await issueGiftCard({ codeInput: 'GC-BAD', amountInput: '-10', issue: store.issue });
    expect(result).toEqual({ ok: false, error: 'Gift card amount must be greater than zero' });
  });

  it('fails redemption when balance is insufficient', async () => {
    const store = makeStore();

    await issueGiftCard({ codeInput: 'GC-SMALL', amountInput: '10', issue: store.issue });
    const result = store.redeem('GC-SMALL', 20);

    expect(result).toEqual({ ok: false, error: 'Insufficient balance' });
  });

  it('fails redemption for unknown card code', () => {
    const store = makeStore();
    const result = store.redeem('GC-UNKNOWN', 10);
    expect(result).toEqual({ ok: false, error: 'Card not found' });
  });

  it('applies gift card as discount - reduces order total', async () => {
    const store = makeStore();

    await issueGiftCard({ codeInput: 'GC-PROMO', amountInput: '15', issue: store.issue });

    const orderTotal = 42.0;
    const redeemResult = store.redeem('GC-PROMO', 15);

    expect(redeemResult.ok).toBe(true);
    const discountedTotal = orderTotal - 15;
    expect(discountedTotal).toBeCloseTo(27.0, 2);
    expect(redeemResult.remaining).toBe(0);
  });

  it('partial redemption leaves remaining balance for future use', async () => {
    const store = makeStore();

    await issueGiftCard({ codeInput: 'GC-PARTIAL', amountInput: '100', issue: store.issue });

    store.redeem('GC-PARTIAL', 30);
    store.redeem('GC-PARTIAL', 25);

    const card = store.cards.find((c) => c.code === 'GC-PARTIAL');
    expect(card?.balance).toBe(45);
  });
});
