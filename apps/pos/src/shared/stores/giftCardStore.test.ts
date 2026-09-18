import { describe, it, expect, beforeEach } from 'vitest';
import { useGiftCardStore } from './giftCardStore';

const { getState, setState } = useGiftCardStore;
const act = (fn: () => void) => {
  fn();
};

beforeEach(() => {
  setState({ cards: [] });
});

describe('giftCardStore', () => {
  // ── Initial state ──────────────────────────────────────────────

  it('starts with no cards', () => {
    expect(getState().cards).toEqual([]);
  });

  // ── issue ──────────────────────────────────────────────────────

  it('issues a new gift card', () => {
    const card = getState().issue({ code: 'GC-100', balance: 100 });

    expect(card.id).toMatch(/^gc_/);
    expect(card.code).toBe('GC-100');
    expect(card.balance).toBe(100);
    expect(card.active).toBe(true);
    expect(card.issuedAt).toBeTruthy();
    expect(getState().cards).toHaveLength(1);
  });

  it('issues card with customer name', () => {
    const card = getState().issue({ code: 'GC-200', balance: 50, customerName: 'Jane' });
    expect(card.customerName).toBe('Jane');
  });

  it('rejects issuing a card with non-positive balance', () => {
    expect(() => getState().issue({ code: 'GC-BAD', balance: 0 })).toThrow(
      'Gift card amount must be greater than zero'
    );
    expect(() => getState().issue({ code: 'GC-BAD', balance: -10 })).toThrow(
      'Gift card amount must be greater than zero'
    );
    expect(getState().cards).toHaveLength(0);
  });

  it('prepends new cards (newest first)', () => {
    getState().issue({ code: 'FIRST', balance: 10 });
    getState().issue({ code: 'SECOND', balance: 20 });
    expect(getState().cards[0].code).toBe('SECOND');
  });

  it('rejects issuing a duplicate card code', () => {
    getState().issue({ code: 'GC-DUP', balance: 25 });

    expect(() => getState().issue({ code: 'GC-DUP', balance: 50 })).toThrow(
      'Gift card code already exists'
    );
    expect(getState().cards).toHaveLength(1);
    expect(getState().findByCode('GC-DUP')?.balance).toBe(25);
  });

  // ── redeem ─────────────────────────────────────────────────────

  it('redeems from a gift card successfully', () => {
    getState().issue({ code: 'GC-50', balance: 50 });
    const result = getState().redeem('GC-50', 30);

    expect(result.ok).toBe(true);
    expect(result.remaining).toBe(20);
    expect(getState().findByCode('GC-50')?.balance).toBe(20);
  });

  it('redeems the full balance', () => {
    getState().issue({ code: 'GC-FULL', balance: 100 });
    const result = getState().redeem('GC-FULL', 100);

    expect(result.ok).toBe(true);
    expect(result.remaining).toBe(0);
    expect(getState().findByCode('GC-FULL')?.balance).toBe(0);
  });

  it('fails to redeem from non-existent card', () => {
    const result = getState().redeem('NOPE', 10);
    expect(result.ok).toBe(false);
    expect(result.error).toBe('Card not found');
  });

  it('fails to redeem from inactive card', () => {
    getState().issue({ code: 'GC-INACTIVE', balance: 100 });
    act(() => getState().deactivate('GC-INACTIVE'));

    const result = getState().redeem('GC-INACTIVE', 10);
    expect(result.ok).toBe(false);
    expect(result.error).toBe('Card is inactive');
  });

  it('fails to redeem more than balance', () => {
    getState().issue({ code: 'GC-LOW', balance: 20 });
    const result = getState().redeem('GC-LOW', 50);

    expect(result.ok).toBe(false);
    expect(result.error).toBe('Insufficient balance');
  });

  it('fails to redeem a non-positive amount', () => {
    getState().issue({ code: 'GC-VALIDATE', balance: 20 });

    const zero = getState().redeem('GC-VALIDATE', 0);
    const negative = getState().redeem('GC-VALIDATE', -10);

    expect(zero).toEqual({ ok: false, error: 'Redeem amount must be greater than zero' });
    expect(negative).toEqual({ ok: false, error: 'Redeem amount must be greater than zero' });
    expect(getState().findByCode('GC-VALIDATE')?.balance).toBe(20);
  });

  // ── addBalance ─────────────────────────────────────────────────

  it('adds balance to a gift card', () => {
    getState().issue({ code: 'GC-ADD', balance: 50 });
    act(() => getState().addBalance('GC-ADD', 25));
    expect(getState().findByCode('GC-ADD')?.balance).toBe(75);
  });

  it('rejects adding zero balance', () => {
    getState().issue({ code: 'GC-ZERO', balance: 100 });
    expect(() => getState().addBalance('GC-ZERO', 0)).toThrow(
      'Gift card amount must be greater than zero'
    );
    expect(getState().findByCode('GC-ZERO')?.balance).toBe(100);
  });

  it('rejects adding a negative balance', () => {
    getState().issue({ code: 'GC-ADD-BAD', balance: 100 });

    expect(() => getState().addBalance('GC-ADD-BAD', -25)).toThrow(
      'Gift card amount must be greater than zero'
    );
    expect(getState().findByCode('GC-ADD-BAD')?.balance).toBe(100);
  });

  // ── deactivate ─────────────────────────────────────────────────

  it('deactivates a gift card', () => {
    getState().issue({ code: 'GC-DEACT', balance: 100 });
    act(() => getState().deactivate('GC-DEACT'));

    const card = getState().findByCode('GC-DEACT');
    expect(card?.active).toBe(false);
    expect(card?.balance).toBe(100); // balance preserved
  });

  // ── findByCode ─────────────────────────────────────────────────

  it('finds a card by code', () => {
    getState().issue({ code: 'GC-FIND', balance: 75 });
    const card = getState().findByCode('GC-FIND');
    expect(card).toBeDefined();
    expect(card?.code).toBe('GC-FIND');
  });

  it('returns undefined for unknown code', () => {
    expect(getState().findByCode('UNKNOWN')).toBeUndefined();
  });

  // ── Edge cases ─────────────────────────────────────────────────

  it('multiple redeems drain the card', () => {
    getState().issue({ code: 'GC-DRAIN', balance: 100 });
    getState().redeem('GC-DRAIN', 30);
    getState().redeem('GC-DRAIN', 30);
    getState().redeem('GC-DRAIN', 30);

    expect(getState().findByCode('GC-DRAIN')?.balance).toBe(10);

    const result = getState().redeem('GC-DRAIN', 20);
    expect(result.ok).toBe(false);
    expect(result.error).toBe('Insufficient balance');
  });

  it('deactivation does not affect other cards', () => {
    getState().issue({ code: 'A', balance: 10 });
    getState().issue({ code: 'B', balance: 20 });
    act(() => getState().deactivate('A'));

    expect(getState().findByCode('A')?.active).toBe(false);
    expect(getState().findByCode('B')?.active).toBe(true);
  });
});
