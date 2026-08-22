import { describe, it, expect } from 'vitest';
import {
  formatCurrency,
  formatPercent,
  formatNumber,
  formatDate,
  getInitials,
  truncate,
} from './formatters';

describe('formatCurrency', () => {
  it('formats a positive amount in USD', () => {
    expect(formatCurrency(1234.5)).toBe('$1,234.50');
  });
  it('formats zero as $0.00', () => {
    expect(formatCurrency(0)).toBe('$0.00');
  });
  it('formats a negative amount', () => {
    expect(formatCurrency(-99.9)).toBe('-$99.90');
  });
});

describe('formatPercent', () => {
  it('formats 50 as 50.0%', () => {
    expect(formatPercent(50)).toBe('50.0%');
  });
  it('formats 0 as 0.0%', () => {
    expect(formatPercent(0)).toBe('0.0%');
  });
  it('respects custom decimal places', () => {
    expect(formatPercent(33.3333, 2)).toBe('33.33%');
  });
  it('uses 0 decimals when specified', () => {
    expect(formatPercent(75.9, 0)).toBe('76%');
  });
});

describe('formatNumber', () => {
  it('formats 1000 with a comma separator', () => {
    expect(formatNumber(1000)).toBe('1,000');
  });
  it('formats 0 as "0"', () => {
    expect(formatNumber(0)).toBe('0');
  });
});

describe('formatDate', () => {
  it('returns empty string for null', () => {
    expect(formatDate(null)).toBe('');
  });
  it('returns empty string for an invalid date string', () => {
    expect(formatDate('not-a-date')).toBe('');
  });
  it('formats a valid ISO date string', () => {
    expect(formatDate('2024-01-15')).toBe('Jan 15, 2024');
  });
});

describe('getInitials', () => {
  it('returns two uppercase initials for a full name', () => {
    expect(getInitials('Jane Doe')).toBe('JD');
  });
  it('returns one initial for a single name', () => {
    expect(getInitials('Alice')).toBe('A');
  });
  it('uses first and last word for multi-word names', () => {
    expect(getInitials('Mary Ann Smith')).toBe('MS');
  });
  it('returns empty string for blank input', () => {
    expect(getInitials('')).toBe('');
  });
});

describe('truncate', () => {
  it('returns the string unchanged when shorter than maxLength', () => {
    expect(truncate('hello', 10)).toBe('hello');
  });
  it('truncates and appends the default suffix', () => {
    expect(truncate('hello world', 8)).toBe('hello...');
  });
  it('handles an empty string', () => {
    expect(truncate('', 5)).toBe('');
  });
});
