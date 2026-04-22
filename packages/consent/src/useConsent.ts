import { useEffect, useState } from 'react';

import { readConsent, writeConsent, type ConsentState } from './consent';

/**
 * React hook for consent state. Returns the current state and two
 * action-callbacks. Subscribes to the window `rtylr-consent-change`
 * event so a banner dismissal in one component re-renders another.
 *
 * Typical use: a root-level banner calls `accept()` / `decline()`;
 * the analytics loader in a sibling component reads `state` to
 * decide whether to initialise PostHog.
 */
interface UseConsent {
  state: ConsentState;
  accept: () => void;
  decline: () => void;
}

export function useConsent(): UseConsent {
  // Start as 'undecided' on every render, including SSR. If we read
  // the cookie during `useState(() => readConsent())`, client-side
  // returns 'granted' (when the visitor already consented) while
  // server-side returns 'undecided' — that SSR/hydration mismatch
  // forces React 19 to remount the banner subtree, and the remount
  // can happen between the user's mousedown and click, eating the
  // event. Instead we read the cookie post-mount in an effect so
  // server + initial client render match, then update state.
  const [state, setState] = useState<ConsentState>('undecided');

  useEffect(() => {
    if (typeof window === 'undefined') return;
    setState(readConsent());
    const handler = (e: Event): void => {
      const detail = (e as CustomEvent<ConsentState>).detail;
      if (detail) setState(detail);
    };
    window.addEventListener('rtylr-consent-change', handler);
    return () => window.removeEventListener('rtylr-consent-change', handler);
  }, []);

  return {
    state,
    accept: (): void => {
      writeConsent('granted');
    },
    decline: (): void => {
      writeConsent('denied');
    },
  };
}
