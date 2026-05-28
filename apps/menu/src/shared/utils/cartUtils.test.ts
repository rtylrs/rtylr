import { describe, it, expect } from 'vitest';
import {
  lineSubtotal,
  cartTotal,
  cartItemCount,
  cartIsEmpty,
  findCartItem,
  toOrderLines,
} from './cartUtils';
import type { CartItem } from '../stores/cartStore';

const makeItem = (overrides: Partial<CartItem> = {}): CartItem => ({
  menuItemId: 1,
  name: 'Burger',
  unitPrice: 9.99,
  quantity: 1,
  notes: '',
  ...overrides,
});

describe('lineSubtotal', () => {
  it('multiplies unit price by quantity', () => {
    expect(lineSubtotal(makeItem({ unitPrice: 5, quantity: 3 }))).toBe(15);
  });

  it('returns 0 for a zero quantity item', () => {
    expect(lineSubtotal(makeItem({ unitPrice: 5, quantity: 0 }))).toBe(0);
  });
});

describe('cartTotal', () => {
  it('sums all line subtotals', () => {
    const items = [
      makeItem({ menuItemId: 1, unitPrice: 10, quantity: 2 }),
      makeItem({ menuItemId: 2, unitPrice: 5, quantity: 1 }),
    ];
    expect(cartTotal(items)).toBe(25);
  });

  it('returns 0 for an empty cart', () => {
    expect(cartTotal([])).toBe(0);
  });
});

describe('cartItemCount', () => {
  it('sums quantities across all items', () => {
    const items = [
      makeItem({ menuItemId: 1, quantity: 2 }),
      makeItem({ menuItemId: 2, quantity: 3 }),
    ];
    expect(cartItemCount(items)).toBe(5);
  });

  it('returns 0 for an empty cart', () => {
    expect(cartItemCount([])).toBe(0);
  });
});

describe('cartIsEmpty', () => {
  it('returns true when there are no items', () => {
    expect(cartIsEmpty([])).toBe(true);
  });

  it('returns false when there is at least one item', () => {
    expect(cartIsEmpty([makeItem()])).toBe(false);
  });
});

describe('findCartItem', () => {
  const items = [
    makeItem({ menuItemId: 1, name: 'Burger' }),
    makeItem({ menuItemId: 2, name: 'Fries' }),
  ];

  it('returns the item with the matching menuItemId', () => {
    expect(findCartItem(items, 2)?.name).toBe('Fries');
  });

  it('returns undefined when the item is not found', () => {
    expect(findCartItem(items, 99)).toBeUndefined();
  });
});

describe('toOrderLines', () => {
  it('maps cart items to API order line format', () => {
    const items = [
      makeItem({ menuItemId: 7, name: 'Shake', unitPrice: 5, quantity: 2, notes: 'no ice' }),
    ];
    const lines = toOrderLines(items);
    expect(lines).toEqual([
      { menu_item_id: 7, name: 'Shake', quantity: 2, unit_price: 5, notes: 'no ice' },
    ]);
  });

  it('returns an empty array for an empty cart', () => {
    expect(toOrderLines([])).toEqual([]);
  });
});
