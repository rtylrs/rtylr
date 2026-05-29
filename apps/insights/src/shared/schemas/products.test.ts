import { describe, it, expect } from 'vitest';
import {
  productPerformanceSchema,
  productRankingSchema,
  marginAnalysisSchema,
  categoryBreakdownSchema,
} from './products';

describe('productPerformanceSchema', () => {
  it('parses a product performance summary', () => {
    const data = { totalProducts: 80, avgRevenue: 350, topCategory: 'Beverages' };
    const result = productPerformanceSchema.parse(data);
    expect(result.totalProducts).toBe(80);
    expect(result.topCategory).toBe('Beverages');
  });

  it('accepts a null topCategory', () => {
    const result = productPerformanceSchema.parse({ topCategory: null });
    expect(result.topCategory).toBeNull();
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
    const data = { category: 'Food', revenue: 40000, pct: 0.5, units: 2000 };
    const result = categoryBreakdownSchema.parse(data);
    expect(result.category).toBe('Food');
    expect(result.units).toBe(2000);
  });

  it('accepts the "data" array field', () => {
    const result = categoryBreakdownSchema.parse({ data: [{ date: '2026-01' }] });
    expect(Array.isArray(result.data)).toBe(true);
  });

  it('accepts an empty object', () => {
    expect(() => categoryBreakdownSchema.parse({})).not.toThrow();
  });
});
