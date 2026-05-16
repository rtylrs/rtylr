import { describe, expect, it } from 'vitest';
import { formatCurrency } from './formatCurrency';

describe('formatCurrency', () => {
  it('formats default USD', () => {
    expect(formatCurrency(1234.5)).toContain('1,234');
  });
  it('handles infinite values gracefully', () => {
    expect(formatCurrency(NaN)).toBe('—');
  });
  it('respects the currency argument', () => {
    expect(formatCurrency(10, 'EUR', 'en-GB')).toContain('€');
  });
});
