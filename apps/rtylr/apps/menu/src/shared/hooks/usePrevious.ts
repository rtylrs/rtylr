import { useRef, useEffect } from 'react';

/**
 * Return the previous value of a reactive value from the last render.
 * Returns undefined on the first render.
 */
export function usePrevious<T>(value: T): T | undefined {
  const ref = useRef<T | undefined>(undefined);

  useEffect(() => {
    ref.current = value;
  }, [value]);

  return ref.current;
}
