import { describe, it, expect } from 'vitest';
import { computeTaxBreakdown, formatTaxRate, roundMoney } from './tax';

describe('computeTaxBreakdown', () => {
  it('computes exclusive tax on subtotal', () => {
    const r = computeTaxBreakdown({ subtotal: 1000, taxRate: 20 });
    expect(r.taxAmount).toBeCloseTo(200);
    expect(r.total).toBeCloseTo(1200);
  });

  it('applies discount before tax', () => {
    const r = computeTaxBreakdown({ subtotal: 1000, taxRate: 10, discount: 100 });
    expect(r.taxableAmount).toBeCloseTo(900);
    expect(r.taxAmount).toBeCloseTo(90);
  });

  it('clamps discount to subtotal', () => {
    const r = computeTaxBreakdown({ subtotal: 200, taxRate: 10, discount: 500 });
    expect(r.discount).toBe(200);
  });

  it('extracts inclusive VAT from subtotal', () => {
    const r = computeTaxBreakdown({ subtotal: 1200, taxRate: 20, inclusive: true });
    expect(r.taxableAmount).toBeCloseTo(1000);
    expect(r.taxAmount).toBeCloseTo(200);
  });

  it('ignores negative tax rate', () => {
    expect(computeTaxBreakdown({ subtotal: 1000, taxRate: -5 }).taxAmount).toBe(0);
  });
});

describe('formatTaxRate', () => {
  it('formats 20%', () => {
    expect(formatTaxRate(20)).toBe('20.00%');
  });

  it('respects custom digits', () => {
    expect(formatTaxRate(10, 0)).toBe('10%');
  });
});

describe('roundMoney', () => {
  it('rounds down correctly', () => {
    expect(roundMoney(1.004)).toBe(1);
  });

  it('rounds to 0 decimals', () => {
    expect(roundMoney(4.6, 0)).toBe(5);
  });

  it('handles whole numbers', () => {
    expect(roundMoney(100)).toBe(100);
  });
});
