import { describe, it, expect } from 'vitest';
import {
  totalSegmentRevenue,
  totalSegmentCount,
  segmentRevenueShare,
  topSegmentByRevenue,
  sortSegmentsByRevenue,
  avgRevenuePerCustomer,
} from './segment';
import type { CustomerSegment } from './segment';

function makeSeg(id: string, count: number, revenue: number): CustomerSegment {
  return { id, name: id, count, revenue };
}

const segments = [makeSeg('A', 100, 5000), makeSeg('B', 200, 3000), makeSeg('C', 50, 2000)];

describe('totalSegmentRevenue', () => {
  it('sums revenue across segments', () => {
    expect(totalSegmentRevenue(segments)).toBe(10000);
  });

  it('returns 0 for empty array', () => {
    expect(totalSegmentRevenue([])).toBe(0);
  });
});

describe('totalSegmentCount', () => {
  it('sums counts across segments', () => {
    expect(totalSegmentCount(segments)).toBe(350);
  });
});

describe('segmentRevenueShare', () => {
  it('returns percentage share of total', () => {
    expect(segmentRevenueShare(segments[0], segments)).toBe(50);
  });

  it('returns 0 when total revenue is 0', () => {
    const zero = [makeSeg('X', 10, 0)];
    expect(segmentRevenueShare(zero[0], zero)).toBe(0);
  });
});

describe('topSegmentByRevenue', () => {
  it('returns highest revenue segment', () => {
    expect(topSegmentByRevenue(segments)?.id).toBe('A');
  });

  it('returns undefined for empty array', () => {
    expect(topSegmentByRevenue([])).toBeUndefined();
  });
});

describe('sortSegmentsByRevenue', () => {
  it('sorts descending by revenue', () => {
    const sorted = sortSegmentsByRevenue(segments);
    expect(sorted[0].id).toBe('A');
    expect(sorted[2].id).toBe('C');
  });

  it('does not mutate original', () => {
    sortSegmentsByRevenue(segments);
    expect(segments[0].id).toBe('A');
  });
});

describe('avgRevenuePerCustomer', () => {
  it('calculates average revenue per customer', () => {
    expect(avgRevenuePerCustomer(segments)).toBeCloseTo(28.57, 1);
  });

  it('returns 0 when count is 0', () => {
    expect(avgRevenuePerCustomer([])).toBe(0);
  });
});
