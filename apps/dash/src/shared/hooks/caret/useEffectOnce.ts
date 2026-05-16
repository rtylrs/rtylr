import { useEffect } from 'react';

/** Run an effect exactly once on mount. Equivalent to useEffect(fn, []). */
export function useEffectOnce(fn: () => void | (() => void)): void {
  // eslint-disable-next-line react-hooks/exhaustive-deps
  useEffect(fn, []);
}
