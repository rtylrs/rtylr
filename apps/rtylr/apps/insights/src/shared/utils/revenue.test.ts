import { describe, it, expect } from 'vitest';
import {
  totalRevenue,
  avgDailyRevenue,
  peakRevenueDay,
  revenueSeriesLabels,
  revenueSeriesValues,
  weeklyTotals,
} from './revenue';
import type { RevenuePoint } from '../hooks/queries/types';

function makePoint(date: string, amount: number): RevenuePoint {
  return { date, amountMinorUnits: amount };
}

describe('totalRevenue', () => {
  it('sums all amountMinorUnits', () => {
    expect(totalRevenue([makePoint('2026-01-01', 100), makePoint('2026-01-02', 200)])).toBe(300);
  });

  it('returns 0 for empty series', () => {
    expect(totalRevenue([])).toBe(0);
  });
});

describe('avgDailyRevenue', () => {
  it('returns average per point', () => {
    const series = [makePoint('2026-01-01', 100), makePoint('2026-01-02', 300)];
    expect(avgDailyRevenue(series)).toBe(200);
  });

  it('returns 0 for empty series', () => {
    expect(avgDailyRevenue([])).toBe(0);
  });
});

describe('peakRevenueDay', () => {
  it('returns point with highest amount', () => {
    const series = [makePoint('2026-01-01', 100), makePoint('2026-01-02', 500)];
    expect(peakRevenueDay(series)?.date).toBe('2026-01-02');
  });

  it('returns undefined for empty series', () => {
    expect(peakRevenueDay([])).toBeUndefined();
  });
});

describe('revenueSeriesLabels', () => {
  it('extracts date strings', () => {
    const series = [makePoint('2026-01-01', 0), makePoint('2026-01-02', 0)];
    expect(revenueSeriesLabels(series)).toEqual(['2026-01-01', '2026-01-02']);
  });
});

describe('revenueSeriesValues', () => {
  it('extracts amountMinorUnits', () => {
    const series = [makePoint('2026-01-01', 100), makePoint('2026-01-02', 200)];
    expect(revenueSeriesValues(series)).toEqual([100, 200]);
  });
});

describe('weeklyTotals', () => {
  it('groups into 7-day chunks and sums each', () => {
    const series = Array.from({ length: 14 }, (_, i) =>
      makePoint(`2026-01-${String(i + 1).padStart(2, '0')}`, 10)
    );
    expect(weeklyTotals(series)).toEqual([70, 70]);
  });

  it('includes partial week as last group', () => {
    const series = Array.from({ length: 8 }, (_, i) =>
      makePoint(`2026-01-${String(i + 1).padStart(2, '0')}`, 10)
    );
    expect(weeklyTotals(series)).toEqual([70, 10]);
  });

  it('returns empty for empty series', () => {
    expect(weeklyTotals([])).toEqual([]);
  });
});
