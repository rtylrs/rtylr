import { describe, it, expect } from 'vitest';
import {
  isLow,
  isEmpty,
  itemStatus,
  stockValue,
  reorderQty,
  groupByCategory,
  alertItems,
  qtyLabel,
} from './inventoryUtils';
import type { InventoryItem } from './inventoryUtils';

const makeItem = (overrides: Partial<InventoryItem> = {}): InventoryItem => ({
  id: 'i1',
  name: 'Flour',
  qty: 10,
  unit: 'kg',
  lowStockThreshold: 5,
  ...overrides,
});

describe('isLow', () => {
  it('returns true when qty is within the threshold', () => {
    expect(isLow(makeItem({ qty: 3 }))).toBe(true);
  });

  it('returns false when qty is above the threshold', () => {
    expect(isLow(makeItem({ qty: 10 }))).toBe(false);
  });

  it('returns false when qty is 0 (empty, not low)', () => {
    expect(isLow(makeItem({ qty: 0 }))).toBe(false);
  });
});

describe('isEmpty', () => {
  it('returns true when qty is 0', () => {
    expect(isEmpty(makeItem({ qty: 0 }))).toBe(true);
  });

  it('returns true when qty is negative', () => {
    expect(isEmpty(makeItem({ qty: -1 }))).toBe(true);
  });

  it('returns false when qty is positive', () => {
    expect(isEmpty(makeItem({ qty: 5 }))).toBe(false);
  });
});

describe('itemStatus', () => {
  it('returns "empty" when qty is 0', () => {
    expect(itemStatus(makeItem({ qty: 0 }))).toBe('empty');
  });

  it('returns "low" when qty is within the threshold', () => {
    expect(itemStatus(makeItem({ qty: 3 }))).toBe('low');
  });

  it('returns "ok" when qty is above the threshold', () => {
    expect(itemStatus(makeItem({ qty: 20 }))).toBe('ok');
  });
});

describe('stockValue', () => {
  it('multiplies cost by qty', () => {
    expect(stockValue(makeItem({ qty: 10, cost: 2.5 }))).toBeCloseTo(25);
  });

  it('returns 0 when cost is not set', () => {
    expect(stockValue(makeItem({ qty: 10 }))).toBe(0);
  });

  it('returns 0 for non-finite values', () => {
    expect(stockValue(makeItem({ qty: NaN, cost: 2.5 }))).toBe(0);
  });
});

describe('reorderQty', () => {
  it('returns the amount needed to reach target', () => {
    expect(reorderQty(makeItem({ qty: 3 }), 20)).toBe(17);
  });

  it('returns 0 when already above target', () => {
    expect(reorderQty(makeItem({ qty: 30 }), 20)).toBe(0);
  });

  it('returns 0 for non-finite target', () => {
    expect(reorderQty(makeItem({ qty: 5 }), NaN)).toBe(0);
  });
});

describe('groupByCategory', () => {
  it('groups items by category', () => {
    const items = [
      makeItem({ id: 'a', category: 'Dairy' }),
      makeItem({ id: 'b', category: 'Bakery' }),
      makeItem({ id: 'c', category: 'Dairy' }),
    ];
    const grouped = groupByCategory(items);
    expect(grouped['Dairy']).toHaveLength(2);
    expect(grouped['Bakery']).toHaveLength(1);
  });

  it('places uncategorized items under "Uncategorized"', () => {
    const items = [makeItem({ id: 'a' })];
    const grouped = groupByCategory(items);
    expect(grouped['Uncategorized']).toHaveLength(1);
  });
});

describe('alertItems', () => {
  it('returns items that are low or empty', () => {
    const items = [
      makeItem({ id: 'a', qty: 0 }),
      makeItem({ id: 'b', qty: 3 }),
      makeItem({ id: 'c', qty: 20 }),
    ];
    const alerts = alertItems(items);
    expect(alerts).toHaveLength(2);
    expect(alerts.map((i) => i.id)).toEqual(['a', 'b']);
  });
});

describe('qtyLabel', () => {
  it('returns a formatted label with name, qty, and unit', () => {
    expect(qtyLabel(makeItem({ name: 'Sugar', qty: 5, unit: 'lbs' }))).toBe('Sugar: 5 lbs');
  });
});
