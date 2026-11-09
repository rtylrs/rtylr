import { APP_CONFIG, APP_URLS, PORTAL_URL, resolvePortalReturn } from '@voxire/config';

/** The query params that say where someone was going before they signed in. */
const DESTINATION_PARAMS = ['redirect', 'returnTo'] as const;

/**
 * The sign-in or sign-up page, keeping where the visitor was going. Without
 * this, switching to sign-up dropped returnTo and a new account landed on the
 * portal home instead of the page that sent them here.
 */
export function authPageLink(path: '/' | '/register', search: string): string {
  const from = new URLSearchParams(search);
  const to = new URLSearchParams();
  for (const key of DESTINATION_PARAMS) {
    const value = from.get(key);
    if (value) to.set(key, value);
  }
  const query = to.toString();
  return query ? `${path}?${query}` : path;
}

/**
 * Where a newly signed-up visitor goes: back to the portal page they came from
 * (kept on the portal's origin), or the app they named, or the portal home.
 */
export function signedInDestination(redirect: string | null, returnTo: string | null): string {
  if (redirect === 'dash' || redirect === 'portal') return resolvePortalReturn(returnTo);
  if (redirect && Object.prototype.hasOwnProperty.call(APP_CONFIG, redirect)) {
    return APP_URLS[redirect as keyof typeof APP_URLS] as string;
  }
  return PORTAL_URL;
}
