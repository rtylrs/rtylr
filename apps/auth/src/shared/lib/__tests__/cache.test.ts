import { describe, it, expect, vi, afterEach } from 'vitest';
import { TtlCache } from '../cache';

afterEach(() => {
  vi.useRealTimers();
});

describe('TtlCache', () => {
  it('stores and retrieves a value', () => {
    const cache = new TtlCache<string>(1000);
    cache.set('key', 'value');
    expect(cache.get('key')).toBe('value');
  });

  it('returns undefined for missing keys', () => {
    const cache = new TtlCache<string>();
    expect(cache.get('missing')).toBeUndefined();
  });

  it('expires entries after TTL', () => {
    vi.useFakeTimers();
    const cache = new TtlCache<string>(500);
    cache.set('key', 'value');
    vi.advanceTimersByTime(600);
    expect(cache.get('key')).toBeUndefined();
  });

  it('has returns true for live entries', () => {
    const cache = new TtlCache<number>(5000);
    cache.set('n', 42);
    expect(cache.has('n')).toBe(true);
  });

  it('has returns false for expired entries', () => {
    vi.useFakeTimers();
    const cache = new TtlCache<number>(100);
    cache.set('n', 42);
    vi.advanceTimersByTime(200);
    expect(cache.has('n')).toBe(false);
  });

  it('delete removes an entry', () => {
    const cache = new TtlCache<string>();
    cache.set('x', 'y');
    cache.delete('x');
    expect(cache.get('x')).toBeUndefined();
  });

  it('evict removes expired entries and returns count', () => {
    vi.useFakeTimers();
    const cache = new TtlCache<string>(200);
    cache.set('a', '1');
    cache.set('b', '2');
    vi.advanceTimersByTime(300);
    const evicted = cache.evict();
    expect(evicted).toBe(2);
    expect(cache.size).toBe(0);
  });

  it('clear empties all entries', () => {
    const cache = new TtlCache<string>();
    cache.set('a', '1');
    cache.set('b', '2');
    cache.clear();
    expect(cache.size).toBe(0);
  });

  it('supports per-entry TTL override', () => {
    vi.useFakeTimers();
    const cache = new TtlCache<string>(5000);
    cache.set('short', 'val', 100);
    vi.advanceTimersByTime(200);
    expect(cache.get('short')).toBeUndefined();
  });
});
