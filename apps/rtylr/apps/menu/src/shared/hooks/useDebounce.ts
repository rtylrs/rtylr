import { useEffect, useState } from 'react';

/**
 * Returns a debounced copy of the given value that only updates after the
 * specified delay (in milliseconds) has elapsed with no new changes.
 */
export function useDebounce<T>(value: T, delayMs: number): T {
  const [debounced, setDebounced] = useState<T>(value);

  useEffect(() => {
    const timer = setTimeout(() => {
      setDebounced(value);
    }, delayMs);

    return () => clearTimeout(timer);
  }, [value, delayMs]);

  return debounced;
}
