import { describe, it, expect } from 'vitest';
import {
  isAvailable,
  isOutOfStock,
  isLowStock,
  isOrderable,
  orderableItems,
  itemsByCategory,
  uniqueCategories,
  priceLabel,
  itemTaxRate,
} from './menuItemUtils';
import type { MenuItem } from '@voxire/voxie/pos/menu_item/v1/types';

const makeItem = (overrides: Partial<MenuItem> = {}): MenuItem => ({
  id: 'item1',
  name: 'Burger',
  salePrice: 10,
  category: 'Mains',
  ...overrides,
});

describe('isAvailable', () => {
  it('returns true when isActive is not set', () => {
    expect(isAvailable(makeItem())).toBe(true);
  });

  it('returns true when isActive is explicitly true', () => {
    expect(isAvailable(makeItem({ isActive: true }))).toBe(true);
  });

  it('returns false when isActive is false', () => {
    expect(isAvailable(makeItem({ isActive: false }))).toBe(false);
  });
});

describe('isOutOfStock', () => {
  it('returns true when stock is 0', () => {
    expect(isOutOfStock(makeItem({ stock: 0 }))).toBe(true);
  });

  it('returns false when stock is positive', () => {
    expect(isOutOfStock(makeItem({ stock: 5 }))).toBe(false);
  });

  it('returns false when stock is not set', () => {
    expect(isOutOfStock(makeItem())).toBe(false);
  });
});

describe('isLowStock', () => {
  it('returns true when stock is within the threshold', () => {
    expect(isLowStock(makeItem({ stock: 3 }))).toBe(true);
  });

  it('returns false when stock exceeds the threshold', () => {
    expect(isLowStock(makeItem({ stock: 10 }))).toBe(false);
  });

  it('returns false when stock is 0 (out of stock, not low)', () => {
    expect(isLowStock(makeItem({ stock: 0 }))).toBe(false);
  });

  it('respects a custom threshold', () => {
    expect(isLowStock(makeItem({ stock: 8 }), 10)).toBe(true);
  });
});

describe('isOrderable', () => {
  it('returns true when item is available and in stock', () => {
    expect(isOrderable(makeItem({ isActive: true, stock: 5 }))).toBe(true);
  });

  it('returns false when item is unavailable', () => {
    expect(isOrderable(makeItem({ isActive: false }))).toBe(false);
  });

  it('returns false when item is out of stock', () => {
    expect(isOrderable(makeItem({ stock: 0 }))).toBe(false);
  });
});

describe('orderableItems', () => {
  it('filters to only orderable items', () => {
    const items = [
      makeItem({ id: 'a', isActive: true }),
      makeItem({ id: 'b', isActive: false }),
      makeItem({ id: 'c', stock: 0 }),
    ];
    const result = orderableItems(items);
    expect(result).toHaveLength(1);
    expect(result[0].id).toBe('a');
  });
});

describe('itemsByCategory', () => {
  it('groups items by their category', () => {
    const items = [
      makeItem({ id: 'a', category: 'Mains' }),
      makeItem({ id: 'b', category: 'Drinks' }),
      makeItem({ id: 'c', category: 'Mains' }),
    ];
    const grouped = itemsByCategory(items);
    expect(grouped['Mains']).toHaveLength(2);
    expect(grouped['Drinks']).toHaveLength(1);
  });
});

describe('uniqueCategories', () => {
  it('returns unique category names sorted alphabetically', () => {
    const items = [
      makeItem({ category: 'Mains' }),
      makeItem({ category: 'Drinks' }),
      makeItem({ category: 'Mains' }),
    ];
    expect(uniqueCategories(items)).toEqual(['Drinks', 'Mains']);
  });
});

describe('priceLabel', () => {
  it('formats the price to 2 decimal places', () => {
    expect(priceLabel(makeItem({ salePrice: 12.5 }))).toBe('12.50');
    expect(priceLabel(makeItem({ salePrice: 10 }))).toBe('10.00');
  });
});

describe('itemTaxRate', () => {
  it('returns the tax rate when set', () => {
    expect(itemTaxRate(makeItem({ taxRate: 0.08 }))).toBe(0.08);
  });

  it('returns 0 when taxRate is not set', () => {
    expect(itemTaxRate(makeItem())).toBe(0);
  });

  it('returns 0 for NaN tax rate', () => {
    expect(itemTaxRate(makeItem({ taxRate: NaN }))).toBe(0);
  });
});
