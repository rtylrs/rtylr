import { describe, it, expect } from 'vitest';
import {
  parseCurrencyString,
  formatWithSymbol,
  absAmount,
  centsToDollars,
  dollarsToCents,
  amountsEqual,
} from './currencyUtils';

describe('parseCurrencyString', () => {
  it('parses a plain numeric string', () => {
    expect(parseCurrencyString('12.50')).toBe(12.5);
  });

  it('parses a dollar-sign prefixed string', () => {
    expect(parseCurrencyString('$12.50')).toBe(12.5);
  });

  it('parses a string with commas', () => {
    expect(parseCurrencyString('$1,000.00')).toBe(1000);
  });

  it('returns null for an empty string', () => {
    expect(parseCurrencyString('')).toBeNull();
  });

  it('returns null for a non-numeric string', () => {
    expect(parseCurrencyString('abc')).toBeNull();
  });

  it('parses zero', () => {
    expect(parseCurrencyString('0')).toBe(0);
  });
});

describe('formatWithSymbol', () => {
  it('formats a positive amount', () => {
    expect(formatWithSymbol(12.5)).toBe('$12.50');
  });

  it('formats zero', () => {
    expect(formatWithSymbol(0)).toBe('$0.00');
  });

  it('formats a negative amount with a leading minus', () => {
    expect(formatWithSymbol(-5.99)).toBe('-$5.99');
  });

  it('uses a custom symbol', () => {
    expect(formatWithSymbol(10, 'EUR ')).toBe('EUR 10.00');
  });

  it('returns "$0.00" for NaN', () => {
    expect(formatWithSymbol(NaN)).toBe('$0.00');
  });
});

describe('absAmount', () => {
  it('returns the absolute value of a positive number', () => {
    expect(absAmount(10)).toBe(10);
  });

  it('returns the absolute value of a negative number', () => {
    expect(absAmount(-10)).toBe(10);
  });

  it('returns 0 for NaN', () => {
    expect(absAmount(NaN)).toBe(0);
  });
});

describe('centsToDollars', () => {
  it('converts 100 cents to $1', () => {
    expect(centsToDollars(100)).toBe(1);
  });

  it('converts 1250 cents to $12.50', () => {
    expect(centsToDollars(1250)).toBe(12.5);
  });

  it('returns 0 for NaN', () => {
    expect(centsToDollars(NaN)).toBe(0);
  });
});

describe('dollarsToCents', () => {
  it('converts $1 to 100 cents', () => {
    expect(dollarsToCents(1)).toBe(100);
  });

  it('converts $12.50 to 1250 cents', () => {
    expect(dollarsToCents(12.5)).toBe(1250);
  });

  it('rounds correctly for floating point values', () => {
    expect(dollarsToCents(0.1 + 0.2)).toBe(30);
  });

  it('returns 0 for NaN', () => {
    expect(dollarsToCents(NaN)).toBe(0);
  });
});

describe('amountsEqual', () => {
  it('returns true for equal amounts', () => {
    expect(amountsEqual(10, 10)).toBe(true);
  });

  it('returns true within the default tolerance', () => {
    expect(amountsEqual(10.0001, 10)).toBe(true);
  });

  it('returns false when amounts differ more than the tolerance', () => {
    expect(amountsEqual(10.01, 10)).toBe(false);
  });

  it('returns false for NaN inputs', () => {
    expect(amountsEqual(NaN, 10)).toBe(false);
    expect(amountsEqual(10, NaN)).toBe(false);
  });

  it('accepts a custom tolerance', () => {
    expect(amountsEqual(10.05, 10, 0.1)).toBe(true);
    expect(amountsEqual(10.2, 10, 0.1)).toBe(false);
  });
});
