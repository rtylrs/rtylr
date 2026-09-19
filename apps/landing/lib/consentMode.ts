/** Basic consent mode: no Google library, configuration, or pings before opt-in. */
export const CONSENT_COOKIE = 'rtylr-consent';
export const CONSENT_EVENT = 'rtylr-consent-change';
export const GA_MEASUREMENT_ID = 'G-6XKFL1Z8HE';

type ConsentState = 'granted' | 'denied' | 'undecided';
interface ConsentWindow extends Window {
  rtylrSyncAnalytics?: (state: ConsentState) => void;
}

// Runs before React and also handles returning visitors. Keep all tag loading
// here: an unconditional script in _document would bypass basic consent mode.
export const CONSENT_MODE_BOOTSTRAP = `(function () {
  var loaded = false;
  var id = '${GA_MEASUREMENT_ID}';
  window.dataLayer = window.dataLayer || [];
  window.gtag = function () { window.dataLayer.push(arguments); };
  window['ga-disable-' + id] = true;
  window.gtag('consent', 'default', {
    ad_storage: 'denied', ad_user_data: 'denied',
    ad_personalization: 'denied', analytics_storage: 'denied'
  });
  window.rtylrSyncAnalytics = function (state) {
    var granted = state === 'granted';
    window['ga-disable-' + id] = !granted;
    if (!granted) {
      // Remove identifiers on both the current host and parent domain. A
      // previous grant must not leave analytics cookies behind on withdrawal.
      document.cookie.split(';').forEach(function (part) {
        var name = part.trim().split('=')[0];
        if (!/^_ga(?:_|$)/.test(name)) return;
        var suffix = '=; Path=/; Max-Age=0; SameSite=Lax';
        document.cookie = name + suffix;
        var host = window.location.hostname;
        if (host === 'rtylr.com' || host.endsWith('.rtylr.com')) {
          document.cookie = name + suffix + '; Domain=.rtylr.com';
        }
      });
      if (loaded) window.gtag('consent', 'update', { analytics_storage: 'denied' });
      return;
    }
    window.gtag('consent', 'update', { analytics_storage: 'granted' });
    if (loaded) return;
    loaded = true;
    window.gtag('js', new Date());
    window.gtag('config', id);
    var script = document.createElement('script');
    script.async = true;
    script.src = 'https://www.googletagmanager.com/gtag/js?id=' + id;
    document.head.appendChild(script);
  };
  var state = 'undecided';
  try {
    var match = document.cookie.match(/(?:^|;\\s*)${CONSENT_COOKIE}=([^;]*)/);
    if (match && decodeURIComponent(match[1]) === 'granted') state = 'granted';
  } catch (e) {}
  window.rtylrSyncAnalytics(state);
})();`;

export function syncConsentMode(state: ConsentState): void {
  if (typeof window !== 'undefined') {
    (window as ConsentWindow).rtylrSyncAnalytics?.(state);
  }
}

export function readConsentCookie(): ConsentState {
  if (typeof document === 'undefined') return 'undecided';
  try {
    const match = document.cookie.match(/(?:^|;\s*)rtylr-consent=([^;]*)/);
    const value = match ? decodeURIComponent(match[1]) : '';
    if (value === 'granted' || value === 'denied') return value;
  } catch {
    // A malformed cookie is not permission to track.
  }
  return 'undecided';
}

export function watchConsentMode(): () => void {
  if (typeof window === 'undefined') return () => undefined;
  const sync = (): void => syncConsentMode(readConsentCookie());
  sync();
  // Read the persisted choice rather than trusting arbitrary event details.
  window.addEventListener(CONSENT_EVENT, sync);
  // Reconcile a choice changed on another tab/subdomain when returning here.
  window.addEventListener('focus', sync);
  return () => {
    window.removeEventListener(CONSENT_EVENT, sync);
    window.removeEventListener('focus', sync);
  };
}
