import { useEffect, useState } from 'react';

/**
 * Returns true when the given CSS media query string matches.
 * Subscribes to changes and updates automatically.
 */
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState<boolean>(() => {
    if (typeof window === 'undefined') return false;
    return window.matchMedia(query).matches;
  });

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const mql = window.matchMedia(query);
    setMatches(mql.matches);

    function onChange(e: MediaQueryListEvent) {
      setMatches(e.matches);
    }

    mql.addEventListener('change', onChange);
    return () => mql.removeEventListener('change', onChange);
  }, [query]);

  return matches;
}

/** Convenience hook - returns true on mobile-width screens (max-width: 767px). */
export function useIsMobile(): boolean {
  return useMediaQuery('(max-width: 767px)');
}

/** Convenience hook - returns true when the user prefers reduced motion. */
export function usePrefersReducedMotion(): boolean {
  return useMediaQuery('(prefers-reduced-motion: reduce)');
}
