/**
 * Strict opt-in cookie consent, shared across every rtylr app.
 *
 * State lives in a plain cookie named `rtylr-consent` with value
 * `granted`, `denied`, or absent (meaning "undecided, banner should
 * show"). Keyed on the root domain (`.rtylr.com`) so a user who
 * accepts on landing doesn't see the banner again on flow.
 *
 * Consent is irrevocable here for simplicity — you can change it by
 * clearing cookies. A settings-page revoke flow is a fast-follow.
 */

export type ConsentState = 'granted' | 'denied' | 'undecided';

const COOKIE_NAME = 'rtylr-consent';
const COOKIE_MAX_AGE_DAYS = 365;

function consentCookieDomain(): string {
  // In local dev we can't set a leading-dot domain on localhost; let
  // the browser default apply.
  if (typeof window === 'undefined') return '';
  if (window.location.hostname === 'localhost') return '';
  return '.rtylr.com';
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
  // Secure only when served over https — localhost dev stays plain.
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
