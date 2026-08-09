import { describe, it, expect } from 'vitest';
import { convertCurrency, roundCurrency, buildRateMap, convertWithMap } from './currencyUtils';
import type { ExchangeRate } from './currencyUtils';

const rates: ExchangeRate[] = [
  { from: 'USD', to: 'EUR', rate: 0.92 },
  { from: 'USD', to: 'GBP', rate: 0.78 },
  { from: 'EUR', to: 'USD', rate: 1.087 },
];

describe('convertCurrency', () => {
  it('converts USD to EUR using direct rate', () => {
    expect(convertCurrency(100, 'USD', 'EUR', rates)).toBeCloseTo(92);
  });

  it('converts EUR to USD using direct rate', () => {
    expect(convertCurrency(100, 'EUR', 'USD', rates)).toBeCloseTo(108.7);
  });

  it('returns amount unchanged when from === to', () => {
    expect(convertCurrency(50, 'USD', 'USD', rates)).toBe(50);
  });

  it('uses inverse rate when direct is missing', () => {
    // GBP to USD: inverse of USD->GBP (0.78) = 1/0.78
    const result = convertCurrency(78, 'GBP', 'USD', rates);
    expect(result).toBeCloseTo(100, 0);
  });

  it('returns null when no rate found', () => {
    expect(convertCurrency(100, 'JPY', 'EUR', rates)).toBeNull();
  });
});

describe('roundCurrency', () => {
  it('rounds to 2 decimal places by default', () => {
    expect(roundCurrency(10.555)).toBe(10.56);
  });

  it('rounds to specified decimal places', () => {
    expect(roundCurrency(10.1234, 3)).toBe(10.123);
  });
});

describe('buildRateMap', () => {
  it('builds a map with correct keys', () => {
    const map = buildRateMap(rates);
    expect(map.get('USD:EUR')).toBe(0.92);
    expect(map.get('USD:GBP')).toBe(0.78);
  });
});

describe('convertWithMap', () => {
  it('converts using map', () => {
    const map = buildRateMap(rates);
    expect(convertWithMap(100, 'USD', 'EUR', map)).toBeCloseTo(92);
  });

  it('returns same amount for same currency', () => {
    const map = buildRateMap(rates);
    expect(convertWithMap(50, 'USD', 'USD', map)).toBe(50);
  });

  it('returns null for unknown pair', () => {
    const map = buildRateMap(rates);
    expect(convertWithMap(100, 'JPY', 'CNY', map)).toBeNull();
  });
});
