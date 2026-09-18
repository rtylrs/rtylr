import { describe, it, expect } from 'vitest';
import {
  formatCurrency,
  formatPercent,
  formatDate,
  formatTime,
  formatShortDate,
  clamp,
} from './format';

describe('formatCurrency', () => {
  it('formats a positive amount as USD by default', () => {
    expect(formatCurrency(12.5)).toBe('$12.50');
  });

  it('formats zero', () => {
    expect(formatCurrency(0)).toBe('$0.00');
  });

  it('formats a negative amount', () => {
    expect(formatCurrency(-5.99)).toBe('-$5.99');
  });

  it('falls back gracefully on invalid currency', () => {
    const result = formatCurrency(10, 'INVALID');
    expect(typeof result).toBe('string');
    expect(result).toMatch(/10/);
  });

  it('formats with an explicit currency code', () => {
    expect(formatCurrency(100, 'USD')).toBe('$100.00');
  });
});

describe('formatPercent', () => {
  it('converts 0.1 to 10% with 0 digits', () => {
    expect(formatPercent(0.1)).toBe('10%');
  });

  it('converts 0.175 to 17.5% with 1 digit', () => {
    expect(formatPercent(0.175, 1)).toBe('17.5%');
  });

  it('converts 1 to 100%', () => {
    expect(formatPercent(1)).toBe('100%');
  });

  it('converts 0 to 0%', () => {
    expect(formatPercent(0)).toBe('0%');
  });

  it('respects the digits parameter for rounding', () => {
    expect(formatPercent(0.1234, 2)).toBe('12.34%');
  });
});

describe('formatDate', () => {
  it('returns a non-empty string for a valid ISO date', () => {
    const result = formatDate('2024-01-15T10:30:00Z');
    expect(typeof result).toBe('string');
    expect(result.length).toBeGreaterThan(0);
  });

  it('includes the year for a valid ISO date', () => {
    const result = formatDate('2024-01-15T10:30:00Z');
    expect(result).toMatch(/2024/);
  });
});

describe('formatTime', () => {
  it('returns a string in HH:MM format for a valid ISO date', () => {
    const result = formatTime('2024-01-15T10:30:00Z');
    expect(typeof result).toBe('string');
    expect(result).toMatch(/\d{1,2}:\d{2}/);
  });

  it('returns a non-empty string for any valid ISO string', () => {
    const result = formatTime('2024-06-01T09:05:00Z');
    expect(result.length).toBeGreaterThan(0);
  });
});

describe('formatShortDate', () => {
  it('returns a string for a valid ISO date', () => {
    const result = formatShortDate('2024-06-01T00:00:00Z');
    expect(typeof result).toBe('string');
    expect(result.length).toBeGreaterThan(0);
  });

  it('includes date digits for a valid ISO date', () => {
    const result = formatShortDate('2024-06-01T00:00:00Z');
    expect(result).toMatch(/\d/);
  });
});

describe('clamp', () => {
  it('returns the value when within range', () => {
    expect(clamp(5, 0, 10)).toBe(5);
  });

  it('clamps to min when value is below range', () => {
    expect(clamp(-1, 0, 10)).toBe(0);
  });

  it('clamps to max when value is above range', () => {
    expect(clamp(15, 0, 10)).toBe(10);
  });

  it('returns min when value equals min', () => {
    expect(clamp(0, 0, 10)).toBe(0);
  });

  it('returns max when value equals max', () => {
    expect(clamp(10, 0, 10)).toBe(10);
  });

  it('works with negative ranges', () => {
    expect(clamp(-5, -10, -1)).toBe(-5);
  });
});
