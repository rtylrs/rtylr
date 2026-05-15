import { useEffect, useRef } from 'react';

/**
 * Returns a ref to attach to the scroll container, and an effect that keeps
 * the container pinned to the bottom while the user hasn't scrolled away.
 * Triggered on changes to `deps`.
 */
export function useScrollToBottom<T extends HTMLElement>(deps: unknown[]): React.RefObject<T> {
  const ref = useRef<T>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    // 64px tolerance: if the user is near the bottom we follow new content,
    // otherwise leave their scroll position alone.
    const distance = el.scrollHeight - el.scrollTop - el.clientHeight;
    if (distance < 96) {
      el.scrollTop = el.scrollHeight;
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, deps);

  return ref;
}
