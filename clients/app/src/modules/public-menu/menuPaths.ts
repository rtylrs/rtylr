import { isTenantMenuHost } from '@voxire/config';

/**
 * Canonical in-app paths for the public menu. Every navigation and `<Link>` in
 * this module goes through these helpers so the URL shape lives in one place.
 *
 * On `portal.rtylr.com` / local dev:
 *   `/$slug/menu`, `/$slug/menu/$location`, `/menu/checkout`,
 *   `/menu/order-status/$id`, `/b/$businessId/menu`.
 *
 * On tenant storefront hosts (`https://<slug>.rtylr.com` or custom domains):
 *   `/`, `/$location`, `/checkout`, `/order-status/$id`.
 */

/** Location picker / menu home for a business slug. */
export function menuHomePath(slug: string, hostname?: string): string {
  if (isTenantMenuHost(hostname)) return '/';
  return `/${slug}/menu`;
}

/** A single location's menu. */
export function menuLocationPath(slug: string, location: string, hostname?: string): string {
  if (isTenantMenuHost(hostname)) return `/${location}`;
  return `/${slug}/menu/${location}`;
}

/** Slug-free checkout path. */
export function menuCheckoutPath(hostname?: string): string {
  if (isTenantMenuHost(hostname)) return '/checkout';
  return '/menu/checkout';
}

/** Slug-free order status path. */
export function menuOrderStatusPath(id: string | number, hostname?: string): string {
  if (isTenantMenuHost(hostname)) return `/order-status/${id}`;
  return `/menu/order-status/${id}`;
}

/** POS-backed menu addressed by numeric id. */
export function menuByBusinessIdPath(businessId: string | number): string {
  return `/b/${businessId}/menu`;
}
