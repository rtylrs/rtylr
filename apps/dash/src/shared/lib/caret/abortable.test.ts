import { describe, expect, it } from 'vitest';
import { abortable } from './abortable';

describe('abortable', () => {
  it('resolves when the underlying promise resolves', async () => {
    const ac = new AbortController();
    await expect(abortable(Promise.resolve(42), ac.signal)).resolves.toBe(42);
  });

  it('rejects with AbortError when cancelled', async () => {
    const ac = new AbortController();
    const p = abortable(new Promise<number>(() => undefined), ac.signal);
    ac.abort();
    await expect(p).rejects.toMatchObject({ name: 'AbortError' });
  });

  it('rejects immediately if signal is already aborted', async () => {
    const ac = new AbortController();
    ac.abort();
    await expect(abortable(Promise.resolve(1), ac.signal)).rejects.toMatchObject({
      name: 'AbortError',
    });
  });
});
