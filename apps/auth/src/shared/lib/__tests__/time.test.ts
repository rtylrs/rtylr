import { describe, it, expect } from 'vitest';
import { nowSeconds, daysToSeconds, futureTimestamp, isPast, relativeTime } from '../time';

describe('nowSeconds', () => {
  it('returns an integer close to Date.now() / 1000', () => {
    const expected = Math.floor(Date.now() / 1000);
    expect(Math.abs(nowSeconds() - expected)).toBeLessThanOrEqual(1);
  });
});

describe('daysToSeconds', () => {
  it('converts 1 day to 86400 seconds', () => {
    expect(daysToSeconds(1)).toBe(86400);
  });
});

describe('futureTimestamp', () => {
  it('returns a timestamp greater than nowSeconds', () => {
    expect(futureTimestamp(3600)).toBeGreaterThan(nowSeconds());
  });
});

describe('isPast', () => {
  it('returns true for a timestamp in the past', () => {
    expect(isPast(nowSeconds() - 1)).toBe(true);
  });

  it('returns false for a timestamp in the future', () => {
    expect(isPast(nowSeconds() + 3600)).toBe(false);
  });
});

describe('relativeTime', () => {
  const now = 1_700_000_000_000;

  it('returns "just now" for very small differences', () => {
    expect(relativeTime(now + 2000, now)).toBe('just now');
  });

  it('returns seconds for differences under a minute', () => {
    expect(relativeTime(now + 30_000, now)).toBe('in 30 seconds');
  });

  it('returns minutes for differences under an hour', () => {
    expect(relativeTime(now + 120_000, now)).toBe('in 2 minutes');
  });

  it('appends "ago" for past timestamps', () => {
    expect(relativeTime(now - 120_000, now)).toBe('2 minutes ago');
  });
});
