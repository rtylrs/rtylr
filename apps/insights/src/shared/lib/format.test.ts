import { describe, it, expect, vi, afterEach } from 'vitest';
import { formatCurrency, formatNumber, formatPercent, timeAgo } from './format';

describe('formatCurrency', () => {
  it('formats a positive integer as USD currency', () => {
    expect(formatCurrency(1000)).toBe('$1,000');
  });

  it('formats zero as $0', () => {
    expect(formatCurrency(0)).toBe('$0');
  });

  it('formats a large value with comma separators', () => {
    expect(formatCurrency(1000000)).toBe('$1,000,000');
  });
});

describe('formatNumber', () => {
  it('formats an integer with comma separators', () => {
    expect(formatNumber(1000)).toBe('1,000');
  });

  it('formats zero as "0"', () => {
    expect(formatNumber(0)).toBe('0');
  });

  it('formats a large number correctly', () => {
    expect(formatNumber(1000000)).toBe('1,000,000');
  });
});

describe('formatPercent', () => {
  it('prefixes positive values with a + sign', () => {
    expect(formatPercent(5.5)).toBe('+5.5%');
  });

  it('prefixes negative values with a - sign (no +)', () => {
    expect(formatPercent(-3.2)).toBe('-3.2%');
  });

  it('formats zero with + sign', () => {
    expect(formatPercent(0)).toBe('+0.0%');
  });

  it('rounds to one decimal place', () => {
    expect(formatPercent(1.234)).toBe('+1.2%');
  });
});

describe('timeAgo', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('returns "Just now" for dates less than 1 minute ago', () => {
    const now = new Date();
    const date = new Date(now.getTime() - 30 * 1000).toISOString();
    vi.setSystemTime(now);
    expect(timeAgo(date)).toBe('Just now');
  });

  it('returns minutes for dates less than 1 hour ago', () => {
    const now = new Date();
    const date = new Date(now.getTime() - 15 * 60 * 1000).toISOString();
    vi.setSystemTime(now);
    expect(timeAgo(date)).toBe('15m ago');
  });

  it('returns hours for dates less than 24 hours ago', () => {
    const now = new Date();
    const date = new Date(now.getTime() - 3 * 60 * 60 * 1000).toISOString();
    vi.setSystemTime(now);
    expect(timeAgo(date)).toBe('3h ago');
  });

  it('returns days for dates less than 30 days ago', () => {
    const now = new Date();
    const date = new Date(now.getTime() - 5 * 24 * 60 * 60 * 1000).toISOString();
    vi.setSystemTime(now);
    expect(timeAgo(date)).toBe('5d ago');
  });

  it('returns months for dates 30+ days ago', () => {
    const now = new Date();
    const date = new Date(now.getTime() - 60 * 24 * 60 * 60 * 1000).toISOString();
    vi.setSystemTime(now);
    expect(timeAgo(date)).toBe('2mo ago');
  });
});
