import { PORTAL_URL } from './apps';

/**
 * Public menu URLs live under the Portal origin: `/<business-slug>/menu`
 * (and `/<business-slug>/menu/<location-slug>` for a location).
 *
 * Portal routes own their first path segment, so a business slug that collides
 * with a Portal route name would be shadowed. The same goes for Rennili's
 * routes, where the venue page is `/<business-slug>`. The Go backend rejects those
 * slugs; this list mirrors that decision on the client so the settings form can
 * refuse them before the request is sent.
 */
export const RESERVED_BUSINESS_SLUGS = [
  'automations',
  'webhooks',
  'menus',
  'menu',
  'create-item',
  'create-banner',
  'orders',
  'qr-codes',
  'design',
  'chat',
  'caret',
  'team',
  'roles',
  'audit',
  'notifications',
  'account',
  'billing',
  'theme',
  'help',
  'extensions',
  'store',
  'integrations',
  'insights',
  'flow',
  'crm',
  'erp',
  'finance',
  'hr',
  'recruit',
  'settings',
  'security',
  'api-keys',
  'pos',
  'b',
  'api',
  'auth',
  'login',
  'signup',
  'assets',
  'static',
  'public',
  'admin',
  'www',
  'app',
  'portal',
  // Rennili owns the first path segment on its own origin too.
  'board',
  'display',
  'poster',
  'try',
  'setup',
  't',
] as const;

const PORTAL_ORIGIN = PORTAL_URL.replace(/\/+$/, '');

/**
 * Platform-owned subdomains of `rtylr.com` that must never be treated as
 * tenant storefront subdomains. Kept aligned with the Go host resolver and
 * Cloudflare Worker bypass list.
 */
export const RESERVED_PLATFORM_SUBDOMAINS = [
  'www',
  'portal',
  'pos',
  'pos-downloads',
  'sso',
  'auth',
  'api',
  'app',
  'admin',
  'dash',
  'flow',
  'insights',
  'crm',
  'erp',
  'finance',
  'hr',
  'menu',
  'docs',
  'status',
  'rennili',
  'recruit',
  'account',
  'support',
  'cdn',
  'mail',
  'static',
  'assets',
] as const;

const RFC1123_LABEL_RE = /^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$/;

function currentHostname(hostname?: string): string {
  const raw =
    hostname !== undefined
      ? hostname
      : typeof window !== 'undefined'
        ? window.location.hostname
        : '';
  return raw.trim().toLowerCase().replace(/\.+$/, '');
}

/**
 * Extracts the tenant business slug from `<slug>.rtylr.com` when the host is
 * a single valid non-reserved subdomain label; returns `null` otherwise.
 */
export function extractTenantSubdomain(hostname?: string): string | null {
  const host = currentHostname(hostname);
  const suffix = '.rtylr.com';
  if (!host.endsWith(suffix)) return null;
  const sub = host.slice(0, -suffix.length);
  if (!sub || sub.includes('.') || !RFC1123_LABEL_RE.test(sub)) return null;
  if ((RESERVED_PLATFORM_SUBDOMAINS as readonly string[]).includes(sub)) return null;
  return sub;
}

/**
 * Returns true when the current hostname belongs to a tenant digital menu
 * storefront — either `<slug>.rtylr.com` or a merchant custom domain — and
 * NOT a RTYLR platform host (`portal.rtylr.com`, `localhost`, `*.pages.dev`, etc.).
 */
export function isTenantMenuHost(hostname?: string): boolean {
  const host = currentHostname(hostname);
  if (!host) return false;
  if (
    host === 'localhost' ||
    host === '127.0.0.1' ||
    host === '::1' ||
    /^\d+\.\d+\.\d+\.\d+$/.test(host) ||
    host.endsWith('.localhost') ||
    host.endsWith('.test') ||
    host.endsWith('.local') ||
    host.endsWith('.invalid') ||
    host.endsWith('.pages.dev') ||
    host === 'rtylr.com' ||
    host === 'voxire.com' ||
    host.endsWith('.voxire.com')
  ) {
    return false;
  }
  if (host.endsWith('.rtylr.com')) {
    return extractTenantSubdomain(host) !== null;
  }
  // Any other non-platform hostname (e.g. menu.myrestaurant.com) is treated as a
  // merchant custom domain storefront.
  return true;
}

export function isReservedBusinessSlug(slug: string): boolean {
  const normalized = typeof slug === 'string' ? slug.trim().toLowerCase() : '';
  if (!normalized) return false;
  return (RESERVED_BUSINESS_SLUGS as readonly string[]).includes(normalized);
}

export function buildPublicMenuUrl(slug: string, location?: string): string {
  const base = `${PORTAL_ORIGIN}/${encodeURIComponent(slug)}/menu`;
  return location ? `${base}/${encodeURIComponent(location)}` : base;
}

/** Branded subdomain menu URL (`https://<slug>.rtylr.com[/<location>]`). */
export function buildTenantSubdomainMenuUrl(slug: string, location?: string): string {
  const cleanSlug = encodeURIComponent(slug.trim().toLowerCase());
  const base = `https://${cleanSlug}.rtylr.com`;
  return location ? `${base}/${encodeURIComponent(location)}` : base;
}

/** POS-style menu for a business that has no slug yet, addressed by id. */
export function buildPublicMenuUrlById(businessId: string | number): string {
  return `${PORTAL_ORIGIN}/b/${encodeURIComponent(String(businessId))}/menu`;
}
