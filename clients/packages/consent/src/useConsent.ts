import { useEffect, useState } from 'react';

import { clearConsent, readConsent, writeConsent, type ConsentState } from './consent';

/**
 * React hook for consent state. Returns the current state and three
 * action-callbacks. Subscribes to the window `rtylr-consent-change`
 * event so a banner dismissal in one component re-renders another.
 *
 * Typical use: a root-level banner calls `accept()` / `decline()`;
 * the analytics loader in a sibling component reads `state` to
 * decide whether to initialise PostHog. A footer link calls
 * `reopen()` to let the user revisit their choice (GDPR Art. 7(3)).
 */
interface UseConsent {
  state: ConsentState | null;
  accept: () => void;
  decline: () => void;
  reopen: () => void;
}

export function useConsent(): UseConsent {
  // Start as null (hidden) until after mount. Reading the cookie in
  // useState() causes an SSR/hydration mismatch - server always
  // returns 'undecided' while the client may return 'granted', forcing
  // React to remount the subtree and producing a flash. null means
  // "not yet determined": the banner stays invisible on the first
  // paint, then the effect reads the real cookie value and renders
  // only when genuinely undecided.
  const [state, setState] = useState<ConsentState | null>(null);

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
    reopen: (): void => {
      clearConsent();
    },
  };
}
