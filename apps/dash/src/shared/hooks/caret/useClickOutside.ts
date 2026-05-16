import { useEffect, useRef } from 'react';

/** Call `onOutside` when the user clicks outside the returned ref. */
export function useClickOutside<T extends HTMLElement>(
  onOutside: () => void
): React.RefObject<T | null> {
  const ref = useRef<T>(null);
  useEffect(() => {
    const onMouseDown = (e: MouseEvent): void => {
      const root = ref.current;
      if (!root) return;
      if (root.contains(e.target as Node)) return;
      onOutside();
    };
    document.addEventListener('mousedown', onMouseDown);
    return (): void => document.removeEventListener('mousedown', onMouseDown);
  }, [onOutside]);
  return ref;
}
