import { describe, it, expect } from 'vitest';
import {
  totalAppRevenue,
  topApps,
  appLabels,
  appValues,
  findAppShare,
  appPctLabels,
} from './appRevenue';
import type { AppRevenueShare } from '../hooks/queries/types';

function makeShare(app: string, amount: number, pct = 0): AppRevenueShare {
  return { app, amountMinorUnits: amount, pct };
}

describe('totalAppRevenue', () => {
  it('sums all amounts', () => {
    const shares = [makeShare('pos', 5000), makeShare('menu', 2000, 20)];
    expect(totalAppRevenue(shares)).toBe(7000);
  });

  it('returns 0 for empty array', () => {
    expect(totalAppRevenue([])).toBe(0);
  });
});

describe('topApps', () => {
  it('returns top n by revenue', () => {
    const shares = [makeShare('a', 100, 10), makeShare('b', 500, 50), makeShare('c', 200, 20)];
    const top = topApps(shares, 2);
    expect(top[0].app).toBe('b');
    expect(top).toHaveLength(2);
  });
});

describe('appLabels', () => {
  it('extracts app names', () => {
    const shares = [makeShare('pos', 100, 50), makeShare('menu', 100, 50)];
    expect(appLabels(shares)).toEqual(['pos', 'menu']);
  });
});

describe('appValues', () => {
  it('extracts amounts', () => {
    const shares = [makeShare('pos', 100, 50), makeShare('menu', 200, 50)];
    expect(appValues(shares)).toEqual([100, 200]);
  });
});

describe('findAppShare', () => {
  it('finds a share by app name', () => {
    const shares = [makeShare('pos', 100, 50), makeShare('menu', 200, 50)];
    expect(findAppShare(shares, 'menu')?.app).toBe('menu');
  });

  it('returns undefined when not found', () => {
    expect(findAppShare([], 'any')).toBeUndefined();
  });
});

describe('appPctLabels', () => {
  it('formats pct as percentage string', () => {
    const shares = [makeShare('pos', 100, 33.3), makeShare('menu', 200, 66.7)];
    expect(appPctLabels(shares)).toEqual(['33.3%', '66.7%']);
  });
});
