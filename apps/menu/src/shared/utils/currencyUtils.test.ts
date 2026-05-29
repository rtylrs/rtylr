import { describe, it, expect } from 'vitest';
import {
  currencySymbol,
  roundTo,
  centsToDollars,
  dollarsToCents,
  applyDiscount,
  calcTax,
} from './currencyUtils';

describe('currencySymbol', () => {
  it('returns $ for USD', () => {
    expect(currencySymbol('USD')).toBe('$');
  });

  it('falls back to the code for unknown currencies', () => {
    expect(currencySymbol('XYZ')).toBe('XYZ');
  });
});

describe('roundTo', () => {
  it('rounds to two decimal places by default', () => {
    expect(roundTo(1.125)).toBe(1.13);
  });

  it('rounds to zero decimal places', () => {
    expect(roundTo(2.6, 0)).toBe(3);
  });

  it('leaves integers unchanged', () => {
    expect(roundTo(5, 2)).toBe(5);
  });
});

describe('centsToDollars', () => {
  it('converts 100 cents to 1.00', () => {
    expect(centsToDollars(100)).toBe(1);
  });

  it('converts 999 cents to 9.99', () => {
    expect(centsToDollars(999)).toBe(9.99);
  });

  it('converts 0 cents to 0', () => {
    expect(centsToDollars(0)).toBe(0);
  });
});

describe('dollarsToCents', () => {
  it('converts $9.99 to 999', () => {
    expect(dollarsToCents(9.99)).toBe(999);
  });

  it('converts $0 to 0', () => {
    expect(dollarsToCents(0)).toBe(0);
  });

  it('converts $1.00 to 100', () => {
    expect(dollarsToCents(1.0)).toBe(100);
  });
});

describe('applyDiscount', () => {
  it('applies 10% discount to $10 giving $9', () => {
    expect(applyDiscount(10, 10)).toBe(9);
  });

  it('returns 0 for 100% discount', () => {
    expect(applyDiscount(10, 100)).toBe(0);
  });

  it('returns the original price for 0% discount', () => {
    expect(applyDiscount(10, 0)).toBe(10);
  });

  it('returns the original price for negative discount', () => {
    expect(applyDiscount(10, -5)).toBe(10);
  });
});

describe('calcTax', () => {
  it('calculates 10% tax on $100 as $10', () => {
    expect(calcTax(100, 10)).toBe(10);
  });

  it('calculates 8.5% tax on $10 and rounds correctly', () => {
    expect(calcTax(10, 8.5)).toBe(0.85);
  });

  it('returns 0 tax for 0% rate', () => {
    expect(calcTax(50, 0)).toBe(0);
  });
});
