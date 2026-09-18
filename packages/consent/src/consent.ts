/**
 * Strict opt-in cookie consent, shared across every rtylr app.
 *
 * State lives in a plain cookie named `rtylr-consent` with value
 * `granted`, `denied`, or absent (meaning "undecided, banner should
 * show"). Keyed on the root domain (`.rtylr.com`) so a user who
 * accepts on landing doesn't see the banner again on flow.
 *
 * clearConsent reopens the banner; listeners must stop optional tracking
 * whenever the state is anything other than granted.
 */

export type ConsentState = 'granted' | 'denied' | 'undecided';

const COOKIE_NAME = 'rtylr-consent';
const COOKIE_MAX_AGE_DAYS = 365;

function consentCookieDomain(): string {
  // The leading-dot `.rtylr.com` domain only works when the current
  // host actually is `rtylr.com` or a subdomain of it. On localhost,
  // on Cloudflare Pages preview URLs (`*.pages.dev`), or on any
  // other host, the browser silently rejects the Set-Cookie header
  // when the Domain attribute doesn't match - which is exactly how
  // "Accept button does nothing" manifested in preview envs.
  //
  // Fix: set the shared domain ONLY for rtylr.com hosts; let the
  // browser default (current host, host-only cookie) handle every
  // other case.
  if (typeof window === 'undefined') return '';
  const host = window.location.hostname;
  if (host === 'rtylr.com' || host.endsWith('.rtylr.com')) {
    return '.rtylr.com';
  }
  return '';
}

export function readConsent(): ConsentState {
  if (typeof document === 'undefined') return 'undecided';
  const match = document.cookie.match(new RegExp(`(?:^|;\\s*)${COOKIE_NAME}=([^;]+)`));
  if (!match) return 'undecided';
  const value = decodeURIComponent(match[1]);
  if (value === 'granted' || value === 'denied') return value;
  return 'undecided';
}

export function writeConsent(state: Exclude<ConsentState, 'undecided'>): void {
  if (typeof document === 'undefined') return;
  const maxAge = COOKIE_MAX_AGE_DAYS * 24 * 60 * 60;
  const domain = consentCookieDomain();
  const parts = [
    `${COOKIE_NAME}=${encodeURIComponent(state)}`,
    'Path=/',
    `Max-Age=${maxAge}`,
    'SameSite=Lax',
  ];
  if (domain) parts.push(`Domain=${domain}`);
  // Secure only when served over https - localhost dev stays plain.
  if (typeof window !== 'undefined' && window.location.protocol === 'https:') {
    parts.push('Secure');
  }
  document.cookie = parts.join('; ');
  // Fire a window event so any interested module (analytics loader,
  // sentry replay init) can react without polling.
  if (typeof window !== 'undefined') {
    window.dispatchEvent(new CustomEvent<ConsentState>('rtylr-consent-change', { detail: state }));
  }
}

/**
 * Clear consent and re-show the banner. Required by GDPR Art. 7(3):
 * withdrawing consent must be as easy as giving it.
 *
 * Writes a same-shape Set-Cookie with Max-Age=0 + matching Domain so
 * the browser actually deletes the cookie set by writeConsent. A
 * plain `document.cookie = 'rtylr-consent=; ...'` without the Domain
 * attr would leave the parent-domain cookie alive on rtylr.com - same
 * shape of bug as the auth-cookie logout regression (commit c81034fc).
 */
export function clearConsent(): void {
  if (typeof document === 'undefined') return;
  const domain = consentCookieDomain();
  const parts = [`${COOKIE_NAME}=`, 'Path=/', 'Max-Age=0', 'SameSite=Lax'];
  if (domain) parts.push(`Domain=${domain}`);
  if (typeof window !== 'undefined' && window.location.protocol === 'https:') {
    parts.push('Secure');
  }
  document.cookie = parts.join('; ');
  // Belt + suspenders: also delete a host-only copy in case an older
  // session wrote one without the Domain attr.
  document.cookie = `${COOKIE_NAME}=; Path=/; Max-Age=0; SameSite=Lax`;
  if (typeof window !== 'undefined') {
    window.dispatchEvent(
      new CustomEvent<ConsentState>('rtylr-consent-change', { detail: 'undecided' })
    );
  }
}
