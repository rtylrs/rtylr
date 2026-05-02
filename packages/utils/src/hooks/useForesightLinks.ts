import { RefObject, useEffect, useRef } from 'react';

type RouteMap = Record<string, () => Promise<unknown>>;

/**
 * Prefetches route component chunks when the user hovers over sidebar links.
 * Improves perceived navigation speed by loading modules before the click.
 *
 * @param containerRef - Ref to the sidebar/nav container element
 * @param routes - Map of route paths to their dynamic import functions
 */
export function useForesightLinks(
  containerRef: RefObject<HTMLElement | null>,
  routes: RouteMap,
): void {
  const prefetchedRef = useRef<Set<string>>(new Set());

  useEffect(() => {
    const container = containerRef.current;
    if (!container) return;

    const handleMouseEnter = (event: MouseEvent) => {
      const target = event.target as HTMLElement;
      const anchor = target.closest('a');
      if (!anchor) return;

      const href = anchor.getAttribute('href');
      if (!href) return;

      // Find a matching route prefix
      const matchedPath = Object.keys(routes).find(
        (path) => href === path || href.startsWith(path + '/') || href.startsWith(path + '?'),
      );

      if (matchedPath && !prefetchedRef.current.has(matchedPath)) {
        prefetchedRef.current.add(matchedPath);
        routes[matchedPath]().catch(() => {
          // Remove from cache on failure so it can retry
          prefetchedRef.current.delete(matchedPath);
        });
      }
    };

    container.addEventListener('mouseover', handleMouseEnter);
    return () => {
      container.removeEventListener('mouseover', handleMouseEnter);
    };
  }, [containerRef, routes]);
}
