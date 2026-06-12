import { describe, it, expect } from 'vitest';
import { daySummarySchema } from './daySummary';

describe('daySummarySchema', () => {
  const base = {
    date: '2024-06-01',
    grossSales: 1000,
    netSales: 920,
    taxCollected: 80,
    tips: 50,
    refunds: 20,
    voids: 0,
    orderCount: 40,
    averageTicket: 25,
    cashDrops: 200,
    openingFloat: 100,
    closingCash: 350,
  };

  it('parses a fully-populated day summary', () => {
    const result = daySummarySchema.parse(base);
    expect(result.date).toBe('2024-06-01');
    expect(result.grossSales).toBe(1000);
    expect(result.netSales).toBe(920);
    expect(result.taxCollected).toBe(80);
    expect(result.tips).toBe(50);
    expect(result.refunds).toBe(20);
    expect(result.voids).toBe(0);
    expect(result.orderCount).toBe(40);
    expect(result.averageTicket).toBe(25);
    expect(result.cashDrops).toBe(200);
    expect(result.openingFloat).toBe(100);
    expect(result.closingCash).toBe(350);
  });

  it('defaults all numeric fields to 0 when missing', () => {
    const result = daySummarySchema.parse({ date: '2024-06-01' });
    expect(result.grossSales).toBe(0);
    expect(result.netSales).toBe(0);
    expect(result.taxCollected).toBe(0);
    expect(result.tips).toBe(0);
    expect(result.refunds).toBe(0);
    expect(result.voids).toBe(0);
    expect(result.orderCount).toBe(0);
    expect(result.averageTicket).toBe(0);
    expect(result.cashDrops).toBe(0);
    expect(result.openingFloat).toBe(0);
    expect(result.closingCash).toBe(0);
  });

  it('defaults date to today when missing', () => {
    const result = daySummarySchema.parse({});
    expect(result.date).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});
