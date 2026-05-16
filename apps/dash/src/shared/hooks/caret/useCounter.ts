import { useCallback, useState } from 'react';

/** Number state with stable inc/dec/reset helpers. */
export function useCounter(initial = 0): {
  value: number;
  inc: (delta?: number) => void;
  dec: (delta?: number) => void;
  set: (next: number) => void;
  reset: () => void;
} {
  const [value, setValue] = useState(initial);
  const inc = useCallback((delta = 1) => setValue((v) => v + delta), []);
  const dec = useCallback((delta = 1) => setValue((v) => v - delta), []);
  const reset = useCallback(() => setValue(initial), [initial]);
  return { value, inc, dec, set: setValue, reset };
}
