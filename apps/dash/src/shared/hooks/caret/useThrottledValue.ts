import { useEffect, useRef, useState } from 'react';

/** Returns `value` throttled so it updates at most once per `delay`ms. */
export function useThrottledValue<T>(value: T, delay = 100): T {
  const [throttled, setThrottled] = useState(value);
  const last = useRef(Date.now());

  useEffect(() => {
    const elapsed = Date.now() - last.current;
    if (elapsed >= delay) {
      last.current = Date.now();
      setThrottled(value);
      return;
    }
    const id = setTimeout(() => {
      last.current = Date.now();
      setThrottled(value);
    }, delay - elapsed);
    return (): void => clearTimeout(id);
  }, [value, delay]);

  return throttled;
}
