import { describe, it, expect } from 'vitest';
import { formatDate, formatRelativeTime, formatCurrency, formatNumber } from './formatting';

describe('formatDate', () => {
  it('formats a valid ISO string', () => {
    expect(formatDate('2024-01-15')).toBe('Jan 15, 2024');
  });
  it('formats a Date object', () => {
    expect(formatDate(new Date(2024, 0, 15))).toBe('Jan 15, 2024');
  });
  it('returns Invalid Date for bad input', () => {
    expect(formatDate('not-a-date')).toBe('Invalid Date');
  });
});

describe('formatCurrency', () => {
  it('formats USD by default', () => {
    expect(formatCurrency(1234.56)).toBe('$1,234.56');
  });
  it('formats zero dollars', () => {
    expect(formatCurrency(0)).toBe('$0.00');
  });
  it('formats negative amounts', () => {
    expect(formatCurrency(-50)).toContain('50');
  });
});

describe('formatNumber', () => {
  it('adds thousand separators', () => {
    expect(formatNumber(1234567)).toBe('1,234,567');
  });
  it('formats zero', () => {
    expect(formatNumber(0)).toBe('0');
  });
});

describe('formatRelativeTime', () => {
  it('returns just now for very recent dates', () => {
    expect(formatRelativeTime(new Date().toISOString())).toBe('just now');
  });
  it('returns minutes ago', () => {
    const d = new Date(Date.now() - 5 * 60 * 1000);
    expect(formatRelativeTime(d)).toBe('5 minutes ago');
  });
  it('returns hours ago', () => {
    const d = new Date(Date.now() - 3 * 60 * 60 * 1000);
    expect(formatRelativeTime(d)).toBe('3 hours ago');
  });
});
