import { describe, it, expect } from 'vitest';
import { insightsKpisSchema, insightsDashboardSummarySchema } from './dashboard';

describe('insightsKpisSchema', () => {
  it('parses a full KPIs object', () => {
    const data = {
      revenue: 500000,
      orderCount: 1200,
      staffCount: 30,
      customerCount: 800,
      lowStockCount: 3,
      activeApps: 5,
    };
    const result = insightsKpisSchema.parse(data);
    expect(result.revenue).toBe(500000);
    expect(result.lowStockCount).toBe(3);
  });

  it('rejects an object missing required KPI fields', () => {
    expect(() => insightsKpisSchema.parse({})).toThrow();
    expect(() => insightsKpisSchema.parse({ revenue: 1, orderCount: 2 })).toThrow();
  });
});

describe('insightsDashboardSummarySchema', () => {
  it('parses a full dashboard summary of metric cards', () => {
    const data = {
      revenue: { value: 200000, deltaPct: 0.12 },
      customers: { value: 300, deltaPct: 0.05 },
      orders: { value: 500, deltaPct: -0.02 },
      margin: { value: 35.4, deltaPct: 1.1 },
    };
    const result = insightsDashboardSummarySchema.parse(data);
    expect(result.revenue.value).toBe(200000);
    expect(result.margin.deltaPct).toBe(1.1);
  });

  it('rejects a summary missing one of the four cards', () => {
    const data = {
      revenue: { value: 200000, deltaPct: 0.12 },
      customers: { value: 300, deltaPct: 0.05 },
      orders: { value: 500, deltaPct: -0.02 },
    };
    expect(() => insightsDashboardSummarySchema.parse(data)).toThrow();
  });

  it('rejects a card missing deltaPct', () => {
    const data = {
      revenue: { value: 200000 },
      customers: { value: 300, deltaPct: 0.05 },
      orders: { value: 500, deltaPct: -0.02 },
      margin: { value: 35.4, deltaPct: 1.1 },
    };
    expect(() => insightsDashboardSummarySchema.parse(data)).toThrow();
  });
});
