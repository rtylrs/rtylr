import { useEffect, useState, type RefObject } from 'react';

/** Finds the nearest scrollable container (`.pager-root` on Rennili pages) or null for window. */
export function findScrollRoot(element: Element | null): Element | null {
  return element?.closest('.pager-root') ?? null;
}

/** Reactive wrapper around `prefers-reduced-motion: reduce`. */
export function useReducedMotion(): boolean {
  const [reduced, setReduced] = useState(() =>
    typeof window !== 'undefined' && typeof window.matchMedia === 'function'
      ? window.matchMedia('(prefers-reduced-motion: reduce)').matches
      : false
  );

  useEffect(() => {
    if (typeof window === 'undefined' || typeof window.matchMedia !== 'function') return undefined;
    const mq = window.matchMedia('(prefers-reduced-motion: reduce)');
    const update = (): void => setReduced(mq.matches);
    update();
    mq.addEventListener('change', update);
    return () => mq.removeEventListener('change', update);
  }, []);

  return reduced;
}

interface UseInViewOptions {
  once?: boolean;
  threshold?: number;
  rootMargin?: string;
}

/** Observes whether an element intersects the viewport (or `.pager-root`). */
export function useInView(
  ref: RefObject<Element | null>,
  { once = true, threshold = 0.15, rootMargin = '0px' }: UseInViewOptions = {}
): boolean {
  const [visible, setVisible] = useState(() => typeof IntersectionObserver === 'undefined');

  useEffect(() => {
    const node = ref.current;
    if (!node) return undefined;
    if (typeof IntersectionObserver === 'undefined') {
      setVisible(true);
      return undefined;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (!entry) return;
        if (entry.isIntersecting) {
          setVisible(true);
          if (once) observer.disconnect();
        } else if (!once) {
          setVisible(false);
        }
      },
      { root: findScrollRoot(node), threshold, rootMargin }
    );

    observer.observe(node);
    return () => observer.disconnect();
  }, [ref, once, threshold, rootMargin]);

  return visible;
}

/** Tracks which section id is currently active as the user scrolls `.pager-root`. */
export function useScrollSpy(sectionIds: readonly string[]): string | null {
  const [activeId, setActiveId] = useState<string | null>(null);
  const key = sectionIds.join(',');

  useEffect(() => {
    if (typeof document === 'undefined') return undefined;
    const ids = key.split(',').filter(Boolean);
    const elements = ids
      .map((id) => document.getElementById(id))
      .filter((el): el is HTMLElement => el !== null);
    if (elements.length === 0) return undefined;

    const scrollRoot = findScrollRoot(elements[0]);
    const target: EventTarget = scrollRoot ?? window;
    let frame = 0;

    const measure = (): void => {
      frame = 0;
      const rootRect = scrollRoot
        ? scrollRoot.getBoundingClientRect()
        : { top: 0, height: window.innerHeight };
      const triggerY = rootRect.top + Math.min(220, rootRect.height * 0.34);

      let current: string | null = null;
      for (const el of elements) {
        const rect = el.getBoundingClientRect();
        if (rect.top <= triggerY && rect.bottom > rootRect.top + 80) {
          current = el.id;
        }
      }
      setActiveId(current);
    };

    const onScroll = (): void => {
      if (!frame) frame = requestAnimationFrame(measure);
    };

    measure();
    target.addEventListener('scroll', onScroll, { passive: true });
    window.addEventListener('resize', onScroll, { passive: true });
    return () => {
      if (frame) cancelAnimationFrame(frame);
      target.removeEventListener('scroll', onScroll);
      window.removeEventListener('resize', onScroll);
    };
  }, [key]);

  return activeId;
}

/** True once the referenced element has scrolled completely above the fold. */
export function useScrolledPast(ref: RefObject<Element | null>): boolean {
  const [past, setPast] = useState(false);

  useEffect(() => {
    const node = ref.current;
    if (!node || typeof IntersectionObserver === 'undefined') return undefined;
    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        if (!entry) return;
        setPast(!entry.isIntersecting && entry.boundingClientRect.top < 0);
      },
      { root: findScrollRoot(node), threshold: 0 }
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [ref]);

  return past;
}
