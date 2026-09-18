import { describe, expect, it, vi } from 'vitest';
import { runExclusiveDrain } from './drainGuard';

describe('runExclusiveDrain', () => {
  it('runs the function when no drain is active', async () => {
    const fn = vi.fn(async () => 'result');
    await expect(runExclusiveDrain('k', fn)).resolves.toBe('result');
    expect(fn).toHaveBeenCalledOnce();
  });

  it('shares a single in-flight run between concurrent triggers', async () => {
    let release: (value: number) => void = () => undefined;
    const fn = vi.fn(
      () =>
        new Promise<number>((resolve) => {
          release = resolve;
        })
    );

    const first = runExclusiveDrain('shared', fn);
    const second = runExclusiveDrain('shared', fn);
    const third = runExclusiveDrain('shared', fn);

    release(7);
    await Promise.all([first, second, third]);

    expect(fn).toHaveBeenCalledOnce();
  });

  it('allows a new run after the previous one settles', async () => {
    const fn = vi.fn(async () => undefined);
    await runExclusiveDrain('reuse', fn);
    await runExclusiveDrain('reuse', fn);
    expect(fn).toHaveBeenCalledTimes(2);
  });

  it('clears the guard even when the run rejects', async () => {
    const failing = vi.fn(async () => {
      throw new Error('boom');
    });
    await expect(runExclusiveDrain('rejecting', failing)).rejects.toThrow('boom');

    const next = vi.fn(async () => 'ok');
    await expect(runExclusiveDrain('rejecting', next)).resolves.toBe('ok');
  });

  it('keys guards independently per queue type', async () => {
    let releaseA: () => void = () => undefined;
    const a = vi.fn(
      () =>
        new Promise<void>((resolve) => {
          releaseA = resolve;
        })
    );
    const b = vi.fn(async () => 'b');

    const first = runExclusiveDrain('a', a);
    await expect(runExclusiveDrain('b', b)).resolves.toBe('b');
    releaseA();
    await first;

    expect(a).toHaveBeenCalledOnce();
    expect(b).toHaveBeenCalledOnce();
  });
});
