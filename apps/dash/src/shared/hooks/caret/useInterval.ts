import { useEffect, useRef } from 'react';

/** Call `fn` every `ms` ms. Pass `null` to pause. */
export function useInterval(fn: () => void, ms: number | null): void {
  const ref = useRef(fn);
  useEffect(() => {
    ref.current = fn;
  }, [fn]);
  useEffect(() => {
    if (ms == null) return;
    const id = setInterval(() => ref.current(), ms);
    return (): void => clearInterval(id);
  }, [ms]);
}
