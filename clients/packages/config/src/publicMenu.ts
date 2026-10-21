import { PORTAL_URL } from './apps';

/**
 * Public menu URLs live under the Portal origin: `/<business-slug>/menu`
 * (and `/<business-slug>/menu/<location-slug>` for a location).
 *
 * Portal routes own their first path segment, so a business slug that collides
 * with a Portal route name would be shadowed. The Go backend rejects those
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
] as const;

const PORTAL_ORIGIN = PORTAL_URL.replace(/\/+$/, '');

export function isReservedBusinessSlug(slug: string): boolean {
  const normalized = typeof slug === 'string' ? slug.trim().toLowerCase() : '';
  if (!normalized) return false;
  return (RESERVED_BUSINESS_SLUGS as readonly string[]).includes(normalized);
}

export function buildPublicMenuUrl(slug: string, location?: string): string {
  const base = `${PORTAL_ORIGIN}/${encodeURIComponent(slug)}/menu`;
  return location ? `${base}/${encodeURIComponent(location)}` : base;
}

/** POS-style menu for a business that has no slug yet, addressed by id. */
export function buildPublicMenuUrlById(businessId: string | number): string {
  return `${PORTAL_ORIGIN}/b/${encodeURIComponent(String(businessId))}/menu`;
}
