import { useEffect, useRef } from 'react';

/** Returns the value from the previous render. Useful for transition logic. */
export function usePreviousValue<T>(value: T): T | undefined {
  const ref = useRef<T | undefined>(undefined);
  useEffect(() => {
    ref.current = value;
  }, [value]);
  return ref.current;
}
