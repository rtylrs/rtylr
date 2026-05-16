import { useEffect, useRef } from 'react';

/** Focus the returned ref on mount. */
export function useFocusOnMount<T extends HTMLElement>(active = true): React.RefObject<T | null> {
  const ref = useRef<T>(null);
  useEffect(() => {
    if (active) ref.current?.focus();
  }, [active]);
  return ref;
}
