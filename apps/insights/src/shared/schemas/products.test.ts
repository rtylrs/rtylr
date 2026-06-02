import { describe, it, expect } from 'vitest';
import {
  productPerformanceSchema,
  productRankingSchema,
  marginAnalysisSchema,
  categoryBreakdownSchema,
} from './products';

describe('productPerformanceSchema', () => {
  it('parses a product performance summary', () => {
    const data = { totalSkus: 80, activeSkus: 70, outOfStock: 5, avgMarginPct: 35 };
    const result = productPerformanceSchema.parse(data);
    expect(result.totalSkus).toBe(80);
    expect(result.avgMarginPct).toBe(35);
  });

  it('accepts an empty object (all fields optional)', () => {
    expect(() => productPerformanceSchema.parse({})).not.toThrow();
  });
});

describe('productRankingSchema', () => {
  it('coerces numeric id to string', () => {
    const result = productRankingSchema.parse({ id: 5, name: 'Espresso Shot', revenue: 8000 });
    expect(result.id).toBe('5');
  });

  it('accepts a string id', () => {
    const result = productRankingSchema.parse({ id: 'sku-001', name: 'Latte' });
    expect(result.id).toBe('sku-001');
  });

  it('accepts a nullable sku field', () => {
    const result = productRankingSchema.parse({ sku: null });
    expect(result.sku).toBeNull();
  });

  it('accepts an empty object', () => {
    expect(() => productRankingSchema.parse({})).not.toThrow();
  });
});

describe('marginAnalysisSchema', () => {
  it('parses a margin analysis summary', () => {
    const data = {
      avgMargin: 0.35,
      averageMargin: 0.35,
      totalRevenue: 200000,
      totalCogs: 130000,
      grossProfit: 70000,
    };
    const result = marginAnalysisSchema.parse(data);
    expect(result.grossProfit).toBe(70000);
    expect(result.avgMargin).toBe(0.35);
  });

  it('accepts an empty object', () => {
    expect(() => marginAnalysisSchema.parse({})).not.toThrow();
  });
});

describe('categoryBreakdownSchema', () => {
  it('parses a category breakdown entry', () => {
    const data = { categoryId: 3, categoryName: 'Food', unitsSold: 2000, revenueMinorUnits: 40000 };
    const result = categoryBreakdownSchema.parse(data);
    expect(result.categoryName).toBe('Food');
    expect(result.unitsSold).toBe(2000);
    expect(result.categoryId).toBe('3');
  });

  it('accepts the "data" array field', () => {
    const result = categoryBreakdownSchema.parse({ data: [{ date: '2026-01' }] });
    expect(Array.isArray(result.data)).toBe(true);
  });

  it('accepts an empty object', () => {
    expect(() => categoryBreakdownSchema.parse({})).not.toThrow();
  });
});
