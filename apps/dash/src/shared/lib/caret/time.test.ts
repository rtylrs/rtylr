import { describe, expect, it } from 'vitest';
import { relativeTime } from './time';

const NOW = new Date('2026-05-16T12:00:00Z').getTime();

describe('relativeTime', () => {
  it('returns "just now" for under a minute', () => {
    const t = new Date(NOW - 30_000).toISOString();
    expect(relativeTime(t, NOW)).toBe('just now');
  });

  it('returns minutes for under an hour', () => {
    const t = new Date(NOW - 5 * 60_000).toISOString();
    expect(relativeTime(t, NOW)).toBe('5m ago');
  });

  it('returns hours for under a day', () => {
    const t = new Date(NOW - 3 * 3_600_000).toISOString();
    expect(relativeTime(t, NOW)).toBe('3h ago');
  });

  it('returns days for under a week', () => {
    const t = new Date(NOW - 4 * 86_400_000).toISOString();
    expect(relativeTime(t, NOW)).toBe('4d ago');
  });

  it('falls back to a localized date past a week', () => {
    const t = new Date(NOW - 14 * 86_400_000).toISOString();
    expect(relativeTime(t, NOW)).not.toContain('ago');
  });

  it('handles unparseable inputs gracefully', () => {
    expect(relativeTime('not-a-date', NOW)).toBe('');
  });
});
