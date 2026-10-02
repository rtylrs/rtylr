import { describe, it, expect, vi, afterEach } from 'vitest';
import { resolveDateRange } from './dateRange';

describe('resolveDateRange', () => {
  const FIXED_NOW = new Date('2026-06-01T12:00:00Z');

  afterEach(() => {
    vi.useRealTimers();
  });

  it('uses provided from and to when both are set', () => {
    const result = resolveDateRange({ from: '2026-01-01', to: '2026-01-31' });
    expect(result.from).toBe('2026-01-01');
    expect(result.to).toBe('2026-01-31');
  });

  it('defaults to today when to is omitted', () => {
    vi.setSystemTime(FIXED_NOW);
    const result = resolveDateRange({ from: '2026-05-01' });
    expect(result.to).toBe('2026-06-01');
  });

  it('defaults from to 30 days ago when from is omitted', () => {
    vi.setSystemTime(FIXED_NOW);
    const result = resolveDateRange({ to: '2026-06-01' });
    expect(result.from).toBe('2026-05-02');
  });

  it('defaults both from and to when params is undefined', () => {
    vi.setSystemTime(FIXED_NOW);
    const result = resolveDateRange(undefined);
    expect(result.to).toBe('2026-06-01');
    expect(result.from).toBe('2026-05-02');
  });

  it('defaults both from and to when params is empty object', () => {
    vi.setSystemTime(FIXED_NOW);
    const result = resolveDateRange({});
    expect(result.to).toBe('2026-06-01');
    expect(result.from).toBe('2026-05-02');
  });

  it('uses the local calendar day, not the UTC day, just after local midnight', () => {
    // 00:30 local on 1 June. At UTC+3 that is 21:30 UTC on 31 May, which is
    // where the UTC getters used to put the boundary (rtylr#1201).
    vi.setSystemTime(new Date(2026, 5, 1, 0, 30, 0));
    const result = resolveDateRange(undefined);
    expect(result.to).toBe('2026-06-01');
  });

  it('returns ISO date strings in YYYY-MM-DD format', () => {
    const result = resolveDateRange({ from: '2025-12-31', to: '2026-01-01' });
    expect(result.from).toMatch(/^\d{4}-\d{2}-\d{2}$/);
    expect(result.to).toMatch(/^\d{4}-\d{2}-\d{2}$/);
  });
});
