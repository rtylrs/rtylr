import { useCallback } from 'react';
import { announce } from '../../lib/caret/announce';

/** Stable wrapper around the aria-live announcer. */
export function useAnnounce(): (message: string) => void {
  return useCallback((message: string) => announce(message), []);
}
