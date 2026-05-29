import { describe, it, expect, vi } from 'vitest';
import { sleep, tryCatch, withRetry, withTimeout } from '../promise';
describe('sleep', () => {
  it('resolves after the given time', async () => {
    vi.useFakeTimers();
    const p = sleep(100);
    vi.advanceTimersByTime(100);
    await expect(p).resolves.toBeUndefined();
    vi.useRealTimers();
  });
});
describe('tryCatch', () => {
  it('returns [value, null] on success', async () => {
    const [val, err] = await tryCatch(Promise.resolve(42));
    expect(val).toBe(42);
    expect(err).toBeNull();
  });
  it('returns [null, Error] on failure', async () => {
    const [val, err] = await tryCatch(Promise.reject(new Error('oops')));
    expect(val).toBeNull();
    expect(err?.message).toBe('oops');
  });
});
describe('withRetry', () => {
  it('returns on first success', async () => {
    expect(await withRetry(vi.fn().mockResolvedValue('ok'), 3, 0)).toBe('ok');
  });
  it('throws after exhausting retries', async () => {
    await expect(withRetry(vi.fn().mockRejectedValue(new Error('always')), 2, 0)).rejects.toThrow(
      'always'
    );
  });
});
describe('withTimeout', () => {
  it('resolves when promise is faster', async () => {
    expect(await withTimeout(Promise.resolve('fast'), 1000)).toBe('fast');
  });
});
