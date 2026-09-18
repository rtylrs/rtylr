import { describe, it, expect } from 'vitest';
import {
  netSales,
  averageTicket,
  refundRate,
  voidRate,
  tipRate,
  effectiveTaxRate,
  sumField,
  bestDay,
  worstDay,
  aggregateSummaries,
} from './reportingUtils';
import type { DaySummary } from '../types/pos';

const makeDay = (overrides: Partial<DaySummary> = {}): DaySummary => ({
  date: '2024-06-01',
  grossSales: 1000,
  netSales: 950,
  taxCollected: 85,
  tips: 100,
  refunds: 50,
  voids: 20,
  orderCount: 40,
  averageTicket: 25,
  cashDrops: 200,
  openingFloat: 100,
  closingCash: 350,
  ...overrides,
});

describe('netSales', () => {
  it('subtracts refunds from gross sales', () => {
    expect(netSales(1000, 50)).toBe(950);
  });

  it('floors at 0 when refunds exceed gross sales', () => {
    expect(netSales(0, 100)).toBe(0);
  });

  it('returns 0 for non-finite inputs', () => {
    expect(netSales(NaN, 50)).toBe(0);
    expect(netSales(1000, NaN)).toBe(0);
  });
});

describe('averageTicket', () => {
  it('divides total sales by order count', () => {
    expect(averageTicket(1000, 40)).toBeCloseTo(25);
  });

  it('returns 0 when orderCount is 0', () => {
    expect(averageTicket(1000, 0)).toBe(0);
  });

  it('returns 0 for non-finite totalSales', () => {
    expect(averageTicket(NaN, 10)).toBe(0);
  });
});

describe('refundRate', () => {
  it('calculates refund rate as percentage', () => {
    expect(refundRate(50, 1000)).toBe(5);
  });

  it('returns 0 when grossSales is 0', () => {
    expect(refundRate(50, 0)).toBe(0);
  });

  it('returns 0 for non-finite inputs', () => {
    expect(refundRate(NaN, 1000)).toBe(0);
  });
});

describe('voidRate', () => {
  it('calculates void rate as percentage', () => {
    expect(voidRate(20, 1000)).toBe(2);
  });

  it('returns 0 for 0 gross sales', () => {
    expect(voidRate(20, 0)).toBe(0);
  });
});

describe('tipRate', () => {
  it('calculates tip rate against net sales', () => {
    expect(tipRate(100, 1000)).toBe(10);
  });

  it('returns 0 for 0 net sales', () => {
    expect(tipRate(100, 0)).toBe(0);
  });
});

describe('effectiveTaxRate', () => {
  it('calculates tax as percentage of net sales', () => {
    expect(effectiveTaxRate(85, 1000)).toBeCloseTo(8.5);
  });

  it('returns 0 for 0 net sales', () => {
    expect(effectiveTaxRate(85, 0)).toBe(0);
  });
});

describe('sumField', () => {
  it('sums a numeric field across summaries', () => {
    const days = [makeDay({ grossSales: 500 }), makeDay({ grossSales: 700 })];
    expect(sumField(days, 'grossSales')).toBe(1200);
  });

  it('returns 0 for an empty array', () => {
    expect(sumField([], 'grossSales')).toBe(0);
  });
});

describe('bestDay', () => {
  it('returns the day with the highest gross sales', () => {
    const days = [makeDay({ date: 'A', grossSales: 500 }), makeDay({ date: 'B', grossSales: 800 })];
    expect(bestDay(days)?.date).toBe('B');
  });

  it('returns undefined for an empty array', () => {
    expect(bestDay([])).toBeUndefined();
  });
});

describe('worstDay', () => {
  it('returns the day with the lowest gross sales', () => {
    const days = [makeDay({ date: 'A', grossSales: 500 }), makeDay({ date: 'B', grossSales: 800 })];
    expect(worstDay(days)?.date).toBe('A');
  });

  it('returns undefined for an empty array', () => {
    expect(worstDay([])).toBeUndefined();
  });
});

describe('aggregateSummaries', () => {
  it('sums numeric fields across all days', () => {
    const days = [
      makeDay({ grossSales: 500, refunds: 25, orderCount: 20 }),
      makeDay({ grossSales: 700, refunds: 35, orderCount: 28 }),
    ];
    const agg = aggregateSummaries(days);
    expect(agg.grossSales).toBe(1200);
    expect(agg.refunds).toBe(60);
    expect(agg.netSales).toBe(1140);
    expect(agg.orderCount).toBe(48);
    expect(agg.averageTicket).toBeCloseTo(25);
  });

  it('returns zero values for an empty array', () => {
    const agg = aggregateSummaries([]);
    expect(agg.grossSales).toBe(0);
    expect(agg.orderCount).toBe(0);
    expect(agg.openingFloat).toBe(0);
  });
});
