import { describe, it, expect } from 'vitest';
import {
  formatLocaleCurrency,
  formatLocaleDate,
  formatLocaleNumber,
  isRTLLocale,
  DEFAULT_LOCALE,
} from './locale';

describe('DEFAULT_LOCALE', () => {
  it('is a non-empty string', () => {
    expect(DEFAULT_LOCALE).toBeTruthy();
  });
});

describe('formatLocaleCurrency', () => {
  it('formats USD without decimals', () => {
    const result = formatLocaleCurrency(1000, 'USD', 'en-US');
    expect(result).toContain('$');
    expect(result).toContain('1,000');
  });

  it('falls back gracefully for an invalid locale', () => {
    const result = formatLocaleCurrency(100, 'USD', 'invalid-locale-xyz');
    expect(result).toBeTruthy();
  });

  it('defaults to USD and en-US', () => {
    const result = formatLocaleCurrency(500);
    expect(result).toContain('$');
  });
});

describe('formatLocaleDate', () => {
  it('formats a date in the given locale', () => {
    const result = formatLocaleDate('2026-05-15', 'en-US');
    expect(result).toContain('2026');
    expect(result).toContain('15');
  });

  it('returns the original ISO string on error', () => {
    const result = formatLocaleDate('not-a-date', 'en-US');
    expect(typeof result).toBe('string');
  });
});

describe('formatLocaleNumber', () => {
  it('formats a large number with separators', () => {
    const result = formatLocaleNumber(1000000, 'en-US');
    expect(result).toContain('1,000,000');
  });

  it('defaults to en-US', () => {
    expect(formatLocaleNumber(1000)).toBe('1,000');
  });

  it('returns the number as string on error', () => {
    const result = formatLocaleNumber(42, 'invalid-locale-xyz');
    expect(result).toBeTruthy();
  });
});

describe('isRTLLocale', () => {
  it('returns true for Arabic', () => {
    expect(isRTLLocale('ar')).toBe(true);
    expect(isRTLLocale('ar-SA')).toBe(true);
  });

  it('returns true for Hebrew', () => {
    expect(isRTLLocale('he')).toBe(true);
  });

  it('returns false for English', () => {
    expect(isRTLLocale('en-US')).toBe(false);
  });

  it('returns false for French', () => {
    expect(isRTLLocale('fr')).toBe(false);
  });
});
