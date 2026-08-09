import { describe, it, expect } from 'vitest';
import {
  formatPriceRange,
  formatCartQuantity,
  excerptDescription,
  lineTotal,
  isValidCategorySlug,
} from './menuUtils';

describe('formatPriceRange', () => {
  it('returns a single price when min equals max', () => {
    expect(formatPriceRange({ min: 10, max: 10 })).toBe('$10.00');
  });

  it('returns a range when min differs from max', () => {
    expect(formatPriceRange({ min: 5, max: 15 })).toBe('$5.00 - $15.00');
  });
});

describe('formatCartQuantity', () => {
  it('prefixes the quantity with x', () => {
    expect(formatCartQuantity(3)).toBe('x3');
  });

  it('handles quantity of 1', () => {
    expect(formatCartQuantity(1)).toBe('x1');
  });
});

describe('excerptDescription', () => {
  it('returns the full description when within limit', () => {
    expect(excerptDescription('Short text', 120)).toBe('Short text');
  });

  it('truncates long descriptions with ellipsis', () => {
    const long = 'a'.repeat(130);
    const result = excerptDescription(long, 120);
    expect(result.length).toBe(120);
    expect(result.endsWith('...')).toBe(true);
  });

  it('uses 120 as the default max length', () => {
    const long = 'b'.repeat(130);
    const result = excerptDescription(long);
    expect(result.length).toBe(120);
  });
});

describe('lineTotal', () => {
  it('multiplies unit price by quantity', () => {
    expect(lineTotal(9.99, 3)).toBe(29.97);
  });

  it('returns 0 for 0 quantity', () => {
    expect(lineTotal(5.0, 0)).toBe(0);
  });

  it('rounds to 2 decimal places', () => {
    expect(lineTotal(0.1, 3)).toBe(0.3);
  });
});

describe('isValidCategorySlug', () => {
  it('returns true for a valid slug', () => {
    expect(isValidCategorySlug('burgers-and-fries')).toBe(true);
  });

  it('returns true for a slug with digits', () => {
    expect(isValidCategorySlug('combo123')).toBe(true);
  });

  it('returns false for a slug with uppercase letters', () => {
    expect(isValidCategorySlug('Burgers')).toBe(false);
  });

  it('returns false for an empty string', () => {
    expect(isValidCategorySlug('')).toBe(false);
  });

  it('returns false for a slug with spaces', () => {
    expect(isValidCategorySlug('hot dogs')).toBe(false);
  });
});
