import { describe, it, expect } from 'vitest';
import {
  toCents,
  fromCents,
  subtotal,
  applyDiscountPercent,
  taxAmount,
  grandTotal,
  totalItemCount,
} from './cartMathUtils';

describe('toCents', () => {
  it('converts dollars to cents', () => {
    expect(toCents(9.99)).toBe(999);
  });

  it('rounds correctly', () => {
    expect(toCents(0.1 + 0.2)).toBe(30);
  });
});

describe('fromCents', () => {
  it('converts cents to dollars', () => {
    expect(fromCents(999)).toBe(9.99);
  });
});

describe('subtotal', () => {
  it('computes the sum of price times quantity for each line', () => {
    const lines = [
      { price: 10.0, quantity: 2 },
      { price: 5.5, quantity: 1 },
    ];
    expect(subtotal(lines)).toBe(25.5);
  });

  it('returns 0 for an empty cart', () => {
    expect(subtotal([])).toBe(0);
  });
});

describe('applyDiscountPercent', () => {
  it('applies a 10% discount correctly', () => {
    expect(applyDiscountPercent(100, 10)).toBe(90);
  });

  it('returns the original amount for 0% discount', () => {
    expect(applyDiscountPercent(50, 0)).toBe(50);
  });

  it('returns 0 for a 100% discount', () => {
    expect(applyDiscountPercent(50, 100)).toBe(0);
  });
});

describe('taxAmount', () => {
  it('computes the tax on a subtotal', () => {
    expect(taxAmount(100, 10)).toBe(10);
  });

  it('rounds correctly', () => {
    expect(taxAmount(9.99, 10)).toBe(1.0);
  });
});

describe('grandTotal', () => {
  it('adds tax to the subtotal', () => {
    expect(grandTotal(100, 10)).toBe(110);
  });

  it('subtracts a discount', () => {
    expect(grandTotal(100, 10, 5)).toBe(105);
  });

  it('handles zero tax', () => {
    expect(grandTotal(50, 0)).toBe(50);
  });
});

describe('totalItemCount', () => {
  it('sums all quantities', () => {
    const lines = [{ quantity: 3 }, { quantity: 2 }];
    expect(totalItemCount(lines)).toBe(5);
  });

  it('returns 0 for empty cart', () => {
    expect(totalItemCount([])).toBe(0);
  });
});
