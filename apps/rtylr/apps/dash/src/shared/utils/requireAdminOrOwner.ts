import type { StoredUser } from '../stores/userStore';

const MANAGEMENT_PERMISSIONS = ['hub:team:edit', 'hub:roles:view', 'hub:licenses:manage'] as const;

/**
 * Admin-route predicate. Denies by default: passes only when the user is
 * explicitly an owner or holds a management grant. Mirrors the BusinessRoute
 * authorization sources (normalized `effectiveAccess` document when present,
 * legacy `role`/`dashRole` fallback otherwise), but unlike the render-layer
 * guard it treats an omitted `isOwner` as deny rather than allow.
 */
export function isAdminOrOwner(user: StoredUser | null | undefined): boolean {
  if (!user) return false;
  if (user.isOwner === true) return true;
  if (user.effectiveAccess) {
    return user.effectiveAccess.dashPermissions.some((permission) =>
      (MANAGEMENT_PERMISSIONS as readonly string[]).includes(permission)
    );
  }
  const role = (user.role ?? '').toLowerCase();
  const dashRole = (user.dashRole ?? '').toLowerCase();
  return role === 'admin' || role === 'owner' || dashRole === 'admin' || dashRole === 'owner';
}
