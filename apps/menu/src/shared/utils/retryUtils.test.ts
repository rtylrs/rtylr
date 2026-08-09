import { describe, it, expect, vi } from 'vitest';
import { backoffDelay, withRetry, once } from './retryUtils';

describe('backoffDelay', () => {
  it('returns the base delay on the first attempt', () => {
    expect(backoffDelay(0, 500, 30_000, false)).toBe(500);
  });

  it('doubles on each attempt', () => {
    expect(backoffDelay(1, 500, 30_000, false)).toBe(1000);
    expect(backoffDelay(2, 500, 30_000, false)).toBe(2000);
  });

  it('clamps at maxMs', () => {
    expect(backoffDelay(10, 500, 1000, false)).toBe(1000);
  });
});

describe('withRetry', () => {
  it('returns the result on first success', async () => {
    const fn = vi.fn().mockResolvedValue(42);
    const result = await withRetry(fn, 3, 0);
    expect(result).toBe(42);
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('retries on failure and eventually succeeds', async () => {
    const fn = vi.fn().mockRejectedValueOnce(new Error('fail')).mockResolvedValue('ok');
    const result = await withRetry(fn, 3, 0);
    expect(result).toBe('ok');
    expect(fn).toHaveBeenCalledTimes(2);
  });

  it('throws after all attempts fail', async () => {
    const fn = vi.fn().mockRejectedValue(new Error('always fails'));
    await expect(withRetry(fn, 3, 0)).rejects.toThrow('always fails');
    expect(fn).toHaveBeenCalledTimes(3);
  });
});

describe('once', () => {
  it('calls the function only once', () => {
    const fn = vi.fn().mockReturnValue(99);
    const onced = once(fn);
    expect(onced()).toBe(99);
    expect(onced()).toBe(99);
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('returns the same result on subsequent calls', () => {
    let count = 0;
    const onced = once(() => ++count);
    expect(onced()).toBe(1);
    expect(onced()).toBe(1);
    expect(count).toBe(1);
  });
});
