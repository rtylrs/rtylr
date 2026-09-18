import { useEffect, useRef } from 'react';

/**
 * Attach an event listener to a target (defaults to window) and
 * clean it up when the component unmounts or dependencies change.
 */
export function useEventListener<K extends keyof WindowEventMap>(
  type: K,
  handler: (event: WindowEventMap[K]) => void,
  target: EventTarget = window,
  options?: AddEventListenerOptions
): void {
  const savedHandler = useRef(handler);

  useEffect(() => {
    savedHandler.current = handler;
  }, [handler]);

  useEffect(() => {
    const listener = (event: Event) => savedHandler.current(event as WindowEventMap[K]);
    target.addEventListener(type, listener, options);
    return () => target.removeEventListener(type, listener, options);
  }, [type, target, options]);
}
