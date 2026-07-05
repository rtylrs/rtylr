import { describe, it, expect } from 'vitest';
import { taxConfigSchema, taxReportSchema, plReportSchema } from '../tax';

describe('taxConfigSchema', () => {
  it('parses a valid config', () => {
    const result = taxConfigSchema.parse({ id: '1', name: 'VAT', rate: 20 });
    expect(result.name).toBe('VAT');
    expect(result.rate).toBe(20);
  });

  it('defaults name and rate when missing', () => {
    const result = taxConfigSchema.parse({});
    expect(result.name).toBe('');
    expect(result.rate).toBe(0);
  });

  it('allows null id', () => {
    const result = taxConfigSchema.parse({ id: null, name: 'GST', rate: 10 });
    expect(result.id).toBeNull();
  });
});

describe('taxReportSchema', () => {
  it('defaults summary fields to zero values', () => {
    const result = taxReportSchema.parse({});
    expect(result.period).toBe('');
    expect(result.taxCollected).toBe(0);
    expect(result.taxOwed).toBe(0);
  });

  it('parses the backend tax report shape', () => {
    const result = taxReportSchema.parse({
      period: '2026-Q1',
      taxCollected: 1600,
      taxOwed: 1200,
    });
    expect(result.period).toBe('2026-Q1');
    expect(result.taxCollected).toBe(1600);
    expect(result.taxOwed).toBe(1200);
  });
});

describe('plReportSchema', () => {
  it('defaults all fields to zero and empty string', () => {
    const result = plReportSchema.parse({});
    expect(result.period).toBe('');
    expect(result.revenue).toBe(0);
    expect(result.cogs).toBe(0);
    expect(result.grossProfit).toBe(0);
    expect(result.netProfit).toBe(0);
  });

  it('parses a valid PL report', () => {
    const result = plReportSchema.parse({
      period: '2026-Q1',
      revenue: 100000,
      cogs: 40000,
      operatingExpenses: 20000,
      grossProfit: 60000,
      netProfit: 40000,
    });
    expect(result.revenue).toBe(100000);
    expect(result.grossProfit).toBe(60000);
  });
});
