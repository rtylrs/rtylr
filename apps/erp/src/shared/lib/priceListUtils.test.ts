import { describe, it, expect } from 'vitest';
import {
  applyPricingRule,
  discountAmount,
  discountPct,
  qualifiesForEntry,
  bestPrice,
  groupByProduct,
  sortByPrice,
} from './priceListUtils';
import type { PriceListEntry, PricingRule } from './priceListUtils';

const makeEntry = (productId: number, listPrice: number, minQuantity?: number): PriceListEntry => ({
  productId,
  sku: `SKU-${productId}`,
  basePrice: 100,
  listPrice,
  currency: 'USD',
  minQuantity,
});

describe('applyPricingRule', () => {
  it('applies percentage discount', () => {
    const rule: PricingRule = { type: 'percentage', value: 10 };
    expect(applyPricingRule(100, rule)).toBe(90);
  });

  it('applies fixed discount', () => {
    const rule: PricingRule = { type: 'fixed', value: 15 };
    expect(applyPricingRule(100, rule)).toBe(85);
  });

  it('applies multiplier', () => {
    const rule: PricingRule = { type: 'multiplier', value: 0.8 };
    expect(applyPricingRule(100, rule)).toBe(80);
  });
});

describe('discountAmount', () => {
  it('computes correct discount amount', () => {
    expect(discountAmount(100, 80)).toBe(20);
  });

  it('returns 0 when list price is higher than base', () => {
    expect(discountAmount(80, 100)).toBe(0);
  });
});

describe('discountPct', () => {
  it('computes percentage discount', () => {
    expect(discountPct(100, 80)).toBe(20);
  });

  it('returns 0 when base price is 0', () => {
    expect(discountPct(0, 50)).toBe(0);
  });
});

describe('qualifiesForEntry', () => {
  it('qualifies when quantity meets minQuantity', () => {
    expect(qualifiesForEntry(makeEntry(1, 90, 5), 5)).toBe(true);
    expect(qualifiesForEntry(makeEntry(1, 90, 5), 10)).toBe(true);
  });

  it('does not qualify below minQuantity', () => {
    expect(qualifiesForEntry(makeEntry(1, 90, 5), 4)).toBe(false);
  });

  it('qualifies with any quantity when minQuantity is undefined', () => {
    expect(qualifiesForEntry(makeEntry(1, 90), 1)).toBe(true);
  });
});

describe('bestPrice', () => {
  it('returns lowest qualifying price', () => {
    const entries = [makeEntry(1, 90, 1), makeEntry(1, 80, 10), makeEntry(1, 75, 100)];
    expect(bestPrice(entries, 1, 10)).toBe(80);
    expect(bestPrice(entries, 1, 1)).toBe(90);
  });

  it('returns 0 when no qualifying entries', () => {
    const entries = [makeEntry(1, 80, 100)];
    expect(bestPrice(entries, 1, 5)).toBe(0);
  });
});

describe('groupByProduct', () => {
  it('groups entries by product id', () => {
    const entries = [makeEntry(1, 90), makeEntry(1, 80), makeEntry(2, 50)];
    const groups = groupByProduct(entries);
    expect(groups.get(1)!).toHaveLength(2);
    expect(groups.get(2)!).toHaveLength(1);
  });
});

describe('sortByPrice', () => {
  it('sorts entries by list price ascending', () => {
    const entries = [makeEntry(1, 100), makeEntry(2, 50), makeEntry(3, 75)];
    const sorted = sortByPrice(entries);
    expect(sorted[0].listPrice).toBe(50);
    expect(sorted[2].listPrice).toBe(100);
  });
});
