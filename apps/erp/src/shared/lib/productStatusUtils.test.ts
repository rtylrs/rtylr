import { describe, it, expect } from 'vitest';
import {
  stockStatus,
  stockStatusLabel,
  stockStatusColor,
  grossMarginPct,
  profitabilityBand,
  filterByStockStatus,
  needsReorder,
  totalInventoryValue,
  potentialRevenue,
  topByValue,
} from './productStatusUtils';
import type { Product } from '../services/erpApi';

const makeProduct = (overrides: Partial<Product> = {}): Product => ({
  id: 1,
  name: 'Widget',
  sku: 'WGT-001',
  category: 'Hardware',
  stock: 50,
  lowStock: 10,
  costPrice: 5,
  salePrice: 10,
  status: 'In Stock',
  ...overrides,
});

describe('stockStatus', () => {
  it('returns out-of-stock when stock is 0', () => {
    expect(stockStatus(0, 10)).toBe('out-of-stock');
  });

  it('returns low-stock when stock is at or below threshold', () => {
    expect(stockStatus(10, 10)).toBe('low-stock');
    expect(stockStatus(5, 10)).toBe('low-stock');
  });

  it('returns in-stock when stock is above threshold', () => {
    expect(stockStatus(11, 10)).toBe('in-stock');
  });
});

describe('stockStatusLabel', () => {
  it('returns correct label for each status', () => {
    expect(stockStatusLabel('in-stock')).toBe('In Stock');
    expect(stockStatusLabel('low-stock')).toBe('Low Stock');
    expect(stockStatusLabel('out-of-stock')).toBe('Out of Stock');
  });
});

describe('stockStatusColor', () => {
  it('returns correct color for each status', () => {
    expect(stockStatusColor('in-stock')).toBe('green');
    expect(stockStatusColor('low-stock')).toBe('yellow');
    expect(stockStatusColor('out-of-stock')).toBe('red');
  });
});

describe('grossMarginPct', () => {
  it('computes gross margin percentage', () => {
    const p = makeProduct({ costPrice: 5, salePrice: 10 });
    expect(grossMarginPct(p)).toBe(50);
  });

  it('returns 0 when sale price is 0', () => {
    const p = makeProduct({ costPrice: 5, salePrice: 0 });
    expect(grossMarginPct(p)).toBe(0);
  });

  it('returns negative margin when cost exceeds sale price', () => {
    const p = makeProduct({ costPrice: 15, salePrice: 10 });
    expect(grossMarginPct(p)).toBeLessThan(0);
  });
});

describe('profitabilityBand', () => {
  it('classifies high margin correctly', () => {
    expect(profitabilityBand(makeProduct({ costPrice: 5, salePrice: 10 }))).toBe('high');
  });

  it('classifies medium margin correctly', () => {
    expect(profitabilityBand(makeProduct({ costPrice: 8, salePrice: 10 }))).toBe('medium');
  });

  it('classifies low margin correctly', () => {
    expect(profitabilityBand(makeProduct({ costPrice: 9.5, salePrice: 10 }))).toBe('low');
  });

  it('classifies loss correctly', () => {
    expect(profitabilityBand(makeProduct({ costPrice: 12, salePrice: 10 }))).toBe('loss');
  });
});

describe('filterByStockStatus', () => {
  it('filters products by stock status', () => {
    const products = [
      makeProduct({ id: 1, stock: 0, lowStock: 10 }),
      makeProduct({ id: 2, stock: 5, lowStock: 10 }),
      makeProduct({ id: 3, stock: 20, lowStock: 10 }),
    ];
    expect(filterByStockStatus(products, 'out-of-stock')).toHaveLength(1);
    expect(filterByStockStatus(products, 'low-stock')).toHaveLength(1);
    expect(filterByStockStatus(products, 'in-stock')).toHaveLength(1);
  });
});

describe('needsReorder', () => {
  it('returns products at or below reorder point', () => {
    const products = [
      makeProduct({ id: 1, stock: 5, lowStock: 10, reorderPoint: 8 }),
      makeProduct({ id: 2, stock: 20, lowStock: 10, reorderPoint: 8 }),
    ];
    const reorder = needsReorder(products);
    expect(reorder).toHaveLength(1);
    expect(reorder[0].id).toBe(1);
  });

  it('falls back to lowStock when reorderPoint is undefined', () => {
    const products = [makeProduct({ id: 1, stock: 5, lowStock: 10, reorderPoint: undefined })];
    expect(needsReorder(products)).toHaveLength(1);
  });
});

describe('totalInventoryValue', () => {
  it('computes total value at cost price', () => {
    const products = [
      makeProduct({ stock: 10, costPrice: 5 }),
      makeProduct({ stock: 20, costPrice: 3 }),
    ];
    expect(totalInventoryValue(products)).toBe(110);
  });
});

describe('potentialRevenue', () => {
  it('computes total potential revenue at sale price', () => {
    const products = [
      makeProduct({ stock: 10, salePrice: 10 }),
      makeProduct({ stock: 5, salePrice: 20 }),
    ];
    expect(potentialRevenue(products)).toBe(200);
  });
});

describe('topByValue', () => {
  it('returns top N products by inventory value', () => {
    const products = [
      makeProduct({ id: 1, stock: 2, costPrice: 100 }),
      makeProduct({ id: 2, stock: 50, costPrice: 5 }),
      makeProduct({ id: 3, stock: 10, costPrice: 20 }),
    ];
    const top2 = topByValue(products, 2);
    expect(top2).toHaveLength(2);
    expect(top2[0].id).toBe(1);
    expect(top2[1].id).toBe(3);
  });
});
