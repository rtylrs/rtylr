import { resolveAuthUrl } from '@voxire/config';

/**
 * Canonical in-app paths for the pager. Shapes (see `routes.tsx`):
 * `/$slug/pager`, `/pager/t/$token`, the demo at `/pager/try`, the staff board
 * at `/pager`, its counter poster at `/pager/poster` and the sign-up's last step
 * at `/pager/setup`.
 */

export const PAGER_BOARD_PATH = '/pager';
export const PAGER_POSTER_PATH = '/pager/poster';
export const PAGER_SETUP_PATH = '/pager/setup';

/** `/{slug}/pager` - where the venue's QR poster points. */
export function pagerJoinPath(slug: string): string {
  return `/${encodeURIComponent(slug)}/pager`;
}

/** `/pager/t/{token}` - the ticket a phone is following. */
export function pagerTicketPath(publicToken: string): string {
  return `/pager/t/${encodeURIComponent(publicToken)}`;
}

const PUBLIC_PAGER_PATHS = [/^\/[^/]+\/pager\/?$/, /^\/pager\/t\/[^/]+\/?$/];

/** True on the customer's join and ticket pages, never on the staff board. */
export function isPagerPublicPath(pathname: string): boolean {
  return PUBLIC_PAGER_PATHS.some((pattern) => pattern.test(pathname));
}

/**
 * The sign-in app's sign-up or sign-in page, set to come back to the pager
 * setup once the account is signed in and its email verified.
 */
export function pagerClaimUrl(mode: 'register' | 'signIn'): string {
  const url = new URL(mode === 'register' ? '/register' : '/', resolveAuthUrl());
  url.searchParams.set('redirect', 'dash');
  url.searchParams.set('returnTo', PAGER_SETUP_PATH);
  return url.href;
}
