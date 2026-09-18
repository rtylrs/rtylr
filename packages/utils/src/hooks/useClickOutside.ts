import { useEffect, RefObject } from 'react';

export function useClickOutside<T extends HTMLElement = HTMLElement>(
  ref: RefObject<T | null>,
  handler: (event: MouseEvent | TouchEvent) => void
): void {
  useEffect(() => {
    const listener = (event: MouseEvent | TouchEvent) => {
      const el = ref?.current;
      if (!el) return;

      // event.target is typed EventTarget, which is not necessarily a Node -
      // window and XMLHttpRequest are both EventTargets. Narrowing with
      // instanceof rather than asserting keeps contains() honest, and a target
      // that is not a Node cannot be inside the element, so it counts as outside.
      const target = event.target;
      if (target instanceof Node && el.contains(target)) {
        return;
      }

      handler(event);
    };

    document.addEventListener('mousedown', listener);
    document.addEventListener('touchstart', listener);

    return () => {
      document.removeEventListener('mousedown', listener);
      document.removeEventListener('touchstart', listener);
    };
  }, [ref, handler]);
}
