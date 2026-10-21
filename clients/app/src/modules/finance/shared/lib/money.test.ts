import { describe, it, expect } from 'vitest';
import {
  getCurrencySymbol,
  formatMoney,
  formatCompact,
  convert,
  parseMoney,
  CURRENCY_OPTIONS,
  FX_RATES,
} from './money';
import type { Currency } from './money';

describe('getCurrencySymbol', () => {
  it('returns $ for USD', () => {
    expect(getCurrencySymbol('USD')).toBe('$');
  });

  it('returns € for EUR', () => {
    expect(getCurrencySymbol('EUR')).toBe('€');
  });

  it('returns £ for GBP', () => {
    expect(getCurrencySymbol('GBP')).toBe('£');
  });

  it('returns ¥ for JPY', () => {
    expect(getCurrencySymbol('JPY')).toBe('¥');
  });

  it('returns CA$ for CAD', () => {
    expect(getCurrencySymbol('CAD')).toBe('CA$');
  });

  it('returns A$ for AUD', () => {
    expect(getCurrencySymbol('AUD')).toBe('A$');
  });

  it('returns CHF for CHF', () => {
    expect(getCurrencySymbol('CHF')).toBe('CHF');
  });

  it('returns the currency code itself when not in map (unknown)', () => {
    // Cast to Currency to exercise the fallback branch.
    expect(getCurrencySymbol('XYZ' as Currency)).toBe('XYZ');
  });
});

describe('formatMoney', () => {
  it('formats a whole number with 2 decimal places', () => {
    const result = formatMoney(1000);
    expect(result).toBe('$1,000.00');
  });

  it('formats a decimal value', () => {
    const result = formatMoney(9.99);
    expect(result).toBe('$9.99');
  });

  it('formats a string input', () => {
    const result = formatMoney('250');
    expect(result).toBe('$250.00');
  });

  it('formats zero', () => {
    expect(formatMoney(0)).toBe('$0.00');
  });

  it('returns "-" for NaN input', () => {
    expect(formatMoney(NaN)).toBe('-');
  });

  it('returns "-" for Infinity', () => {
    expect(formatMoney(Infinity)).toBe('-');
  });

  it('treats null as 0', () => {
    expect(formatMoney(null)).toBe('$0.00');
  });

  it('treats undefined as 0', () => {
    expect(formatMoney(undefined)).toBe('$0.00');
  });

  it('respects a custom currency', () => {
    const result = formatMoney(100, 'EUR');
    expect(result).toContain('100');
    expect(result).toContain('€');
  });

  it('respects custom fractionDigits', () => {
    const result = formatMoney(1234, 'USD', 0);
    expect(result).toBe('$1,234');
  });
});

describe('formatCompact', () => {
  it('returns plain value for amounts under 1 000', () => {
    expect(formatCompact(500)).toBe('$500');
  });

  it('formats thousands with K suffix', () => {
    expect(formatCompact(1_500)).toBe('$1.5K');
  });

  it('formats millions with M suffix', () => {
    expect(formatCompact(2_500_000)).toBe('$2.5M');
  });

  it('formats billions with B suffix', () => {
    expect(formatCompact(3_200_000_000)).toBe('$3.2B');
  });

  it('works with negative values', () => {
    const result = formatCompact(-1_000);
    // absolute value ≥ 1_000 → K branch
    expect(result).toContain('K');
  });

  it('respects custom currency symbol', () => {
    expect(formatCompact(5_000, 'EUR')).toBe('€5.0K');
  });
});

describe('convert', () => {
  it('returns the same amount when converting USD to USD', () => {
    expect(convert(100, 'USD', 'USD')).toBeCloseTo(100);
  });

  it('converts USD to EUR correctly', () => {
    const expected = 100 * FX_RATES['EUR'];
    expect(convert(100, 'USD', 'EUR')).toBeCloseTo(expected);
  });

  it('converts EUR back to USD', () => {
    // EUR → USD: divide by EUR rate, multiply by USD rate (1)
    const expected = 93 / FX_RATES['EUR'];
    expect(convert(93, 'EUR', 'USD')).toBeCloseTo(expected);
  });

  it('converts between two non-USD currencies via USD pivot', () => {
    // GBP → EUR: (amount / GBP_rate) * EUR_rate
    const inUsd = 79 / FX_RATES['GBP'];
    const expected = inUsd * FX_RATES['EUR'];
    expect(convert(79, 'GBP', 'EUR')).toBeCloseTo(expected);
  });
});

describe('parseMoney', () => {
  it('parses a plain number string', () => {
    expect(parseMoney('123.45')).toBeCloseTo(123.45);
  });

  it('strips currency symbols', () => {
    expect(parseMoney('$1,234.56')).toBeCloseTo(1234.56);
  });

  it('strips € and spaces', () => {
    expect(parseMoney('€ 500')).toBeCloseTo(500);
  });

  it('handles negative values', () => {
    expect(parseMoney('-99.99')).toBeCloseTo(-99.99);
  });

  it('returns 0 for non-numeric strings', () => {
    expect(parseMoney('abc')).toBe(0);
  });

  it('returns 0 for empty string', () => {
    expect(parseMoney('')).toBe(0);
  });
});

describe('CURRENCY_OPTIONS', () => {
  it('contains all 8 supported currencies', () => {
    expect(CURRENCY_OPTIONS).toHaveLength(8);
    expect(CURRENCY_OPTIONS).toContain('USD');
    expect(CURRENCY_OPTIONS).toContain('JPY');
  });
});
