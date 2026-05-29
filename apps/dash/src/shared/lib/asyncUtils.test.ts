import { describe, it, expect, vi, beforeEach } from 'vitest';
import { debounce, throttle, once, memoize } from './asyncUtils';

describe('debounce', () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  it('delays function call', () => {
    const fn = vi.fn();
    const debounced = debounce(fn, 100);
    debounced();
    expect(fn).not.toHaveBeenCalled();
    vi.advanceTimersByTime(100);
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('resets delay on repeated calls', () => {
    const fn = vi.fn();
    const debounced = debounce(fn, 100);
    debounced();
    vi.advanceTimersByTime(50);
    debounced();
    vi.advanceTimersByTime(50);
    expect(fn).not.toHaveBeenCalled();
    vi.advanceTimersByTime(50);
    expect(fn).toHaveBeenCalledTimes(1);
  });
});

describe('throttle', () => {
  it('calls function immediately on first call', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const throttled = throttle(fn, 100);
    throttled();
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('ignores calls within the limit window', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const throttled = throttle(fn, 1000);
    throttled();
    throttled();
    throttled();
    expect(fn).toHaveBeenCalledTimes(1);
  });
});

describe('once', () => {
  it('calls function only once', () => {
    const fn = vi.fn().mockReturnValue(42);
    const onced = once(fn);
    expect(onced()).toBe(42);
    expect(onced()).toBe(42);
    expect(fn).toHaveBeenCalledTimes(1);
  });
});

describe('memoize', () => {
  it('caches results', () => {
    const fn = vi.fn((x: number) => x * 2);
    const memo = memoize(fn);
    expect(memo(5)).toBe(10);
    expect(memo(5)).toBe(10);
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('computes different args separately', () => {
    const fn = vi.fn((x: number) => x * 3);
    const memo = memoize(fn);
    expect(memo(2)).toBe(6);
    expect(memo(4)).toBe(12);
    expect(fn).toHaveBeenCalledTimes(2);
  });
});
