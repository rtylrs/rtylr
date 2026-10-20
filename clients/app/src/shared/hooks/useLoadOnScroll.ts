import { useEffect, useEffectEvent, useState } from 'react';

/** The nearest ancestor that actually scrolls vertically, or null for the viewport. */
function scrollRoot(node: Element): Element | null {
  for (let el = node.parentElement; el; el = el.parentElement) {
    const { overflowY } = getComputedStyle(el);
    if (/(auto|scroll|overlay)/.test(overflowY) && el.scrollHeight > el.clientHeight) return el;
  }
  return null;
}

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
      { root: scrollRoot(node), rootMargin: `${margin}px 0px` }
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [enabled, node, margin]);

  return setNode;
}
