import { act, renderHook, waitFor } from '@testing-library/react';
import { describe, expect, it, vi } from 'vitest';
import useMenuAsync from './useAsync';

type HookFactory = typeof useMenuAsync;

describe.each([['menu', useMenuAsync]] satisfies Array<[string, HookFactory]>)(
  '%s useAsync',
  (_label, useAsync) => {
    it('passes the resolved payload to onSuccess and stores the request data', async () => {
      const payload = { details: { id: 'menu-1' } };
      const fn = vi.fn(async (_data: unknown) => payload);
      const onSuccess = vi.fn();

      const { result } = renderHook(() =>
        useAsync({
          fn,
          onSuccess,
        })
      );

      act(() => {
        result.current.main({ id: 'menu-1' });
      });

      await waitFor(() => expect(onSuccess).toHaveBeenCalledWith(payload));
      expect(fn).toHaveBeenCalledWith({ id: 'menu-1' });
      expect(result.current.result).toEqual(payload);
      expect(result.current.requestData).toEqual({ id: 'menu-1' });
      expect(result.current.error).toBeNull();
    });

    it('re-runs the mount call when callOnMountData changes', async () => {
      const fn = vi.fn(async (data: unknown) => ({ data }));

      const { rerender } = renderHook(
        ({ value }) =>
          useAsync({
            fn,
            callOnMount: true,
            callOnMountData: value,
          }),
        {
          initialProps: { value: 'first' },
        }
      );

      await waitFor(() => expect(fn).toHaveBeenCalledWith('first'));

      rerender({ value: 'second' });

      await waitFor(() => expect(fn).toHaveBeenCalledWith('second'));
    });
  }
);
