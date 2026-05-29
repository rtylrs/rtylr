import { describe, it, expect } from 'vitest';
import {
  formatNumber,
  formatCurrency,
  formatPercent,
  parseNumeric,
  isValidNumber,
} from '../number';

describe('formatNumber', () => {
  it('formats an integer', () => {
    expect(formatNumber(1000)).toBe('1,000');
  });

  it('formats a decimal', () => {
    expect(formatNumber(3.14)).toBe('3.14');
  });
});

describe('formatCurrency', () => {
  it('formats as USD by default', () => {
    expect(formatCurrency(9.99)).toBe('$9.99');
  });
});

describe('formatPercent', () => {
  it('formats a decimal as percent', () => {
    expect(formatPercent(0.5)).toBe('50%');
  });
});

describe('parseNumeric', () => {
  it('parses a plain number string', () => {
    expect(parseNumeric('42')).toBe(42);
  });

  it('strips non-numeric characters', () => {
    expect(parseNumeric('$1,234.56')).toBe(1234.56);
  });

  it('returns NaN for non-numeric input', () => {
    expect(parseNumeric('abc')).toBeNaN();
  });
});

describe('isValidNumber', () => {
  it('returns true for finite numbers', () => {
    expect(isValidNumber(42)).toBe(true);
  });

  it('returns false for NaN', () => {
    expect(isValidNumber(NaN)).toBe(false);
  });

  it('returns false for Infinity', () => {
    expect(isValidNumber(Infinity)).toBe(false);
  });

  it('returns false for strings', () => {
    expect(isValidNumber('42')).toBe(false);
  });
});
