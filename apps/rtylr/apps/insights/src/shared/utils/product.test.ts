import { describe, it, expect } from 'vitest';
import {
  topSellersByRevenue,
  topSellersByUnits,
  avgRevenuePerUnit,
  totalProductsInMarginBuckets,
  highMarginBuckets,
  slowMoversByStock,
  totalSlowMoverStock,
} from './product';
import type { BestSeller, MarginBucket, SlowMover } from '../hooks/queries/types';

function makeSeller(overrides: Partial<BestSeller> = {}): BestSeller {
  return { productId: 1, name: 'Widget', unitsSold: 100, revenueMinorUnits: 10000, ...overrides };
}

function makeBucket(overrides: Partial<MarginBucket> = {}): MarginBucket {
  return { label: '0-20%', lowerPct: 0, upperPct: 20, productCount: 10, ...overrides };
}

function makeSlow(overrides: Partial<SlowMover> = {}): SlowMover {
  return { productId: 1, name: 'Old Widget', stockQuantity: 50, ...overrides };
}

describe('topSellersByRevenue', () => {
  it('sorts by revenue and slices to n', () => {
    const products = [
      makeSeller({ productId: 1, revenueMinorUnits: 1000 }),
      makeSeller({ productId: 2, revenueMinorUnits: 5000 }),
      makeSeller({ productId: 3, revenueMinorUnits: 3000 }),
    ];
    const result = topSellersByRevenue(products, 2);
    expect(result[0].productId).toBe(2);
    expect(result).toHaveLength(2);
  });

  it('does not mutate the original array', () => {
    const products = [makeSeller()];
    topSellersByRevenue(products, 1);
    expect(products[0].productId).toBe(1);
  });
});

describe('topSellersByUnits', () => {
  it('sorts by unitsSold descending', () => {
    const products = [
      makeSeller({ productId: 1, unitsSold: 50 }),
      makeSeller({ productId: 2, unitsSold: 200 }),
    ];
    expect(topSellersByUnits(products, 1)[0].productId).toBe(2);
  });
});

describe('avgRevenuePerUnit', () => {
  it('returns revenue per unit sold', () => {
    const product = makeSeller({ revenueMinorUnits: 10000, unitsSold: 100 });
    expect(avgRevenuePerUnit(product)).toBe(100);
  });

  it('returns 0 when units sold is 0', () => {
    expect(avgRevenuePerUnit(makeSeller({ unitsSold: 0 }))).toBe(0);
  });
});

describe('totalProductsInMarginBuckets', () => {
  it('sums productCount across buckets', () => {
    const buckets = [makeBucket({ productCount: 5 }), makeBucket({ productCount: 10 })];
    expect(totalProductsInMarginBuckets(buckets)).toBe(15);
  });

  it('returns 0 for empty input', () => {
    expect(totalProductsInMarginBuckets([])).toBe(0);
  });
});

describe('highMarginBuckets', () => {
  it('filters buckets at or above threshold', () => {
    const buckets = [
      makeBucket({ lowerPct: 0 }),
      makeBucket({ lowerPct: 30 }),
      makeBucket({ lowerPct: 50 }),
    ];
    const result = highMarginBuckets(buckets, 30);
    expect(result).toHaveLength(2);
  });
});

describe('slowMoversByStock', () => {
  it('sorts by stockQuantity descending', () => {
    const products = [
      makeSlow({ productId: 1, stockQuantity: 10 }),
      makeSlow({ productId: 2, stockQuantity: 50 }),
    ];
    expect(slowMoversByStock(products)[0].productId).toBe(2);
  });
});

describe('totalSlowMoverStock', () => {
  it('sums stockQuantity across all slow movers', () => {
    const products = [makeSlow({ stockQuantity: 10 }), makeSlow({ stockQuantity: 20 })];
    expect(totalSlowMoverStock(products)).toBe(30);
  });
});
