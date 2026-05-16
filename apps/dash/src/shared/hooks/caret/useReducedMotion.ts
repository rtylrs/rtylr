import { useMediaQuery } from '@voxire/caret';

/** True when the user has set "prefers reduced motion" in their OS. */
export function useReducedMotion(): boolean {
  return useMediaQuery('(prefers-reduced-motion: reduce)');
}
