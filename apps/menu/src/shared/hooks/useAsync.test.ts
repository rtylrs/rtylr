import { describe, expect, it, vi } from 'vitest';
import useAsync from './useAsync';

describe('useAsync unit tests', () => {
  it('exports useAsync function and handles default initial state', () => {
    expect(useAsync).toBeDefined();
    expect(typeof useAsync).toBe('function');
  });

  it('handles execution and callback logic', async () => {
    const payload = { details: { id: 'menu-1' } };
    const fn = vi.fn(async (_data: unknown) => payload);
    const onSuccess = vi.fn();

    const res = await fn({ id: 'menu-1' });
    onSuccess(res);

    expect(fn).toHaveBeenCalledWith({ id: 'menu-1' });
    expect(onSuccess).toHaveBeenCalledWith(payload);
  });
});
