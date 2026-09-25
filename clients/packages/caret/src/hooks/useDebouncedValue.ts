import { useEffect, useState } from 'react';

/** Returns `value` debounced by `delay`ms - handy for search inputs. */
export function useDebouncedValue<T>(value: T, delay = 200): T {
  const [debounced, setDebounced] = useState(value);

  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), delay);
    return (): void => clearTimeout(id);
  }, [value, delay]);

  return debounced;
}
