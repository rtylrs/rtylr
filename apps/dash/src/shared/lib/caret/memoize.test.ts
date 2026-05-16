import { describe, expect, it, vi } from 'vitest';
import { memoize } from './memoize';

describe('memoize', () => {
  it('caches results by argument identity', () => {
    const spy = vi.fn((a: number, b: number) => a + b);
    const m = memoize(spy);
    expect(m(1, 2)).toBe(3);
    expect(m(1, 2)).toBe(3);
    expect(spy).toHaveBeenCalledTimes(1);
  });

  it('respects a custom keyOf', () => {
    const spy = vi.fn((x: { id: number }) => x.id);
    const m = memoize(spy, (x) => String(x.id));
    m({ id: 1 });
    m({ id: 1 });
    expect(spy).toHaveBeenCalledTimes(1);
  });
});
