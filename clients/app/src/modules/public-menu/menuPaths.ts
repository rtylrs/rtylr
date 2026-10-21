/**
 * Canonical in-app paths for the public menu. Every navigation and `<Link>` in
 * this module goes through these helpers so the URL shape lives in one place.
 *
 * Shapes (see `routes.tsx`): `/$slug/menu`, `/$slug/menu/$location`,
 * `/menu/checkout`, `/menu/order-status/$id`, `/b/$businessId/menu`.
 */

/** `/{slug}/menu` - the location picker for a business slug. */
export function menuHomePath(slug: string): string {
  return `/${slug}/menu`;
}

/** `/{slug}/menu/{location}` - a single location's menu. */
export function menuLocationPath(slug: string, location: string): string {
  return `/${slug}/menu/${location}`;
}

/** `/menu/checkout` - slug-free checkout, as today. */
export function menuCheckoutPath(): string {
  return '/menu/checkout';
}

/** `/menu/order-status/{id}` - slug-free order status. */
export function menuOrderStatusPath(id: string | number): string {
  return `/menu/order-status/${id}`;
}

/** `/b/{businessId}/menu` - the POS-backed menu addressed by numeric id. */
export function menuByBusinessIdPath(businessId: string | number): string {
  return `/b/${businessId}/menu`;
}
