import { describe, it, expect } from 'vitest';
import { taxConfigSchema, taxReportSchema, plReportSchema } from '../tax';

describe('taxConfigSchema', () => {
  it('parses a valid config', () => {
    const result = taxConfigSchema.parse({ id: 1, name: 'VAT', rate: 20 });
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
  it('defaults vatSummary to zero values', () => {
    const result = taxReportSchema.parse({});
    expect(result.vatSummary.collected).toBe(0);
    expect(result.vatSummary.owed).toBe(0);
    expect(result.vatSummary.paid).toBe(0);
  });

  it('defaults periods to empty array', () => {
    const result = taxReportSchema.parse({});
    expect(result.periods).toEqual([]);
  });

  it('parses a tax period with status', () => {
    const result = taxReportSchema.parse({
      periods: [
        {
          id: 1,
          period: 'Q1 2026',
          revenue: 10000,
          taxableAmount: 8000,
          taxRate: 20,
          collected: 1600,
          status: 'Filed',
          dueDate: '2026-04-30',
        },
      ],
    });
    expect(result.periods[0]?.id).toBe('1');
    expect(result.periods[0]?.status).toBe('Filed');
  });

  it('defaults period status to Pending when invalid', () => {
    const result = taxReportSchema.parse({
      periods: [{ id: '2', period: 'Q2', status: 'Unknown' }],
    });
    expect(result.periods[0]?.status).toBe('Pending');
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
