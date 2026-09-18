import { describe, it, expect } from 'vitest';
import {
  summarizeDataPoints,
  periodGrowthPct,
  aggregateByPeriod,
  rollingAverage,
  formatReportValue,
} from './reportUtils';
import type { ReportDataPoint } from './reportUtils';

const pts: ReportDataPoint[] = [
  { label: 'Jan', value: 100, date: '2026-01-15' },
  { label: 'Feb', value: 200, date: '2026-02-15' },
  { label: 'Mar', value: 150, date: '2026-03-15' },
];

describe('summarizeDataPoints', () => {
  it('computes correct statistics', () => {
    const summary = summarizeDataPoints(pts);
    expect(summary.total).toBe(450);
    expect(summary.average).toBe(150);
    expect(summary.min).toBe(100);
    expect(summary.max).toBe(200);
    expect(summary.count).toBe(3);
  });

  it('detects upward trend', () => {
    const growing = [
      { label: 'a', value: 100 },
      { label: 'b', value: 200 },
    ];
    expect(summarizeDataPoints(growing).trend).toBe('up');
  });

  it('detects downward trend', () => {
    const shrinking = [
      { label: 'a', value: 200 },
      { label: 'b', value: 100 },
    ];
    expect(summarizeDataPoints(shrinking).trend).toBe('down');
  });

  it('returns flat defaults for empty array', () => {
    const summary = summarizeDataPoints([]);
    expect(summary.total).toBe(0);
    expect(summary.trend).toBe('flat');
  });
});

describe('periodGrowthPct', () => {
  it('computes growth correctly', () => {
    expect(periodGrowthPct(120, 100)).toBe(20);
    expect(periodGrowthPct(80, 100)).toBe(-20);
  });

  it('returns 100 for growth from 0', () => {
    expect(periodGrowthPct(50, 0)).toBe(100);
  });

  it('returns 0 when both values are 0', () => {
    expect(periodGrowthPct(0, 0)).toBe(0);
  });
});

describe('aggregateByPeriod', () => {
  it('aggregates by month', () => {
    const points = [
      { label: 'a', value: 100, date: '2026-01-10' },
      { label: 'b', value: 200, date: '2026-01-20' },
      { label: 'c', value: 50, date: '2026-02-05' },
    ];
    const result = aggregateByPeriod(points, 'monthly');
    expect(result).toHaveLength(2);
    const jan = result.find((p) => p.label === '2026-01');
    expect(jan!.value).toBe(300);
  });

  it('aggregates by year', () => {
    const points = [
      { label: 'a', value: 100, date: '2026-01-01' },
      { label: 'b', value: 200, date: '2026-06-01' },
    ];
    const result = aggregateByPeriod(points, 'yearly');
    expect(result).toHaveLength(1);
    expect(result[0].value).toBe(300);
  });
});

describe('rollingAverage', () => {
  it('computes rolling average over N periods', () => {
    const values = [10, 20, 30, 40, 50];
    const avg = rollingAverage(values, 3);
    expect(avg[2]).toBe(20); // (10+20+30)/3
    expect(avg[4]).toBe(40); // (30+40+50)/3
  });

  it('handles window larger than array at start', () => {
    const values = [10, 20];
    const avg = rollingAverage(values, 3);
    expect(avg[0]).toBe(10); // only 1 value
    expect(avg[1]).toBe(15); // (10+20)/2
  });
});

describe('formatReportValue', () => {
  it('formats millions with M suffix', () => {
    expect(formatReportValue(1_500_000)).toBe('1.5M');
  });

  it('formats thousands with K suffix', () => {
    expect(formatReportValue(2500)).toBe('2.5K');
  });

  it('formats small values with 2 decimals', () => {
    expect(formatReportValue(9.99)).toBe('9.99');
  });

  it('applies prefix and suffix', () => {
    expect(formatReportValue(1500, '$', ' USD')).toBe('$1.5K USD');
  });
});
