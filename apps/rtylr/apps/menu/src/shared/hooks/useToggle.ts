import { useCallback, useState } from 'react';

/**
 * Manages a boolean toggle state with a stable toggle function.
 * Returns [value, toggle, setTrue, setFalse].
 */
export function useToggle(initial = false): [boolean, () => void, () => void, () => void] {
  const [value, setValue] = useState(initial);

  const toggle = useCallback(() => setValue((v) => !v), []);
  const setTrue = useCallback(() => setValue(true), []);
  const setFalse = useCallback(() => setValue(false), []);

  return [value, toggle, setTrue, setFalse];
}
