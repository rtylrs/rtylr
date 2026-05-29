import { describe, it, expect } from 'vitest';
import {
  formatCurrency,
  roundMoney,
  discountAmount,
  applyTax,
  extractTax,
  lineItemsSubtotal,
  isValidAmount,
} from './currencyUtils';

describe('formatCurrency', () => {
  it('formats USD by default', () => {
    expect(formatCurrency(10.5)).toContain('10.50');
  });

  it('includes currency symbol', () => {
    expect(formatCurrency(100, 'USD')).toContain('$');
  });

  it('formats zero', () => {
    expect(formatCurrency(0)).toContain('0.00');
  });
});

describe('roundMoney', () => {
  it('rounds to 2 decimal places', () => {
    expect(roundMoney(1.005)).toBeCloseTo(1.01, 1);
  });

  it('returns integer amounts unchanged', () => {
    expect(roundMoney(100)).toBe(100);
  });
});

describe('discountAmount', () => {
  it('computes 10% discount', () => {
    expect(discountAmount(100, 10)).toBe(10);
  });

  it('returns 0 for 0% discount', () => {
    expect(discountAmount(100, 0)).toBe(0);
  });

  it('caps at 100%', () => {
    expect(discountAmount(100, 150)).toBe(100);
  });
});

describe('applyTax', () => {
  it('adds tax to amount', () => {
    expect(applyTax(100, 10)).toBe(110);
  });

  it('returns unchanged for 0% tax', () => {
    expect(applyTax(100, 0)).toBe(100);
  });
});

describe('extractTax', () => {
  it('extracts tax from inclusive total', () => {
    expect(extractTax(110, 10)).toBeCloseTo(10, 1);
  });

  it('returns 0 for 0% tax', () => {
    expect(extractTax(100, 0)).toBe(0);
  });
});

describe('lineItemsSubtotal', () => {
  it('sums line item totals', () => {
    const items = [
      { quantity: 2, unitPrice: 10 },
      { quantity: 3, unitPrice: 5 },
    ];
    expect(lineItemsSubtotal(items)).toBe(35);
  });

  it('returns 0 for empty items', () => {
    expect(lineItemsSubtotal([])).toBe(0);
  });
});

describe('isValidAmount', () => {
  it('returns true for positive finite number', () => {
    expect(isValidAmount(100)).toBe(true);
  });

  it('returns false for zero', () => {
    expect(isValidAmount(0)).toBe(false);
  });

  it('returns false for negative', () => {
    expect(isValidAmount(-10)).toBe(false);
  });

  it('returns false for Infinity', () => {
    expect(isValidAmount(Infinity)).toBe(false);
  });

  it('returns false for string', () => {
    expect(isValidAmount('100')).toBe(false);
  });
});
