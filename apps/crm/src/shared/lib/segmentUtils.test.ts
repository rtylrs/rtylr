import { describe, it, expect } from 'vitest';
import {
  sortBySize,
  getTrendingSegments,
  totalCustomers,
  largestSegment,
  segmentNameExists,
  getSegmentsWithRules,
} from './segmentUtils';
import type { Segment } from '../schemas/segments';

const makeSegment = (overrides: Partial<Segment> = {}): Segment => ({
  id: 1,
  name: 'VIP',
  description: 'Top customers',
  customerCount: 100,
  color: '#3b82f6',
  rules: [],
  trend: '+5%',
  trendUp: true,
  ...overrides,
});

const segments: Segment[] = [
  makeSegment({ id: 1, name: 'VIP', customerCount: 120, trendUp: true, rules: ['spend > 500'] }),
  makeSegment({ id: 2, name: 'Regular', customerCount: 350, trendUp: false, rules: [] }),
  makeSegment({
    id: 3,
    name: 'At-Risk',
    customerCount: 80,
    trendUp: false,
    rules: ['days_inactive > 30'],
  }),
  makeSegment({ id: 4, name: 'New', customerCount: 45, trendUp: true, rules: ['age_days < 30'] }),
];

describe('sortBySize', () => {
  it('sorts segments by customer count, largest first', () => {
    const sorted = sortBySize(segments);
    expect(sorted[0].name).toBe('Regular');
    expect(sorted[sorted.length - 1].name).toBe('New');
  });

  it('does not mutate the original array', () => {
    const original = [...segments];
    sortBySize(segments);
    expect(segments[0]).toEqual(original[0]);
  });
});

describe('getTrendingSegments', () => {
  it('returns only segments with trendUp true', () => {
    const result = getTrendingSegments(segments);
    expect(result).toHaveLength(2);
    expect(result.every((s) => s.trendUp)).toBe(true);
  });
});

describe('totalCustomers', () => {
  it('sums all customer counts', () => {
    expect(totalCustomers(segments)).toBe(595);
  });

  it('returns 0 for empty array', () => {
    expect(totalCustomers([])).toBe(0);
  });
});

describe('largestSegment', () => {
  it('returns the segment with the highest customer count', () => {
    expect(largestSegment(segments)?.name).toBe('Regular');
  });

  it('returns null for empty array', () => {
    expect(largestSegment([])).toBeNull();
  });
});

describe('segmentNameExists', () => {
  it('returns true for an existing name (exact match)', () => {
    expect(segmentNameExists(segments, 'VIP')).toBe(true);
  });

  it('is case-insensitive', () => {
    expect(segmentNameExists(segments, 'vip')).toBe(true);
    expect(segmentNameExists(segments, 'REGULAR')).toBe(true);
  });

  it('returns false for a non-existing name', () => {
    expect(segmentNameExists(segments, 'Platinum')).toBe(false);
  });
});

describe('getSegmentsWithRules', () => {
  it('returns only segments that have at least one rule', () => {
    const result = getSegmentsWithRules(segments);
    expect(result).toHaveLength(3);
    expect(result.every((s) => s.rules.length > 0)).toBe(true);
  });

  it('returns empty array when no segments have rules', () => {
    const noRules = [makeSegment({ rules: [] })];
    expect(getSegmentsWithRules(noRules)).toHaveLength(0);
  });
});
