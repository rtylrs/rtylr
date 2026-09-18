import { initSentry, enableSentryReplays } from './sentry';
import { initPostHog } from './posthog';

/**
 * bootObservability - one call, done. Every app's main.tsx imports
 * this and hands it an app identifier.
 *
 * What it does:
 *   - Initialises Sentry eagerly (error capture is safe pre-consent -
 *     no session data, no replays, just a stack trace + app tag).
 *   - Checks the cookie banner's decision via @voxire/consent.
 *     - Already granted: init PostHog + enable Sentry replays.
 *     - Undecided or denied: wait. The consent package dispatches
 *       `rtylr-consent-change` on the window when the user accepts;
 *       we listen once and initialise then.
 *
 * This keeps every app's main.tsx one line longer - callers don't
 * need to know about consent internals, PostHog, Sentry replays, or
 * the event name.
 */

interface BootOptions {
  app: string;
  release?: string;
}

// Mirrors what @voxire/consent writes into the cookie. Importing the
// type from the consent package would create a dependency cycle
// (consent already imports from observability), so it's duplicated -
// small cost for a cleaner dep graph.
const CONSENT_COOKIE = 'rtylr-consent';
const CONSENT_EVENT = 'rtylr-consent-change';

type ConsentState = 'granted' | 'denied' | 'undecided';

function readConsentCookie(): ConsentState {
  if (typeof document === 'undefined') return 'undecided';
  const nameEQ = `${CONSENT_COOKIE}=`;
  for (const part of document.cookie.split('; ')) {
    if (part.startsWith(nameEQ)) {
      const value = part.substring(nameEQ.length);
      if (value === 'granted' || value === 'denied') return value;
    }
  }
  return 'undecided';
}

export function bootObservability(opts: BootOptions): void {
  initSentry({ app: opts.app, release: opts.release });

  const enableConsented = (): void => {
    initPostHog({ app: opts.app });
    enableSentryReplays();
  };

  if (readConsentCookie() === 'granted') {
    enableConsented();
    return;
  }

  if (typeof window !== 'undefined') {
    const handler = (e: Event): void => {
      // @voxire/consent dispatches `detail: state` (just the string),
      // but in tests we sometimes use the `detail: {state}` shape.
      // Accept both so a refactor on either side doesn't break this.
      const detail = (e as CustomEvent<unknown>).detail;
      const state: ConsentState | undefined =
        typeof detail === 'string'
          ? (detail as ConsentState)
          : (detail as { state?: ConsentState })?.state;
      if (state === 'granted') {
        window.removeEventListener(CONSENT_EVENT, handler);
        enableConsented();
      }
    };
    window.addEventListener(CONSENT_EVENT, handler);
  }
}
