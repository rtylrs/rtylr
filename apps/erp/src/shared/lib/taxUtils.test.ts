import { describe, it, expect } from 'vitest';
import {
  applyTax,
  taxBreakdown,
  backCalculateSubtotal,
  extractTaxAmount,
  defaultTaxRate,
  formatTaxRate,
  combinedTaxAmount,
  effectiveCombinedRate,
  roundCurrency,
} from './taxUtils';
import type { TaxRate } from './taxUtils';

const gst: TaxRate = { id: 1, name: 'GST', rate: 10, isDefault: true };
const vat: TaxRate = { id: 2, name: 'VAT', rate: 20, isDefault: false };

describe('applyTax', () => {
  it('applies tax rate to subtotal', () => {
    expect(applyTax(100, 10)).toBe(10);
    expect(applyTax(200, 15)).toBe(30);
  });

  it('returns 0 for zero rate', () => {
    expect(applyTax(100, 0)).toBe(0);
  });
});

describe('taxBreakdown', () => {
  it('computes correct breakdown', () => {
    const breakdown = taxBreakdown(100, 10);
    expect(breakdown.subtotal).toBe(100);
    expect(breakdown.taxAmount).toBe(10);
    expect(breakdown.total).toBe(110);
    expect(breakdown.effectiveRate).toBe(10);
  });

  it('rounds to 2 decimal places', () => {
    const breakdown = taxBreakdown(99.99, 10);
    expect(breakdown.taxAmount).toBe(10);
    expect(breakdown.total).toBe(109.99);
  });
});

describe('backCalculateSubtotal', () => {
  it('extracts subtotal from tax-inclusive total', () => {
    const subtotal = backCalculateSubtotal(110, 10);
    expect(Math.round(subtotal)).toBe(100);
  });
});

describe('extractTaxAmount', () => {
  it('extracts tax from tax-inclusive total', () => {
    const tax = extractTaxAmount(110, 10);
    expect(Math.round(tax)).toBe(10);
  });
});

describe('defaultTaxRate', () => {
  it('returns the default rate', () => {
    const rates = [gst, vat];
    expect(defaultTaxRate(rates)).toEqual(gst);
  });

  it('returns undefined when no default exists', () => {
    const rates = [
      { ...gst, isDefault: false },
      { ...vat, isDefault: false },
    ];
    expect(defaultTaxRate(rates)).toBeUndefined();
  });
});

describe('formatTaxRate', () => {
  it('formats rate with name and percentage', () => {
    expect(formatTaxRate(gst)).toBe('GST 10%');
    expect(formatTaxRate(vat)).toBe('VAT 20%');
  });
});

describe('combinedTaxAmount', () => {
  it('sums tax across multiple rates', () => {
    expect(combinedTaxAmount(100, [gst, vat])).toBe(30);
  });

  it('returns 0 for empty rates', () => {
    expect(combinedTaxAmount(100, [])).toBe(0);
  });
});

describe('effectiveCombinedRate', () => {
  it('sums all rates', () => {
    expect(effectiveCombinedRate([gst, vat])).toBe(30);
  });
});

describe('roundCurrency', () => {
  it('rounds to 2 decimal places', () => {
    expect(roundCurrency(10.555)).toBe(10.56);
    expect(roundCurrency(10.554)).toBe(10.55);
  });
});
