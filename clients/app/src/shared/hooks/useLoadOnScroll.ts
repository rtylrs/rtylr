import { useEffect, useEffectEvent, useState } from 'react';
import { scrollParent } from '../lib/scrollParent';

/**
 * Infinite lists: calls `onLoad` once the element given the returned ref comes
 * within `margin` of the visible area of its scroll container. Pass
 * `enabled = hasNextPage && !isFetchingNextPage`; re-enabling after a page
 * lands checks again, so a sentinel that is still in view keeps loading.
 */
export function useLoadOnScroll(
  enabled: boolean,
  onLoad: () => void,
  margin = 600
): (node: Element | null) => void {
  const [node, setNode] = useState<Element | null>(null);
  const load = useEffectEvent(onLoad);

  useEffect(() => {
    if (!enabled || !node || typeof IntersectionObserver === 'undefined') return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) load();
      },
      { root: scrollParent(node), rootMargin: `${margin}px 0px` }
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [enabled, node, margin]);

  return setNode;
}
