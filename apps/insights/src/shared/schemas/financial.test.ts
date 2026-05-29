import { describe, it, expect } from 'vitest';
import {
  profitLossSchema,
  expenseBreakdownSchema,
  invoiceAgingSchema,
  cashFlowSchema,
} from './financial';

describe('profitLossSchema', () => {
  it('parses a full profit-and-loss object', () => {
    const data = {
      revenue: 100000,
      totalRevenue: 100000,
      expenses: 60000,
      totalExpenses: 60000,
      netProfit: 40000,
      margin: 0.4,
    };
    const result = profitLossSchema.parse(data);
    expect(result.netProfit).toBe(40000);
    expect(result.margin).toBe(0.4);
  });

  it('accepts an empty object (all fields optional)', () => {
    expect(() => profitLossSchema.parse({})).not.toThrow();
  });

  it('accepts both revenue aliases', () => {
    const r = profitLossSchema.parse({ revenue: 1000, totalRevenue: 1000 });
    expect(r.revenue).toBe(1000);
    expect(r.totalRevenue).toBe(1000);
  });
});

describe('expenseBreakdownSchema', () => {
  it('parses an expense breakdown entry', () => {
    const data = { category: 'salaries', amount: 40000, pct: 0.67, change: 1000 };
    const result = expenseBreakdownSchema.parse(data);
    expect(result.category).toBe('salaries');
    expect(result.amount).toBe(40000);
  });

  it('accepts the "categories" array alias', () => {
    const result = expenseBreakdownSchema.parse({ categories: [{ label: 'ops' }] });
    expect(Array.isArray(result.categories)).toBe(true);
  });

  it('accepts an empty object', () => {
    expect(() => expenseBreakdownSchema.parse({})).not.toThrow();
  });
});

describe('invoiceAgingSchema', () => {
  it('parses an invoice aging object', () => {
    const data = { current: 10000, days30: 5000, days60: 2000, days90: 500, over90: 100 };
    const result = invoiceAgingSchema.parse(data);
    expect(result.current).toBe(10000);
    expect(result.over90).toBe(100);
  });

  it('accepts both "data" and "buckets" array fields', () => {
    const r = invoiceAgingSchema.parse({ data: [], buckets: [] });
    expect(Array.isArray(r.data)).toBe(true);
    expect(Array.isArray(r.buckets)).toBe(true);
  });

  it('accepts an empty object', () => {
    expect(() => invoiceAgingSchema.parse({})).not.toThrow();
  });
});

describe('cashFlowSchema', () => {
  it('parses a cash flow summary', () => {
    const data = { period: 'Q1', inflow: 80000, outflow: 60000, net: 20000 };
    const result = cashFlowSchema.parse(data);
    expect(result.inflow).toBe(80000);
    expect(result.net).toBe(20000);
  });

  it('accepts the "periods" array alias', () => {
    const result = cashFlowSchema.parse({ periods: [{ month: 'Jan', value: 1000 }] });
    expect(Array.isArray(result.periods)).toBe(true);
  });

  it('accepts an empty object', () => {
    expect(() => cashFlowSchema.parse({})).not.toThrow();
  });
});
