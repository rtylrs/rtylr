import { useEffect, useRef } from 'react';

/**
 * Returns a ref that auto-focuses when the component mounts.
 */
export const useAutoFocus = <T extends HTMLElement = HTMLInputElement>(
  enabled: boolean = true
) => {
  const ref = useRef<T | null>(null);

  useEffect(() => {
    if (enabled && ref.current) {
      ref.current.focus();
    }
  }, [enabled]);

  return ref;
};
