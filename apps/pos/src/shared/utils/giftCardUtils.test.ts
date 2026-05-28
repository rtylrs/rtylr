import { describe, it, expect } from 'vitest';
import {
  isGiftCardUsable,
  giftCardChargeAmount,
  giftCardBalanceAfterCharge,
  isValidGiftCardCode,
  formatGiftCardBalance,
  applyGiftCardCharge,
} from './giftCardUtils';
import type { GiftCard } from '../types/pos';

const makeCard = (overrides: Partial<GiftCard> = {}): GiftCard => ({
  id: 'gc_1',
  code: 'GC-ABCD-1234',
  balance: 50,
  issuedAt: '2024-01-01T00:00:00Z',
  active: true,
  ...overrides,
});

describe('isGiftCardUsable', () => {
  it('returns true for an active card with a positive balance', () => {
    expect(isGiftCardUsable(makeCard())).toBe(true);
  });

  it('returns false for an inactive card', () => {
    expect(isGiftCardUsable(makeCard({ active: false }))).toBe(false);
  });

  it('returns false for a card with zero balance', () => {
    expect(isGiftCardUsable(makeCard({ balance: 0 }))).toBe(false);
  });

  it('returns false for a card with a negative balance', () => {
    expect(isGiftCardUsable(makeCard({ balance: -5 }))).toBe(false);
  });
});

describe('giftCardChargeAmount', () => {
  it('charges the full amount when the balance covers it', () => {
    expect(giftCardChargeAmount(makeCard({ balance: 100 }), 30)).toBe(30);
  });

  it('caps the charge at the card balance', () => {
    expect(giftCardChargeAmount(makeCard({ balance: 20 }), 50)).toBe(20);
  });

  it('returns 0 for an unusable card', () => {
    expect(giftCardChargeAmount(makeCard({ active: false }), 30)).toBe(0);
  });

  it('returns 0 for a zero amount due', () => {
    expect(giftCardChargeAmount(makeCard(), 0)).toBe(0);
  });

  it('returns 0 for a negative amount due', () => {
    expect(giftCardChargeAmount(makeCard(), -10)).toBe(0);
  });
});

describe('giftCardBalanceAfterCharge', () => {
  it('reduces the balance by the charged amount', () => {
    expect(giftCardBalanceAfterCharge(makeCard({ balance: 50 }), 20)).toBe(30);
  });

  it('floors at 0 if charged amount exceeds balance', () => {
    expect(giftCardBalanceAfterCharge(makeCard({ balance: 10 }), 50)).toBe(0);
  });

  it('returns the original balance for a zero charge', () => {
    expect(giftCardBalanceAfterCharge(makeCard({ balance: 50 }), 0)).toBe(50);
  });

  it('returns the original balance for a negative charge', () => {
    expect(giftCardBalanceAfterCharge(makeCard({ balance: 50 }), -10)).toBe(50);
  });
});

describe('isValidGiftCardCode', () => {
  it('accepts a valid code', () => {
    expect(isValidGiftCardCode('GC-ABCD-1234')).toBe(true);
  });

  it('accepts a simple alphanumeric code', () => {
    expect(isValidGiftCardCode('ABCD1234')).toBe(true);
  });

  it('rejects an empty string', () => {
    expect(isValidGiftCardCode('')).toBe(false);
  });

  it('rejects a code that is too short', () => {
    expect(isValidGiftCardCode('AB')).toBe(false);
  });

  it('rejects codes with spaces', () => {
    expect(isValidGiftCardCode('GC ABCD')).toBe(false);
  });
});

describe('formatGiftCardBalance', () => {
  it('formats a positive balance', () => {
    expect(formatGiftCardBalance(50)).toBe('$50.00');
  });

  it('formats zero', () => {
    expect(formatGiftCardBalance(0)).toBe('$0.00');
  });

  it('returns "$0.00" for a negative balance', () => {
    expect(formatGiftCardBalance(-5)).toBe('$0.00');
  });

  it('returns "$0.00" for NaN', () => {
    expect(formatGiftCardBalance(NaN)).toBe('$0.00');
  });
});

describe('applyGiftCardCharge', () => {
  it('returns a new card with reduced balance', () => {
    const card = makeCard({ balance: 50 });
    const result = applyGiftCardCharge(card, 20);
    expect(result.balance).toBe(30);
  });

  it('does not mutate the original card', () => {
    const card = makeCard({ balance: 50 });
    applyGiftCardCharge(card, 20);
    expect(card.balance).toBe(50);
  });

  it('preserves all other card fields', () => {
    const card = makeCard({ code: 'MY-CODE', active: true });
    const result = applyGiftCardCharge(card, 10);
    expect(result.code).toBe('MY-CODE');
    expect(result.active).toBe(true);
  });
});
