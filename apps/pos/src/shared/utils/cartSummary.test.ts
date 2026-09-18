import { describe, it, expect } from 'vitest';
import {
  totalItemCount,
  uniqueLineCount,
  itemsByStatus,
  allItemsSettled,
  discountedItems,
  grossSubtotal,
} from './cartSummary';
import type { CartItem } from '../stores/cartStore';

const makeItem = (overrides: Partial<CartItem> = {}): CartItem => ({
  lineId: '1::',
  id: '1::',
  productId: 1,
  name: 'Burger',
  unitPrice: 10,
  price: 10,
  quantity: 1,
  modifiers: [],
  discountPct: 0,
  discountAmount: 0,
  taxRate: 0,
  status: 'pending',
  ...overrides,
});

describe('totalItemCount', () => {
  it('returns 0 for an empty cart', () => {
    expect(totalItemCount([])).toBe(0);
  });

  it('sums quantities across items', () => {
    const items = [makeItem({ quantity: 2 }), makeItem({ lineId: '2::', id: '2::', quantity: 3 })];
    expect(totalItemCount(items)).toBe(5);
  });

  it('returns 1 for a single item with qty 1', () => {
    expect(totalItemCount([makeItem()])).toBe(1);
  });
});

describe('uniqueLineCount', () => {
  it('returns 0 for an empty cart', () => {
    expect(uniqueLineCount([])).toBe(0);
  });

  it('returns the number of line items regardless of quantity', () => {
    const items = [makeItem({ quantity: 5 }), makeItem({ lineId: '2::', id: '2::' })];
    expect(uniqueLineCount(items)).toBe(2);
  });
});

describe('itemsByStatus', () => {
  it('filters items by the given status', () => {
    const items = [
      makeItem({ status: 'pending' }),
      makeItem({ lineId: '2::', id: '2::', status: 'ready' }),
      makeItem({ lineId: '3::', id: '3::', status: 'pending' }),
    ];
    const pending = itemsByStatus(items, 'pending');
    expect(pending).toHaveLength(2);
  });

  it('returns an empty array when no items match', () => {
    const items = [makeItem({ status: 'pending' })];
    expect(itemsByStatus(items, 'served')).toHaveLength(0);
  });
});

describe('allItemsSettled', () => {
  it('returns false for an empty cart', () => {
    expect(allItemsSettled([])).toBe(false);
  });

  it('returns true when all items are served', () => {
    const items = [makeItem({ status: 'served' }), makeItem({ lineId: '2::', id: '2::', status: 'served' })];
    expect(allItemsSettled(items)).toBe(true);
  });

  it('returns true when items are a mix of served and cancelled', () => {
    const items = [makeItem({ status: 'served' }), makeItem({ lineId: '2::', id: '2::', status: 'cancelled' })];
    expect(allItemsSettled(items)).toBe(true);
  });

  it('returns false when at least one item is still pending', () => {
    const items = [makeItem({ status: 'served' }), makeItem({ lineId: '2::', id: '2::', status: 'pending' })];
    expect(allItemsSettled(items)).toBe(false);
  });
});

describe('discountedItems', () => {
  it('returns items with a percentage discount', () => {
    const items = [makeItem({ discountPct: 10 }), makeItem({ lineId: '2::', id: '2::' })];
    expect(discountedItems(items)).toHaveLength(1);
  });

  it('returns items with a flat discount', () => {
    const items = [makeItem({ discountAmount: 2 }), makeItem({ lineId: '2::', id: '2::' })];
    expect(discountedItems(items)).toHaveLength(1);
  });

  it('returns an empty array when no discounts are applied', () => {
    expect(discountedItems([makeItem()])).toHaveLength(0);
  });
});

describe('grossSubtotal', () => {
  it('returns 0 for an empty cart', () => {
    expect(grossSubtotal([])).toBe(0);
  });

  it('calculates unitPrice * quantity for each item', () => {
    const items = [
      makeItem({ unitPrice: 10, quantity: 2 }),
      makeItem({ lineId: '2::', id: '2::', unitPrice: 5, quantity: 3 }),
    ];
    expect(grossSubtotal(items)).toBe(35);
  });
});
