import { useState, useEffect, useRef } from 'react';

interface UseIsVisibleOptions {
  threshold?: number;
  rootMargin?: string;
  once?: boolean;
}

/**
 * Observe whether an element is visible in the viewport using IntersectionObserver.
 * Returns a ref to attach to the target element and a boolean indicating visibility.
 * Pass `once: true` to stop observing after the element becomes visible.
 */
export function useIsVisible<T extends Element>(
  options: UseIsVisibleOptions = {}
): [React.RefObject<T | null>, boolean] {
  const { threshold = 0, rootMargin = '0px', once = false } = options;
  const ref = useRef<T>(null);
  const [isVisible, setIsVisible] = useState(false);

  useEffect(() => {
    const el = ref.current;
    if (!el || typeof IntersectionObserver === 'undefined') return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        const visible = entry.isIntersecting;
        setIsVisible(visible);
        if (visible && once) observer.disconnect();
      },
      { threshold, rootMargin }
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, [threshold, rootMargin, once]);

  return [ref, isVisible];
}
