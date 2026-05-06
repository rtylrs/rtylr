import { describe, expect, it } from 'vitest';
import { formatCurrency } from './formatting';

describe('formatCurrency', () => {
  it('formats amounts with a valid currency code', () => {
    expect(formatCurrency(12.5, 'USD')).toBe('$12.50');
  });

  it('falls back to USD when the currency code is invalid', () => {
    expect(formatCurrency(12.5, '')).toBe('$12.50');
  });
});
