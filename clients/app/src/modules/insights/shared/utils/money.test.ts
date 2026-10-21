import { describe, it, expect } from 'vitest';
import { minorToMajor, formatMoney } from './money';

describe('minorToMajor', () => {
  it('divides cents by 100 for USD', () => {
    expect(minorToMajor(1000, 'USD')).toBe(10);
  });

  it('divides by 100 for EUR', () => {
    expect(minorToMajor(5050, 'EUR')).toBe(50.5);
  });

  it('returns the value unchanged for JPY (zero-decimal)', () => {
    expect(minorToMajor(500, 'JPY')).toBe(500);
  });

  it('defaults to USD when currency is omitted', () => {
    expect(minorToMajor(200)).toBe(2);
  });

  it('returns 0 for 0 input', () => {
    expect(minorToMajor(0, 'GBP')).toBe(0);
  });

  it('handles negative amounts', () => {
    expect(minorToMajor(-100, 'USD')).toBe(-1);
  });
});

describe('formatMoney', () => {
  it('formats USD cents as a dollar string', () => {
    expect(formatMoney(100, 'USD')).toBe('$1.00');
  });

  it('formats JPY without decimal places', () => {
    const result = formatMoney(1000, 'JPY');
    expect(result).toContain('1,000');
    expect(result).not.toContain('.');
  });

  it('defaults to USD when currency is omitted', () => {
    expect(formatMoney(100)).toContain('$');
  });

  it('falls back to a plain string when locale is invalid', () => {
    const result = formatMoney(100, 'USD', 'INVALID_LOCALE_XYZ' as string);
    expect(result).toContain('1.00');
  });

  it('formats large amounts with correct scaling', () => {
    expect(formatMoney(100000, 'USD')).toBe('$1,000.00');
  });
});
