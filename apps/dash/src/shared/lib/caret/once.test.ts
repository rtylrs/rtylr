import { describe, expect, it, vi } from 'vitest';
import { once } from './once';

describe('once', () => {
  it('invokes the underlying function only the first time', () => {
    const spy = vi.fn(() => 42);
    const w = once(spy);
    expect(w()).toBe(42);
    expect(w()).toBe(42);
    expect(spy).toHaveBeenCalledTimes(1);
  });
});
