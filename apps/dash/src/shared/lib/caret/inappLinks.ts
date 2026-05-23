import { APP_URLS } from '@voxire/config';

/**
 * Detect markdown links to in-platform routes and rewrite them as
 * router-friendly relative paths when they point at the dash app, or
 * `target=_blank` external URLs when they point at sibling apps.
 *
 * Caret answers often contain links like "[order](pos://orders/42)" -
 * this helper translates those custom schemes to real URLs.
 */
export interface RewrittenLink {
  href: string;
  external: boolean;
}

export function rewriteAppLink(raw: string): RewrittenLink | null {
  const match = /^([a-z]+):\/\/(.+)$/.exec(raw);
  if (!match) return null;
  const [, scheme, path] = match;
  if (scheme === 'dash') {
    return { href: '/' + path.replace(/^\/+/, ''), external: false };
  }
  if (scheme in APP_URLS) {
    const base = (APP_URLS as Record<string, string>)[scheme];
    return { href: `${base.replace(/\/$/, '')}/${path.replace(/^\/+/, '')}`, external: true };
  }
  return null;
}
