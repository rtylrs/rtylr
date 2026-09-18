/**
 * Google Consent Mode v2 for GA4, wired to the site's own consent cookie.
 *
 * The problem this solves: GA4 was installed in _document.tsx and fired
 * on every page load regardless of what the visitor told the cookie
 * banner. The banner worked, the cookie was written, and GA4 never
 * consulted it. Two systems that both worked, not connected.
 *
 * How it works now, in order:
 *
 *   1. `CONSENT_MODE_BOOTSTRAP` runs inline in the document head,
 *      before the GA4 library and before any bundle. It denies every
 *      storage type, then reads the consent cookie and grants
 *      analytics storage immediately if the visitor already accepted.
 *      A returning consented visitor therefore loses nothing.
 *   2. `syncConsentMode` runs from _app.tsx after mount and on every
 *      `rtylr-consent-change` event, so a decision made in this page
 *      view takes effect in this page view.
 *
 * Denied is not the same as off. Under Consent Mode, gtag still sends
 * cookieless pings while analytics storage is denied, so GA4 sees
 * aggregate traffic without identifying the visitor. Granting flips on
 * the client-id cookie and normal attribution.
 *
 * Ad storage, ad user data and ad personalisation stay denied
 * permanently and are never updated. rtylr runs no advertising, and the
 * banner asks one yes-or-no question about analytics: granting an ad
 * consent the visitor was never shown would be inventing a permission.
 */

// Mirrors what @voxire/consent writes and dispatches. Duplicated rather
// than imported for the same reason packages/observability/src/boot.ts
// duplicates it, plus one specific to this file: the bootstrap below is
// stringified into the document head, where it runs before any module
// has loaded and cannot import anything.
//
// tests/unit/consent-mode.test.mjs asserts both values still match the
// consent package's source, so a rename there fails a test instead of
// silently detaching GA4 from consent again.
export const CONSENT_COOKIE = 'rtylr-consent';
export const CONSENT_EVENT = 'rtylr-consent-change';

/** GA4 measurement ID, property "rtylr.com", stream 15484672807. */
export const GA_MEASUREMENT_ID = 'G-6XKFL1Z8HE';

type ConsentState = 'granted' | 'denied' | 'undecided';

interface GtagWindow extends Window {
  dataLayer?: unknown[];
  gtag?: (...args: unknown[]) => void;
}

/**
 * The inline head script. Emitted by _document.tsx before the gtag.js
 * <script> tag, because `consent default` is only honoured if it is
 * queued before the config call.
 *
 * `wait_for_update: 500` covers the case this bootstrap cannot: a
 * first-time visitor who accepts the banner. Their decision arrives
 * from React, which is milliseconds to seconds away, and the parameter
 * tells gtag to hold the initial hits rather than send them denied.
 */
export const CONSENT_MODE_BOOTSTRAP = `window.dataLayer = window.dataLayer || [];
function gtag(){dataLayer.push(arguments);}
gtag('consent', 'default', {
  ad_storage: 'denied',
  ad_user_data: 'denied',
  ad_personalization: 'denied',
  analytics_storage: 'denied',
  wait_for_update: 500
});
try {
  var m = document.cookie.match(/(?:^|;\\s*)${CONSENT_COOKIE}=([^;]*)/);
  if (m && decodeURIComponent(m[1]) === 'granted') {
    gtag('consent', 'update', { analytics_storage: 'granted' });
  }
} catch (e) {}
gtag('js', new Date());
gtag('config', '${GA_MEASUREMENT_ID}');`;

/**
 * Push a consent update for the given state. Safe to call before gtag
 * exists (it will not, only if the GA4 script was blocked) and safe to
 * call repeatedly: gtag treats consent updates as idempotent.
 *
 * Anything other than 'granted' denies. 'undecided' reaches here when
 * the visitor reopens the banner through ManageConsentLink, and that
 * has to revoke, not hold the previous grant.
 */
export function syncConsentMode(state: ConsentState): void {
  if (typeof window === 'undefined') return;
  const w = window as GtagWindow;
  if (typeof w.gtag !== 'function') return;
  w.gtag('consent', 'update', {
    analytics_storage: state === 'granted' ? 'granted' : 'denied',
  });
}

/**
 * Read the consent cookie without importing @voxire/consent, so this
 * module has no React dependency and can be unit tested on its own.
 * @voxire/consent remains the writer; this only reads.
 */
export function readConsentCookie(): ConsentState {
  if (typeof document === 'undefined') return 'undecided';
  const nameEQ = `${CONSENT_COOKIE}=`;
  for (const part of document.cookie.split('; ')) {
    if (part.startsWith(nameEQ)) {
      const value = decodeURIComponent(part.substring(nameEQ.length));
      if (value === 'granted' || value === 'denied') return value;
    }
  }
  return 'undecided';
}

/**
 * Subscribe to consent changes and mirror them into gtag. Returns an
 * unsubscribe function for a React effect cleanup.
 *
 * On mount this only syncs a decision the visitor has actually made.
 * Pushing `update: denied` for an undecided visitor looks harmless and
 * is not: `wait_for_update` in the bootstrap holds the opening hits
 * precisely so that someone who accepts a moment later is measured
 * properly, and any update, including a denial, ends that wait. So an
 * undecided visitor gets silence here and the banner's own event
 * decides. Caught by e2e/consent-mode.spec.ts.
 *
 * Accepts both event detail shapes for the same reason boot.ts does:
 * @voxire/consent dispatches the bare string, and some tests use
 * `{ state }`.
 */
export function watchConsentMode(): () => void {
  if (typeof window === 'undefined') return () => undefined;

  const stored = readConsentCookie();
  if (stored !== 'undecided') syncConsentMode(stored);

  const handler = (e: Event): void => {
    const detail = (e as CustomEvent<unknown>).detail;
    const state =
      typeof detail === 'string'
        ? (detail as ConsentState)
        : (detail as { state?: ConsentState } | null)?.state;
    if (state) syncConsentMode(state);
  };

  window.addEventListener(CONSENT_EVENT, handler);
  return () => window.removeEventListener(CONSENT_EVENT, handler);
}
