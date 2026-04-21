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
  const [state, setState] = useState<ConsentState>(() => readConsent());

  useEffect(() => {
    if (typeof window === 'undefined') return;
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
