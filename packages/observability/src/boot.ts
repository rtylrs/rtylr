import { initSentry } from './sentry';
import { initPostHog, disablePostHog } from './posthog';

interface BootOptions {
  app: string;
  release?: string;
}

const CONSENT_COOKIE = 'rtylr-consent';
const CONSENT_EVENT = 'rtylr-consent-change';

function hasConsent(): boolean {
  if (typeof document === 'undefined') return false;
  return document.cookie.split(';').some((part) => part.trim() === `${CONSENT_COOKIE}=granted`);
}

/** Keep listening after opt-in so withdrawal stops an already running SDK. */
export function bootObservability(opts: BootOptions): () => void {
  initSentry({ app: opts.app, release: opts.release });
  const sync = (): void => {
    if (hasConsent()) initPostHog({ app: opts.app });
    else disablePostHog();
  };
  sync();
  // The analytics banner does not request consent to session recordings.
  // Do not start Sentry Replay from this generic analytics opt-in.
  if (typeof window === 'undefined') return () => undefined;
  window.addEventListener(CONSENT_EVENT, sync);
  window.addEventListener('focus', sync);
  return () => {
    window.removeEventListener(CONSENT_EVENT, sync);
    window.removeEventListener('focus', sync);
  };
}
