import { useEffect, type RefObject } from 'react';

/**
 * Maps URL paths to their prefetch functions.
 * Each value is a dynamic import or router.prefetch call that loads the route module.
 */
export type ForesightPrefetchMap = Record<string, () => Promise<unknown>>;

let managerReady = false;

/**
 * Registers all <a> elements inside a container with ForesightJS so their
 * target modules are prefetched when the user's cursor trajectory predicts
 * intent - before the click lands.
 *
 * @param containerRef Ref to the element containing the nav links (e.g. sidebar)
 * @param prefetchMap  path → prefetch function pairs (dynamic imports / router.prefetch)
 * @param hitSlop      Extra pixels around each element that count as "near" (default 24)
 */
export function useForesightLinks(
  containerRef: RefObject<HTMLElement | null>,
  prefetchMap: ForesightPrefetchMap,
  hitSlop = 24
): void {
  useEffect(() => {
    if (typeof window === 'undefined') return;

    const container = containerRef.current;
    if (!container) return;

    const links = container.querySelectorAll<HTMLAnchorElement>('a[href]');
    if (!links.length) return;

    import('js.foresight')
      .then(({ ForesightManager }) => {
        if (!managerReady) {
          ForesightManager.initialize({ enableMousePrediction: true });
          managerReady = true;
        }

        ForesightManager.instance.register({
          element: links,
          callback: (elementData) => {
            const href = (elementData.element as HTMLAnchorElement).getAttribute('href');
            if (href && prefetchMap[href]) {
              prefetchMap[href]().catch(() => {});
            }
          },
          hitSlop,
        });
      })
      .catch(() => {});
    // Run once - nav links don't change after mount
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
}
