import { describe, it, expect } from 'vitest';
import { revenueDataSchema, revenueByAppSchema, revenueComparisonSchema } from './revenue';

describe('revenueDataSchema', () => {
  it('parses a full revenue data object', () => {
    const data = {
      period: 'monthly',
      total: 50000,
      change: 5000,
      changePct: 11.1,
      data: [{ date: '2026-01-01', amountMinorUnits: 1000 }],
    };
    const result = revenueDataSchema.parse(data);
    expect(result.period).toBe('monthly');
    expect(result.total).toBe(50000);
    expect(result.data).toHaveLength(1);
  });

  it('accepts an empty object (all fields optional)', () => {
    const result = revenueDataSchema.parse({});
    expect(result.total).toBeUndefined();
  });

  it('accepts the "series" alias for the data array', () => {
    const result = revenueDataSchema.parse({
      series: [{ date: '2026-01-01', amountMinorUnits: 200 }],
    });
    expect(result.series).toHaveLength(1);
  });
});

describe('revenueByAppSchema', () => {
  it('parses a revenue-by-app entry', () => {
    const data = { app: 'pos', amountMinorUnits: 30000, pct: 0.6 };
    const result = revenueByAppSchema.parse(data);
    expect(result.app).toBe('pos');
    expect(result.amountMinorUnits).toBe(30000);
  });

  it('accepts an empty object (all fields optional)', () => {
    expect(() => revenueByAppSchema.parse({})).not.toThrow();
  });
});

describe('revenueComparisonSchema', () => {
  it('parses a revenue comparison object', () => {
    const data = { current: 80000, previous: 70000, change: 10000, changePercent: 14.3 };
    const result = revenueComparisonSchema.parse(data);
    expect(result.current).toBe(80000);
    expect(result.previous).toBe(70000);
  });

  it('accepts both changePercent and changePct aliases', () => {
    const r1 = revenueComparisonSchema.parse({ changePercent: 5 });
    const r2 = revenueComparisonSchema.parse({ changePct: 5 });
    expect(r1.changePercent).toBe(5);
    expect(r2.changePct).toBe(5);
  });

  it('accepts an empty object', () => {
    expect(() => revenueComparisonSchema.parse({})).not.toThrow();
  });
});
