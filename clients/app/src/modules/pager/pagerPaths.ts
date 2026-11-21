/**
 * Canonical in-app paths for the pager. Shapes (see `routes.tsx`):
 * `/$slug/pager`, `/pager/t/$token` and the staff board at `/pager`.
 */

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
