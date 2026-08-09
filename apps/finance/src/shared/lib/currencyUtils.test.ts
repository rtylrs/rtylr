import { describe, it, expect } from 'vitest';
import {
  getCurrencySymbol,
  formatCurrency,
  formatCurrencyCompact,
  convertCurrency,
  parseCurrencyString,
} from './currencyUtils';

describe('getCurrencySymbol', () => {
  it('returns $ for USD', () => {
    expect(getCurrencySymbol('USD')).toBe('$');
  });

  it('returns EUR symbol for EUR', () => {
    expect(getCurrencySymbol('EUR')).toBe('€');
  });

  it('returns GBP symbol for GBP', () => {
    expect(getCurrencySymbol('GBP')).toBe('£');
  });

  it('returns code itself for unknown currency', () => {
    expect(getCurrencySymbol('XYZ')).toBe('XYZ');
  });
});

describe('formatCurrency', () => {
  it('formats a dollar amount', () => {
    const result = formatCurrency(1234.56, 'USD', 'en-US');
    expect(result).toContain('1,234.56');
  });

  it('includes two decimal places', () => {
    const result = formatCurrency(100, 'USD', 'en-US');
    expect(result).toContain('100.00');
  });

  it('formats negative amounts', () => {
    const result = formatCurrency(-500, 'USD', 'en-US');
    expect(result).toContain('500.00');
  });
});

describe('formatCurrencyCompact', () => {
  it('formats millions as M suffix', () => {
    const result = formatCurrencyCompact(2_500_000, 'USD');
    expect(result).toBe('$2.5M');
  });

  it('formats thousands as K suffix', () => {
    const result = formatCurrencyCompact(1500, 'USD');
    expect(result).toBe('$1.5K');
  });

  it('formats small amounts normally', () => {
    const result = formatCurrencyCompact(99, 'USD', 'en-US');
    expect(result).toContain('99.00');
  });
});

describe('convertCurrency', () => {
  it('converts amount using exchange rate', () => {
    expect(convertCurrency(100, 1.2)).toBeCloseTo(120);
  });

  it('returns 0 for zero rate', () => {
    expect(convertCurrency(100, 0)).toBe(0);
  });

  it('returns 0 for negative rate', () => {
    expect(convertCurrency(100, -1)).toBe(0);
  });

  it('rounds to 2 decimal places', () => {
    expect(convertCurrency(100, 1.235)).toBeCloseTo(123.5);
  });
});

describe('parseCurrencyString', () => {
  it('parses a dollar amount string', () => {
    expect(parseCurrencyString('$1,234.56')).toBeCloseTo(1234.56);
  });

  it('parses a plain number string', () => {
    expect(parseCurrencyString('500')).toBe(500);
  });

  it('parses negative values', () => {
    expect(parseCurrencyString('-250.00')).toBe(-250);
  });

  it('returns NaN for empty string', () => {
    expect(parseCurrencyString('')).toBeNaN();
  });
});
