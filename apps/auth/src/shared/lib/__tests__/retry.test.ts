import { describe, it, expect, vi, afterEach } from 'vitest';
import { withExponentialBackoff, withFixedRetry, withCooldown } from '../retry';

afterEach(() => {
  vi.useRealTimers();
});

describe('withExponentialBackoff', () => {
  it('returns result on first success', async () => {
    const result = await withExponentialBackoff(async () => 42, { baseDelayMs: 0 });
    expect(result).toBe(42);
  });

  it('retries and succeeds', async () => {
    let calls = 0;
    const result = await withExponentialBackoff(
      async () => {
        calls++;
        if (calls < 3) throw new Error('fail');
        return 'ok';
      },
      { baseDelayMs: 0, maxAttempts: 3 }
    );
    expect(result).toBe('ok');
    expect(calls).toBe(3);
  });

  it('throws after all attempts fail', async () => {
    await expect(
      withExponentialBackoff(
        async () => {
          throw new Error('always fails');
        },
        { baseDelayMs: 0, maxAttempts: 2 }
      )
    ).rejects.toThrow('always fails');
  });

  it('stops early when shouldRetry returns false', async () => {
    let calls = 0;
    await expect(
      withExponentialBackoff(
        async () => {
          calls++;
          throw new Error('stop');
        },
        { baseDelayMs: 0, maxAttempts: 5, shouldRetry: () => false }
      )
    ).rejects.toThrow('stop');
    expect(calls).toBe(1);
  });
});

describe('withFixedRetry', () => {
  it('returns result on success', async () => {
    expect(await withFixedRetry(async () => 'hello', 3, 0)).toBe('hello');
  });

  it('retries and returns eventual success', async () => {
    let n = 0;
    const result = await withFixedRetry(
      async () => {
        n++;
        if (n < 2) throw new Error('x');
        return n;
      },
      3,
      0
    );
    expect(result).toBe(2);
  });

  it('throws after all retries fail', async () => {
    await expect(
      withFixedRetry(
        async () => {
          throw new Error('oops');
        },
        2,
        0
      )
    ).rejects.toThrow('oops');
  });
});

describe('withCooldown', () => {
  it('calls the function on first invocation', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const cooled = withCooldown(fn, 500);
    cooled();
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('ignores calls within cooldown', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const cooled = withCooldown(fn, 500);
    cooled();
    cooled();
    cooled();
    expect(fn).toHaveBeenCalledTimes(1);
  });

  it('allows call after cooldown expires', () => {
    vi.useFakeTimers();
    const fn = vi.fn();
    const cooled = withCooldown(fn, 500);
    cooled();
    vi.advanceTimersByTime(600);
    cooled();
    expect(fn).toHaveBeenCalledTimes(2);
  });
});
