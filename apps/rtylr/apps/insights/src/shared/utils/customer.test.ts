import { describe, it, expect } from 'vitest';
import {
  topCustomersBySpend,
  avgOrderValue,
  totalCustomersInBuckets,
  highValueBuckets,
  lowestChurnWindow,
  churnRisk,
} from './customer';
import type { TopCustomer, LTVBucket, ChurnWindow } from '../hooks/queries/types';

function makeCustomer(overrides: Partial<TopCustomer> = {}): TopCustomer {
  return {
    customerId: 1,
    name: 'Alice',
    email: 'alice@example.com',
    lifetimeSpendMinorUnits: 10000,
    orderCount: 5,
    ...overrides,
  };
}

function makeBucket(overrides: Partial<LTVBucket> = {}): LTVBucket {
  return {
    label: '$0-$100',
    lowerMinorUnits: 0,
    upperMinorUnits: 10000,
    customerCount: 10,
    ...overrides,
  };
}

function makeWindow(overrides: Partial<ChurnWindow> = {}): ChurnWindow {
  return {
    windowDays: 30,
    churnedCount: 5,
    activeCount: 95,
    churnPct: 5,
    ...overrides,
  };
}

describe('topCustomersBySpend', () => {
  it('sorts descending by lifetimeSpendMinorUnits', () => {
    const a = makeCustomer({ customerId: 1, lifetimeSpendMinorUnits: 500 });
    const b = makeCustomer({ customerId: 2, lifetimeSpendMinorUnits: 2000 });
    const sorted = topCustomersBySpend([a, b]);
    expect(sorted[0].customerId).toBe(2);
  });

  it('returns empty array for empty input', () => {
    expect(topCustomersBySpend([])).toHaveLength(0);
  });
});

describe('avgOrderValue', () => {
  it('returns average spend per order', () => {
    const c = makeCustomer({ lifetimeSpendMinorUnits: 10000, orderCount: 4 });
    expect(avgOrderValue(c)).toBe(2500);
  });

  it('returns 0 when orderCount is 0', () => {
    expect(avgOrderValue(makeCustomer({ orderCount: 0 }))).toBe(0);
  });
});

describe('totalCustomersInBuckets', () => {
  it('sums customerCount across buckets', () => {
    const buckets = [makeBucket({ customerCount: 10 }), makeBucket({ customerCount: 20 })];
    expect(totalCustomersInBuckets(buckets)).toBe(30);
  });

  it('returns 0 for empty buckets', () => {
    expect(totalCustomersInBuckets([])).toBe(0);
  });
});

describe('highValueBuckets', () => {
  it('filters by lowerMinorUnits threshold', () => {
    const buckets = [
      makeBucket({ lowerMinorUnits: 0 }),
      makeBucket({ lowerMinorUnits: 50000 }),
      makeBucket({ lowerMinorUnits: 100000 }),
    ];
    expect(highValueBuckets(buckets, 50000)).toHaveLength(2);
  });
});

describe('lowestChurnWindow', () => {
  it('returns window with lowest churnPct', () => {
    const windows = [
      makeWindow({ churnPct: 15 }),
      makeWindow({ churnPct: 3 }),
      makeWindow({ churnPct: 8 }),
    ];
    expect(lowestChurnWindow(windows)?.churnPct).toBe(3);
  });

  it('returns undefined for empty input', () => {
    expect(lowestChurnWindow([])).toBeUndefined();
  });
});

describe('churnRisk', () => {
  it('classifies high risk at 20+', () => {
    expect(churnRisk(20)).toBe('high');
    expect(churnRisk(25)).toBe('high');
  });

  it('classifies medium risk at 10-19', () => {
    expect(churnRisk(10)).toBe('medium');
    expect(churnRisk(15)).toBe('medium');
  });

  it('classifies low risk below 10', () => {
    expect(churnRisk(5)).toBe('low');
    expect(churnRisk(0)).toBe('low');
  });
});
