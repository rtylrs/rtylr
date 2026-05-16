import { useEffect, useRef } from 'react';

type Listener<E> = (event: E) => void;

/** Bind an event listener to `target` (defaults to window). */
export function useEventListener<K extends keyof WindowEventMap>(
  event: K,
  handler: Listener<WindowEventMap[K]>,
  target?: Window | Document | HTMLElement | null
): void {
  const handlerRef = useRef(handler);
  useEffect(() => {
    handlerRef.current = handler;
  }, [handler]);

  useEffect(() => {
    const tgt = target ?? (typeof window !== 'undefined' ? window : null);
    if (!tgt) return;
    const wrapped = (e: Event): void => handlerRef.current(e as WindowEventMap[K]);
    tgt.addEventListener(event, wrapped as EventListener);
    return (): void => tgt.removeEventListener(event, wrapped as EventListener);
  }, [event, target]);
}
