import { describe, it, expect } from 'vitest';
import {
  totalInventoryValue,
  totalInventoryQuantity,
  sortByValueDesc,
  sortByStockAsc,
  filterBySkuPrefix,
  topValueItems,
  valuePct,
  averageUnitCost,
} from './inventoryUtils';
import type { InventoryReportItem } from '../services/erpApi';

function makeItem(overrides: Partial<InventoryReportItem> = {}): InventoryReportItem {
  return {
    productId: 1,
    name: 'Widget',
    sku: 'WGT-001',
    currentStock: 100,
    value: 500,
    ...overrides,
  };
}

const items: InventoryReportItem[] = [
  makeItem({ productId: 1, name: 'A', sku: 'WGT-001', currentStock: 10, value: 100 }),
  makeItem({ productId: 2, name: 'B', sku: 'GDG-001', currentStock: 50, value: 500 }),
  makeItem({ productId: 3, name: 'C', sku: 'WGT-002', currentStock: 5, value: 250 }),
];

describe('totalInventoryValue', () => {
  it('returns sum of all values', () => {
    expect(totalInventoryValue(items)).toBe(850);
  });

  it('returns 0 for empty array', () => {
    expect(totalInventoryValue([])).toBe(0);
  });
});

describe('totalInventoryQuantity', () => {
  it('returns sum of all stock quantities', () => {
    expect(totalInventoryQuantity(items)).toBe(65);
  });

  it('returns 0 for empty array', () => {
    expect(totalInventoryQuantity([])).toBe(0);
  });
});

describe('sortByValueDesc', () => {
  it('sorts highest value first', () => {
    const sorted = sortByValueDesc(items);
    expect(sorted[0].value).toBe(500);
    expect(sorted[2].value).toBe(100);
  });

  it('does not mutate original', () => {
    sortByValueDesc(items);
    expect(items[0].value).toBe(100);
  });
});

describe('sortByStockAsc', () => {
  it('sorts lowest stock first', () => {
    const sorted = sortByStockAsc(items);
    expect(sorted[0].currentStock).toBe(5);
    expect(sorted[2].currentStock).toBe(50);
  });
});

describe('filterBySkuPrefix', () => {
  it('filters by SKU prefix', () => {
    expect(filterBySkuPrefix(items, 'WGT')).toHaveLength(2);
  });

  it('is case-insensitive', () => {
    expect(filterBySkuPrefix(items, 'wgt')).toHaveLength(2);
  });

  it('returns empty array for no matches', () => {
    expect(filterBySkuPrefix(items, 'XYZ')).toHaveLength(0);
  });
});

describe('topValueItems', () => {
  it('returns top N by value', () => {
    const top2 = topValueItems(items, 2);
    expect(top2).toHaveLength(2);
    expect(top2[0].value).toBe(500);
    expect(top2[1].value).toBe(250);
  });

  it('returns all when n exceeds array length', () => {
    expect(topValueItems(items, 10)).toHaveLength(3);
  });
});

describe('valuePct', () => {
  it('returns percentage of total value', () => {
    expect(valuePct(items[1], 1000)).toBeCloseTo(50);
  });

  it('returns 0 when total is 0', () => {
    expect(valuePct(items[0], 0)).toBe(0);
  });
});

describe('averageUnitCost', () => {
  it('returns value divided by stock', () => {
    expect(averageUnitCost(makeItem({ value: 100, currentStock: 10 }))).toBe(10);
  });

  it('returns 0 when stock is 0', () => {
    expect(averageUnitCost(makeItem({ value: 100, currentStock: 0 }))).toBe(0);
  });
});
