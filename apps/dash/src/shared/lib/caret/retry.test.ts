import { describe, expect, it, vi } from 'vitest';
import { retry } from './retry';

describe('retry', () => {
  it('returns on first success', async () => {
    const spy = vi.fn(() => Promise.resolve(7));
    await expect(retry(spy)).resolves.toBe(7);
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it('retries up to the limit then rethrows', async () => {
    const spy = vi.fn(() => Promise.reject(new Error('bad')));
    await expect(retry(spy, { attempts: 2, baseMs: 0 })).rejects.toThrow('bad');
    expect(spy).toHaveBeenCalledTimes(2);
  });

  it('eventually returns when the function recovers', async () => {
    let calls = 0;
    const spy = (): Promise<number> => {
      calls++;
      return calls < 2 ? Promise.reject(new Error('again')) : Promise.resolve(42);
    };
    await expect(retry(spy, { attempts: 3, baseMs: 0 })).resolves.toBe(42);
  });
});
