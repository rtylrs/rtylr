import { RENNILI_URL, resolveAuthUrl } from '@voxire/config';

/**
 * Paths on the Rennili origin (rennili.rtylr.com), where the whole pager lives.
 * Shapes (see `routes.tsx`): the landing at `/`, the venue page the QR opens
 * at `/$slug`, the ticket at `/t/$token`, the demo at `/try`, and the staff
 * board at `/board` with its `/display`, `/poster`, `/setup` and `/insights`.
 *
 * The Portal paths these replaced (`/$slug/pager`, `/pager/t/$token`,
 * `/pager`...) redirect here from `clients/app/public/_redirects`, because
 * printed posters point at them.
 */

export const PAGER_LANDING_PATH = '/';
export const PAGER_HOW_PATH = '/how-it-works';
export const PAGER_FEATURES_PATH = '/features';
export const PAGER_FAQ_PATH = '/faq';
export const PAGER_RTYLR_PATH = '/rtylr';
export const PAGER_TRY_PATH = '/try';
export const PAGER_BOARD_PATH = '/board';
export const PAGER_POSTER_PATH = '/poster';
export const PAGER_DISPLAY_PATH = '/display';
export const PAGER_SETUP_PATH = '/setup';
export const PAGER_STATS_PATH = '/insights';
export const PAGER_TICKET_PREFIX = '/t/';

/** `/{slug}` - where the venue's one QR points. */
export function pagerJoinPath(slug: string): string {
  return `/${encodeURIComponent(slug)}`;
}

/** The absolute venue link a QR, poster or display carries. */
export function pagerJoinUrl(slug: string): string {
  return `${RENNILI_URL}${pagerJoinPath(slug)}`;
}

/** `/t/{token}` - the ticket a phone is following. */
export function pagerTicketPath(publicToken: string): string {
  return `${PAGER_TICKET_PREFIX}${encodeURIComponent(publicToken)}`;
}

/**
 * The Portal path a Rennili staff page used to live at. The sign-in app sends
 * a signed-in user back to the Portal origin only, so the staff pages hand it
 * this path; the Portal's `_redirects` then forwards it to Rennili.
 */
const PORTAL_STAFF_PATHS: Record<string, string> = {
  [PAGER_BOARD_PATH]: '/pager',
  [PAGER_DISPLAY_PATH]: '/pager/display',
  [PAGER_POSTER_PATH]: '/pager/poster',
  [PAGER_SETUP_PATH]: '/pager/setup',
  [PAGER_STATS_PATH]: '/pager/insights',
};

export function isRenniliStaffPath(pathname: string): boolean {
  const clean = pathname.replace(/\/+$/, '') || '/';
  return Object.prototype.hasOwnProperty.call(PORTAL_STAFF_PATHS, clean);
}

export function portalPathForStaffPage(pathname: string): string {
  const clean = pathname.replace(/\/+$/, '') || '/';
  return PORTAL_STAFF_PATHS[clean] ?? '/pager';
}

/** What a setup link may fill in for the owner, who still confirms it. */
const SETUP_PREFILL_KEYS = ['name', 'slug', 'email'] as const;

/**
 * The sign-in app's sign-up or sign-in page, set to come back to the pager
 * setup once the account is signed in and its email verified. A setup link's
 * venue name, slug and email ride along: the sign-up form starts with the name
 * and email, and the setup with the slug.
 */
export function pagerClaimUrl(mode: 'register' | 'signIn', setupSearch = ''): string {
  const url = new URL(mode === 'register' ? '/register' : '/', resolveAuthUrl());
  const given = new URLSearchParams(setupSearch);
  const prefill = new URLSearchParams();
  for (const key of SETUP_PREFILL_KEYS) {
    const value = given.get(key)?.trim();
    if (value) prefill.set(key, value);
  }
  const query = prefill.toString();
  url.searchParams.set('redirect', 'dash');
  url.searchParams.set(
    'returnTo',
    portalPathForStaffPage(PAGER_SETUP_PATH) + (query ? `?${query}` : '')
  );
  if (mode === 'register') {
    const email = prefill.get('email');
    const name = prefill.get('name');
    if (email) url.searchParams.set('email', email);
    if (name) url.searchParams.set('businessName', name);
  }
  return url.href;
}

/**
 * A setup link sent to a new venue (`/setup?name=&slug=&email=`) opens the
 * sign-up, not the sign-in, for a visitor who is not signed in yet.
 */
export function pagerSignedOutUrl(pathname: string, search: string): string | null {
  const clean = pathname.replace(/\/+$/, '') || '/';
  if (clean !== PAGER_SETUP_PATH) return null;
  const params = new URLSearchParams(search);
  if (!SETUP_PREFILL_KEYS.some((key) => params.get(key)?.trim())) return null;
  return pagerClaimUrl('register', search);
}
