import { describe, it, expect } from 'vitest';
import { resolveLocale, localiseNumber, localiseDate, getCollator } from '../locale';

describe('resolveLocale', () => {
  const supported = ['en', 'fr', 'de', 'es'];

  it('returns exact match', () => {
    expect(resolveLocale(['fr'], supported)).toBe('fr');
  });

  it('returns partial match on base language', () => {
    expect(resolveLocale(['fr-CA'], supported)).toBe('fr');
  });

  it('returns default when no match', () => {
    expect(resolveLocale(['ja'], supported)).toBe('en');
  });

  it('picks first matching language in preference order', () => {
    expect(resolveLocale(['de', 'fr'], supported)).toBe('de');
  });
});

describe('localiseNumber', () => {
  it('formats a number for en-US', () => {
    expect(localiseNumber(1000, 'en-US')).toBe('1,000');
  });
});

describe('localiseDate', () => {
  it('formats a date in en-US long format', () => {
    const result = localiseDate('2024-01-15', 'en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
    expect(result).toBe('January 15, 2024');
  });
});

describe('getCollator', () => {
  it('returns an Intl.Collator', () => {
    const collator = getCollator('en');
    expect(typeof collator.compare).toBe('function');
  });

  it('compares strings correctly', () => {
    const collator = getCollator('en');
    expect(collator.compare('apple', 'banana')).toBeLessThan(0);
  });
});
