import { describe, expect, it, vi, afterEach } from 'vitest';
import { isExpired } from './documentUtils';

describe('isExpired', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('treats a document expiring today as valid', () => {
    vi.setSystemTime(new Date('2026-06-15T10:00:00'));
    expect(isExpired('2026-06-15')).toBe(false);
  });

  it('treats a document expiring yesterday as expired', () => {
    vi.setSystemTime(new Date('2026-06-15T10:00:00'));
    expect(isExpired('2026-06-14')).toBe(true);
  });

  it('treats a document expiring tomorrow as valid', () => {
    vi.setSystemTime(new Date('2026-06-15T10:00:00'));
    expect(isExpired('2026-06-16')).toBe(false);
  });

  it('ignores any time component on the expiry timestamp', () => {
    vi.setSystemTime(new Date('2026-06-15T23:59:00'));
    expect(isExpired('2026-06-15T08:00:00Z')).toBe(false);
  });
});
