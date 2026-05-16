import { useCallback, useState } from 'react';

/** Stateful boolean with stable toggle/set helpers. */
export function useToggle(initial = false): [boolean, () => void, (next: boolean) => void] {
  const [value, setValue] = useState(initial);
  const toggle = useCallback(() => setValue((v) => !v), []);
  return [value, toggle, setValue];
}
