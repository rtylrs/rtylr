import { describe, it, expect } from 'vitest';
import { insightsKpisSchema, insightsDashboardSummarySchema } from './dashboard';

describe('insightsKpisSchema', () => {
  it('parses a full KPIs object', () => {
    const data = {
      totalRevenue: 500000,
      ordersThisMonth: 1200,
      customerCount: 800,
      staffCount: 30,
      avgOrderValue: 42,
      revenueToday: 8000,
      revenueTrendPct: 5.2,
      headcount: 30,
      lowStockItems: 3,
    };
    const result = insightsKpisSchema.parse(data);
    expect(result.totalRevenue).toBe(500000);
    expect(result.lowStockItems).toBe(3);
  });

  it('accepts topProduct as null', () => {
    const result = insightsKpisSchema.parse({ topProduct: null });
    expect(result.topProduct).toBeNull();
  });

  it('accepts topProduct with name and revenue', () => {
    const result = insightsKpisSchema.parse({
      topProduct: { name: 'Espresso', revenue: 12000 },
    });
    expect(result.topProduct?.name).toBe('Espresso');
  });

  it('accepts revenueSeries and ordersSeries as number arrays', () => {
    const result = insightsKpisSchema.parse({
      revenueSeries: [100, 200, 300],
      ordersSeries: [10, 20, 30],
    });
    expect(result.revenueSeries).toHaveLength(3);
    expect(result.ordersSeries).toHaveLength(3);
  });

  it('accepts an empty object (all fields optional)', () => {
    expect(() => insightsKpisSchema.parse({})).not.toThrow();
  });
});

describe('insightsDashboardSummarySchema', () => {
  it('parses a full dashboard summary', () => {
    const data = {
      totalRevenue: 200000,
      ordersThisMonth: 500,
      customerCount: 300,
      staffCount: 20,
      avgOrderValue: 40,
      revenueChange: 0.12,
      newCustomers: 25,
    };
    const result = insightsDashboardSummarySchema.parse(data);
    expect(result.totalRevenue).toBe(200000);
    expect(result.newCustomers).toBe(25);
  });

  it('accepts topProduct as null', () => {
    const result = insightsDashboardSummarySchema.parse({ topProduct: null });
    expect(result.topProduct).toBeNull();
  });

  it('accepts an empty object', () => {
    expect(() => insightsDashboardSummarySchema.parse({})).not.toThrow();
  });
});
