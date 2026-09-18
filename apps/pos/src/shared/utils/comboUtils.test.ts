import { describe, it, expect } from 'vitest';
import {
  comboItemCount,
  pricePerItem,
  comboContainsItem,
  comboItemNames,
  comboLabel,
  sortCombosByPrice,
  combosContainingItem,
} from './comboUtils';
import type { Combo } from '../types/pos';

const makeCombo = (overrides: Partial<Combo> = {}): Combo => ({
  id: 'c1',
  name: 'Meal Deal',
  price: 12,
  items: [
    { itemId: 'burger', qty: 1, name: 'Burger' },
    { itemId: 'fries', qty: 1, name: 'Fries' },
  ],
  ...overrides,
});

describe('comboItemCount', () => {
  it('returns the total quantity across all combo items', () => {
    expect(comboItemCount(makeCombo())).toBe(2);
  });

  it('returns 0 for an empty items list', () => {
    expect(comboItemCount(makeCombo({ items: [] }))).toBe(0);
  });

  it('sums multiple quantities', () => {
    const combo = makeCombo({
      items: [{ itemId: 'a', qty: 2, name: 'A' }, { itemId: 'b', qty: 3, name: 'B' }],
    });
    expect(comboItemCount(combo)).toBe(5);
  });
});

describe('pricePerItem', () => {
  it('divides combo price by total item count', () => {
    expect(pricePerItem(makeCombo({ price: 12, items: [{ itemId: 'a', qty: 2, name: 'A' }] }))).toBe(6);
  });

  it('returns 0 when there are no items', () => {
    expect(pricePerItem(makeCombo({ items: [] }))).toBe(0);
  });

  it('returns 0 for a non-finite price', () => {
    expect(pricePerItem(makeCombo({ price: NaN }))).toBe(0);
  });
});

describe('comboContainsItem', () => {
  it('returns true when the combo contains the item', () => {
    expect(comboContainsItem(makeCombo(), 'burger')).toBe(true);
  });

  it('returns false when the combo does not contain the item', () => {
    expect(comboContainsItem(makeCombo(), 'soda')).toBe(false);
  });
});

describe('comboItemNames', () => {
  it('returns comma-separated item names', () => {
    expect(comboItemNames(makeCombo())).toBe('Burger, Fries');
  });

  it('falls back to itemId when name is not set', () => {
    const combo = makeCombo({ items: [{ itemId: 'item-x', qty: 1 }] });
    expect(comboItemNames(combo)).toBe('item-x');
  });
});

describe('comboLabel', () => {
  it('formats name and price', () => {
    expect(comboLabel(makeCombo({ name: 'Meal Deal', price: 12 }))).toBe('Meal Deal - $12.00');
  });
});

describe('sortCombosByPrice', () => {
  it('sorts combos from lowest to highest price', () => {
    const combos = [makeCombo({ id: 'b', price: 15 }), makeCombo({ id: 'a', price: 10 })];
    const sorted = sortCombosByPrice(combos);
    expect(sorted[0].id).toBe('a');
    expect(sorted[1].id).toBe('b');
  });

  it('does not mutate the original array', () => {
    const combos = [makeCombo({ id: 'b', price: 15 }), makeCombo({ id: 'a', price: 10 })];
    sortCombosByPrice(combos);
    expect(combos[0].id).toBe('b');
  });
});

describe('combosContainingItem', () => {
  it('returns combos that include the specified item', () => {
    const combos = [
      makeCombo({ id: 'a', items: [{ itemId: 'burger', qty: 1, name: 'Burger' }] }),
      makeCombo({ id: 'b', items: [{ itemId: 'soda', qty: 1, name: 'Soda' }] }),
    ];
    const result = combosContainingItem(combos, 'burger');
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe('a');
  });
});
