import { useEffect, useState } from 'react';

/** Reactive matchMedia hook. */
export function useMediaQuery(query: string): boolean {
  const [matches, setMatches] = useState<boolean>(() => {
    if (typeof window === 'undefined') return false;
    return window.matchMedia(query).matches;
  });

  useEffect(() => {
    if (typeof window === 'undefined') return;
    const mql = window.matchMedia(query);
    const onChange = (e: MediaQueryListEvent): void => setMatches(e.matches);
    mql.addEventListener('change', onChange);
    setMatches(mql.matches);
    return (): void => mql.removeEventListener('change', onChange);
  }, [query]);

  return matches;
}
