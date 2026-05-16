import { useMediaQuery } from './useMediaQuery';

/** Wrap useMediaQuery for the reduced-motion preference. */
export function usePreferReducedMotion(): boolean {
  return useMediaQuery('(prefers-reduced-motion: reduce)');
}
