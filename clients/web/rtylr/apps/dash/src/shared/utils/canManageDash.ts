import type { StoredUser } from '../stores/userStore';

const MANAGEMENT_PERMISSIONS = [
  'hub:team:view',
  'hub:team:edit',
  'hub:roles:view',
  'hub:roles:create',
  'hub:roles:edit',
  'hub:licenses:manage',
  'hub:groups:manage',
] as const;

export function canManageDash(user: StoredUser | null | undefined): boolean {
  if (!user) return false;
  const access = user.effectiveAccess;
  return Boolean(
    access &&
    (access.isOwner ||
      access.isAdmin ||
      access.dashPermissions.some((permission) =>
        (MANAGEMENT_PERMISSIONS as readonly string[]).includes(permission)
      ))
  );
}
