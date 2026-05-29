import { describe, it, expect } from 'vitest';
import {
  totalGeoRevenue,
  topRegionByRevenue,
  regionRevenueShare,
  sortRegionsByRevenue,
  avgOrderValueByRegion,
  regionCodes,
} from './geo';
import type { GeoRegion } from './geo';

function makeRegion(code: string, revenue: number, orders: number): GeoRegion {
  return { code, name: code, revenue, orders };
}

const regions = [makeRegion('US', 6000, 120), makeRegion('UK', 3000, 60), makeRegion('AU', 1000, 20)];

describe('totalGeoRevenue', () => {
  it('sums revenue across regions', () => {
    expect(totalGeoRevenue(regions)).toBe(10000);
  });

  it('returns 0 for empty array', () => {
    expect(totalGeoRevenue([])).toBe(0);
  });
});

describe('topRegionByRevenue', () => {
  it('returns the highest revenue region', () => {
    expect(topRegionByRevenue(regions)?.code).toBe('US');
  });

  it('returns undefined for empty array', () => {
    expect(topRegionByRevenue([])).toBeUndefined();
  });
});

describe('regionRevenueShare', () => {
  it('returns percentage of total revenue', () => {
    expect(regionRevenueShare(regions[0], regions)).toBe(60);
  });

  it('returns 0 when total is 0', () => {
    const zero = [makeRegion('X', 0, 0)];
    expect(regionRevenueShare(zero[0], zero)).toBe(0);
  });
});

describe('sortRegionsByRevenue', () => {
  it('sorts descending by revenue', () => {
    const sorted = sortRegionsByRevenue(regions);
    expect(sorted[0].code).toBe('US');
    expect(sorted[2].code).toBe('AU');
  });

  it('does not mutate original', () => {
    sortRegionsByRevenue(regions);
    expect(regions[0].code).toBe('US');
  });
});

describe('avgOrderValueByRegion', () => {
  it('calculates average order value', () => {
    expect(avgOrderValueByRegion(regions[0])).toBe(50);
  });

  it('returns 0 when orders is 0', () => {
    expect(avgOrderValueByRegion(makeRegion('X', 100, 0))).toBe(0);
  });
});

describe('regionCodes', () => {
  it('extracts codes from regions', () => {
    expect(regionCodes(regions)).toEqual(['US', 'UK', 'AU']);
  });
});
