import { describe, expect, it } from 'vitest';
import { formatPercent } from './percent';

describe('formatPercent', () => {
  it('formats with zero decimals by default', () => {
    expect(formatPercent(0.5)).toBe('50%');
  });
  it('accepts a digit count', () => {
    expect(formatPercent(0.123, 1)).toBe('12.3%');
  });
  it('handles non-finite values', () => {
    expect(formatPercent(NaN)).toBe('-');
  });
});
